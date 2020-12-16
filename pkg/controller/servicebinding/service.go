package servicebinding

import (
	"errors"
	"fmt"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebinding/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebinding/nested"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	// errUnspecifiedBackingServiceNamespace is returned when the namespace of a service is
	// unspecified.
	errUnspecifiedBackingServiceNamespace = errors.New("backing service namespace is unspecified")
	// errEmptyServices is returned when no backing service selectors have been
	// informed in the Service Binding.
	errEmptyServices = errors.New("backing service selectors are empty")
	// errEmptyApplication is returned when no application selectors have been
	// informed in the Service Binding.
	errEmptyApplication = errors.New("application selectors are empty")
	// errApplicationNotFound is returned when no application is found
	errApplicationNotFound = errors.New("application not found")
)

func findService(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	name string,
) (
	*unstructured.Unstructured,
	error,
) {
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	if len(ns) == 0 {
		return nil, errUnspecifiedBackingServiceNamespace
	}

	// delegate the search selector's namespaced resource client
	return client.
		Resource(gvr).
		Namespace(ns).
		Get(name, metav1.GetOptions{})
}

// crdGVR is the plural GVR for Kubernetes CRDs.
var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

func findServiceCRD(client dynamic.Interface, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given GVK
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// crdName is the string'fied GroupResource, e.g. "deployments.apps"
	crdName := gvr.GroupResource().String()
	// delegate the search to the CustomResourceDefinition resource client
	return client.Resource(crdGVR).Get(crdName, metav1.GetOptions{})
}

func loadDescriptor(anns map[string]string, path string, descriptor string, root string, objectType string) {
	if !strings.HasPrefix(descriptor, binding.AnnotationPrefix) {
		return
	}

	keys := strings.Split(descriptor, ":")
	key := binding.AnnotationPrefix
	value := ""

	if len(keys) > 1 {
		key += "/" + keys[1]
	} else {
		key += "/" + path
	}

	p := []string{fmt.Sprintf("path={.%s.%s}", root, path)}
	if len(keys) > 1 {
		p = append(p, keys[2:]...)
	}
	if objectType != "" {
		p = append(p, []string{fmt.Sprintf("objectType=%s", objectType)}...)
	}

	value += strings.Join(p, ",")
	anns[key] = value
}

func getObjectType(descriptors []string) string {
	typeAnno := "urn:alm:descriptor:io.kubernetes:"
	for _, desc := range descriptors {
		if strings.HasPrefix(desc, "urn:alm:descriptor:io.kubernetes:") {
			return strings.TrimPrefix(desc, typeAnno)
		}
	}
	return ""
}

func convertCRDDescriptionToAnnotations(crdDescription *olmv1alpha1.CRDDescription) map[string]string {
	anns := make(map[string]string)
	for _, sd := range crdDescription.StatusDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "status", objectType)
		}
	}

	for _, sd := range crdDescription.SpecDescriptors {
		objectType := getObjectType(sd.XDescriptors)
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "spec", objectType)
		}
	}

	return anns
}

// findCRDDescription attempts to find the CRDDescription resource related CustomResourceDefinition.
func findCRDDescription(
	ns string,
	client dynamic.Interface,
	bssGVK schema.GroupVersionKind,
	crd *unstructured.Unstructured,
) (*olmv1alpha1.CRDDescription, error) {
	return newOLM(client, ns).selectCRDByGVK(bssGVK, crd)
}

type bindableResource struct {
	gvk        schema.GroupVersionKind
	gvr        schema.GroupVersionResource
	inputPath  string
	outputPath string
}

var bindableResources = []bindableResource{
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
		inputPath:  "data",
		outputPath: "",
	},
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"},
		inputPath:  "data",
		outputPath: "",
	},
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
		inputPath:  "spec.clusterIP",
		outputPath: "clusterIP",
	},
	{
		gvk: schema.GroupVersionKind{
			Group:   "route.openshift.io",
			Version: "v1",
			Kind:    "Route",
		},
		gvr:        schema.GroupVersionResource{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		inputPath:  "spec.host",
		outputPath: "host",
	},
}

func getOwnedResources(
	logger *log.Log,
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	name string,
	uid types.UID,
) (
	[]*unstructured.Unstructured,
	error,
) {
	var resources []*unstructured.Unstructured
	for _, br := range bindableResources {
		lst, err := client.Resource(br.gvr).Namespace(ns).List(metav1.ListOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				logger.Debug("Resource not found in Bindable Resources", "Error", err)
				continue
			}
			return resources, err
		}
		for idx, item := range lst.Items {
			owners := item.GetOwnerReferences()
			for _, owner := range owners {
				if owner.UID == uid {
					resources = append(resources, &lst.Items[idx])
				}
			}
		}
	}
	return resources, nil
}

func buildOwnedResourceContext(
	client dynamic.Interface,
	obj *unstructured.Unstructured,
	ownerNamePrefix *string,
	restMapper meta.RESTMapper,
	inputPath string,
	outputPath string,
) (*serviceContext, error) {
	svcCtx, err := buildServiceContext(
		reconcilerLog.WithName("buildServiceContext"),
		client,
		obj.GetNamespace(),
		obj.GetObjectKind().GroupVersionKind(),
		obj.GetName(),
		ownerNamePrefix,
		restMapper,
		nil,
	)
	if err != nil {
		return nil, err
	}
	svcCtx.envVars, _, err = nested.GetValue(obj.Object, inputPath, outputPath)
	return svcCtx, err
}

func buildOwnedResourceContexts(
	client dynamic.Interface,
	objs []*unstructured.Unstructured,
	ownerNamePrefix *string,
	restMapper meta.RESTMapper,
) ([]*serviceContext, error) {
	ctxs := make(serviceContextList, 0)

	for _, obj := range objs {
		for _, br := range bindableResources {
			if br.gvk != obj.GetObjectKind().GroupVersionKind() {
				continue
			}
			svcCtx, err := buildOwnedResourceContext(
				client,
				obj,
				ownerNamePrefix,
				restMapper,
				br.inputPath,
				br.outputPath,
			)
			if err != nil {
				return nil, err
			}
			ctxs = append(ctxs, svcCtx)
		}
	}

	return ctxs, nil
}
