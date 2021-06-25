package mocks

import (
	"fmt"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"strings"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
)

// resource details employed in mocks
const (
	// Fixme(Akash): This values are tightly coupled with postgresql operator.
	// Need to make it more dynamic.
	CRDName    = "postgresql.baiju.dev"
	CRDVersion = "v1alpha1"
	CRDKind    = "Database"
)

var (
	// DBNameSpecDesc default spec descriptor to inform the database name.
	DBNameSpecDesc = olmv1alpha1.SpecDescriptor{
		DisplayName:  "Database Name",
		Description:  "Database Name",
		Path:         "dbName",
		XDescriptors: []string{"binding:env:attribute"},
	}
	ImageSpecDesc = olmv1alpha1.SpecDescriptor{
		Path:         "image",
		DisplayName:  "Image",
		Description:  "Image Name",
		XDescriptors: nil,
	}
	// DBNameSpecDesc default spec descriptor to inform the database name.
	DBNameSpecIp = olmv1alpha1.SpecDescriptor{
		DisplayName:  "Database IP",
		Description:  "Database IP",
		Path:         "dbConnectionIp",
		XDescriptors: []string{"binding:env:attribute"},
	}
	// DBConfigMapSpecDesc spec descriptor to describe a operator that export username and password
	// via config-map, instead of a usual secret.
	DBConfigMapSpecDesc = olmv1alpha1.SpecDescriptor{
		DisplayName: "DB ConfigMap",
		Description: "Database ConfigMap",
		Path:        "dbConfigMap",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:ConfigMap",
			"binding:env:object:configmap:username",
			"binding:env:object:configmap:password",
		},
	}
	// DBPasswordCredentialsOnEnvStatusDesc status descriptor to describe a database operator that
	// publishes username and password over a secret. Default approach.
	DBPasswordCredentialsOnEnvStatusDesc = olmv1alpha1.StatusDescriptor{
		DisplayName: "DB Password Credentials",
		Description: "Database credentials secret",
		Path:        "dbCredentials",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:Secret",
			"binding:env:object:secret:username",
			"binding:env:object:secret:password",
		},
	}
	// DBPasswordCredentialsOnVolumeMountStatusDesc status descriptor to describe a operator that
	// informs credentials via a volume.
	DBPasswordCredentialsOnVolumeMountStatusDesc = olmv1alpha1.StatusDescriptor{
		DisplayName: "DB Password Credentials",
		Description: "Database credentials secret",
		Path:        "dbCredentials",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:Secret",
			"binding:volumemount:secret:username",
			"binding:volumemount:secret:password",
		},
	}
)

func DatabaseCRDMock(ns string) apiextensionv1beta1.CustomResourceDefinition {
	CRDPlural := "databases"
	FullCRDName := CRDPlural + "." + CRDName
	annotations := map[string]string{
		"service.binding/username": "path={.status.dbCredentials},objectType=Secret,valueKey=username",
		"service.binding/password": "path={.status.dbCredentials},objectType=Secret,valueKey=password",
	}

	crd := apiextensionv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			//Namespace:   ns,
			Name:        FullCRDName,
			Annotations: annotations,
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group:   CRDName,
			Version: CRDVersion,
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural: CRDPlural,
				Kind:   CRDKind,
			},
		},
	}

	return crd
}

func UnstructuredDatabaseCRDMock(ns string) (*unstructured.Unstructured, error) {
	crd := DatabaseCRDMock(ns)
	return converter.ToUnstructured(&crd)
}

type PostgresDatabaseSpec struct {
	Username string `json:"username"`
}

type PostgresDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PostgresDatabaseSpec `json:"spec,omitempty"`
}

func PostgresDatabaseCRMock(ns, name string) PostgresDatabase {
	return PostgresDatabase{
		// usually TypeMeta should not be explicitly defined in mocked objects, however, on using
		// it via *unstructured.Unstructured it could not find this CR without it.
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: PostgresDatabaseSpec{
			Username: "redhatdeveloper",
		},
	}
}

func UnstructuredSecretMock(ns, name string) (*unstructured.Unstructured, error) {
	s := SecretMock(ns, name, nil)
	return converter.ToUnstructured(&s)
}

func UnstructuredSecretMockRV(ns, name string) (*unstructured.Unstructured, error) {
	s := SecretMockRV(ns, name)
	return converter.ToUnstructured(&s)
}

func UnstructuredPostgresDatabaseCRMock(ns, name string) (*unstructured.Unstructured, error) {
	c := PostgresDatabaseCRMock(ns, name)
	return converter.ToUnstructured(&c)
}

//
// Usage of TypeMeta in Mocks
//
// 	Usually TypeMeta should not be explicitly defined in mocked objects, however, on using
//  it via *unstructured.Unstructured it could not find this CR without it.
//

// crdDescriptionMock based for mocked objects.
func crdDescriptionMock(
	specDescriptor []olmv1alpha1.SpecDescriptor,
	statusDescriptors []olmv1alpha1.StatusDescriptor,
) olmv1alpha1.CRDDescription {
	return olmv1alpha1.CRDDescription{
		Name:              strings.ToLower(fmt.Sprintf("%s.%s", CRDKind, CRDName)),
		DisplayName:       CRDKind,
		Description:       "mock-crd-description",
		Kind:              CRDKind,
		Version:           CRDVersion,
		SpecDescriptors:   specDescriptor,
		StatusDescriptors: statusDescriptors,
	}
}

// ClusterServiceVersionListMock returns a list with a single CSV object inside, reusing mock.
func ClusterServiceVersionListMock(ns, name string) *olmv1alpha1.ClusterServiceVersionList {
	return &olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{ClusterServiceVersionMock(ns, name)},
	}
}

// CRDDescriptionMock based on PostgreSQL operator, returning a mock using default third party
// operator setup.
func CRDDescriptionMock() olmv1alpha1.CRDDescription {
	return crdDescriptionMock(
		[]olmv1alpha1.SpecDescriptor{DBNameSpecDesc, ImageSpecDesc},
		[]olmv1alpha1.StatusDescriptor{DBPasswordCredentialsOnEnvStatusDesc},
	)
}

// CRDDescriptionVolumeMountMock based on PostgreSQL operator, returns a mock having credentials
// in a volume.
func CRDDescriptionVolumeMountMock() olmv1alpha1.CRDDescription {
	return crdDescriptionMock(
		[]olmv1alpha1.SpecDescriptor{DBNameSpecDesc},
		[]olmv1alpha1.StatusDescriptor{DBPasswordCredentialsOnVolumeMountStatusDesc},
	)
}

// clusterServiceVersionMock base object to create a CSV.
func clusterServiceVersionMock(
	ns,
	name string,
	crdDescription olmv1alpha1.CRDDescription,
) olmv1alpha1.ClusterServiceVersion {
	strategy := olmv1alpha1.StrategyDetailsDeployment{
		DeploymentSpecs: []olmv1alpha1.StrategyDeploymentSpec{{
			Name: "deployment",
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						name: "service-binding-operator",
					},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Command: []string{"service-binding-operator"},
						}},
					},
				},
			},
		}},
	}

	return olmv1alpha1.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterServiceVersion",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: olmv1alpha1.ClusterServiceVersionSpec{
			DisplayName: "mock-database-csv",
			InstallStrategy: olmv1alpha1.NamedInstallStrategy{
				StrategyName: "deployment",
				StrategySpec: strategy,
			},
			CustomResourceDefinitions: olmv1alpha1.CustomResourceDefinitions{
				Owned: []olmv1alpha1.CRDDescription{crdDescription},
			},
		},
	}
}

// ClusterServiceVersionMock based on PostgreSQL operator having what's expected as defaults.
func ClusterServiceVersionMock(ns, name string) olmv1alpha1.ClusterServiceVersion {
	return clusterServiceVersionMock(ns, name, CRDDescriptionMock())
}

// UnstructuredClusterServiceVersionMock unstructured object based on ClusterServiceVersionMock.
func UnstructuredClusterServiceVersionMock(ns, name string) (*unstructured.Unstructured, error) {
	csv := ClusterServiceVersionMock(ns, name)
	return converter.ToUnstructured(&csv)
}

// ClusterServiceVersionVolumeMountMock based on PostgreSQL operator.
func ClusterServiceVersionVolumeMountMock(ns, name string) olmv1alpha1.ClusterServiceVersion {
	return clusterServiceVersionMock(ns, name, CRDDescriptionVolumeMountMock())
}

// UnstructuredClusterServiceVersionVolumeMountMock returns ClusterServiceVersionVolumeMountMock as
// unstructured object
func UnstructuredClusterServiceVersionVolumeMountMock(
	ns string,
	name string,
) (*unstructured.Unstructured, error) {
	csv := ClusterServiceVersionVolumeMountMock(ns, name)
	return converter.ToUnstructured(&csv)
}

// ClusterServiceVersionListVolumeMountMock returns a list with a single CSV object inside, reusing mock.
func ClusterServiceVersionListVolumeMountMock(ns, name string) *olmv1alpha1.ClusterServiceVersionList {
	return &olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{ClusterServiceVersionVolumeMountMock(ns, name)},
	}
}

func RouteCRMock(ns, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "route.openshift.io/v1",
			"kind":       "Route",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
			"spec": map[string]interface{}{
				"host": "https://openshift.cluster.com/host_url",
			},
		},
	}
}

// UnstructuredDatabaseCRMock returns a unstructured version of DatabaseCRMock.
func UnstructuredDatabaseCRMock(ns, name string) *ustrv1.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", CRDName, CRDVersion),
			"kind":       CRDKind,
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
			"status": map[string]interface{}{
				"dbCredentials": "db-credentials",
			},
		},
	}
}

// SecretMock returns a Secret based on PostgreSQL operator usage.
func SecretMock(ns, name string, data map[string][]byte) *corev1.Secret {
	if data == nil {
		data = map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		}
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: data,
	}
}

// SecretMockRV returns a Secret with a resourceVersion.
func SecretMockRV(ns, name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			Name:            name,
			ResourceVersion: "116076",
		},
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}
}

// ConfigMapMock returns a dummy config-map object.
func ConfigMapMock(ns, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string]string{
			"username": "user",
			"password": "password",
		},
	}
}

// ServiceBindingMock return a binding-request mock of informed name and match labels.
func ServiceBindingMock(
	ns string,
	name string,
	backingServiceNamespace *string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *v1alpha12.ServiceBinding {
	var services []v1alpha12.Service
	if backingServiceResourceRef == "" {
		services = []v1alpha12.Service{}
	} else {
		services = []v1alpha12.Service{
			{
				NamespacedRef: v1alpha12.NamespacedRef{
					Ref: v1alpha12.Ref{
						Group: CRDName, Version: CRDVersion, Kind: CRDKind, Name: backingServiceResourceRef,
					},
					Namespace: backingServiceNamespace,
				},
			},
		}
	}
	sbr := &v1alpha12.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: v1alpha12.GroupVersion.String(),
		},
		Spec: v1alpha12.ServiceBindingSpec{
			Mappings: []v1alpha12.Mapping{},
			Application: v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group: applicationGVR.Group, Version: applicationGVR.Version, Resource: applicationGVR.Resource, Name: applicationResourceRef,
				},
				LabelSelector: &metav1.LabelSelector{MatchLabels: matchLabels},
			},
			DetectBindingResources: false,
			BindAsFiles:            false,
			Services:               services,
		},
	}
	return sbr
}

// UnstructuredServiceBindingMock returns a unstructured version of SBR.
func UnstructuredServiceBindingMock(
	ns string,
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) (*unstructured.Unstructured, error) {
	sbr := ServiceBindingMock(ns, name, nil, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	return converter.ToUnstructuredAsGVK(&sbr, v1alpha12.GroupVersionKind)
}

// UnstructuredDeploymentConfigMock converts the DeploymentMock to unstructured.
func UnstructuredDeploymentConfigMock(ns, name string) *ustrv1.Unstructured {
	return &ustrv1.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps.openshift.io/v1",
			"kind":       "DeploymentConfig",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
		},
	}
}

// UnstructuredDeploymentMock converts the DeploymentMock to unstructured.
func UnstructuredDeploymentMock(
	ns,
	name string,
	matchLabels map[string]string,
) (*unstructured.Unstructured, error) {
	d := DeploymentMock(ns, name, matchLabels)
	return converter.ToUnstructured(&d)
}

// DeploymentMock creates a mocked Deployment object of busybox.
func DeploymentMock(ns, name string, matchLabels map[string]string) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels:    matchLabels,

			// used by tests to write the binding secret
			// to an arbitrary path.
			ClusterName: "clusterNameNotInUse",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      name,
					Labels:    matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "busybox",
						Image:   "busybox:latest",
						Command: []string{"sleep", "3600"},
					}},
				},
			},
		},
	}
}

// UnstructuredKnativeServiceMock converts the KnativeServiceMock to unstructured.
func UnstructuredKnativeServiceMock(
	ns,
	name string,
	matchLabels map[string]string,
) *ustrv1.Unstructured {
	u := &ustrv1.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      name,
			},
		},
	}
	u.SetLabels(matchLabels)
	return u
}

//ThirdLevel ...
type ThirdLevel struct {
	Something string `json:"something"`
}

// NestedImage ...
type NestedImage struct {
	Name       string     `json:"name"`
	ThirdLevel ThirdLevel `json:"third"`
}

// NestedDatabaseSpec ...
type NestedDatabaseSpec struct {
	Image NestedImage `json:"image"`
}

// NestedDatabase ...
type NestedDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NestedDatabaseSpec `json:"spec,omitempty"`
}

// NestedDatabaseCRMock based on PostgreSQL operator, returning a instantiated object.
func NestedDatabaseCRMock(ns, name string) NestedDatabase {
	return NestedDatabase{
		// usually TypeMeta should not be explicitly defined in mocked objects, however, on using
		// it via *unstructured.Unstructured it could not find this CR without it.
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: NestedDatabaseSpec{
			Image: NestedImage{
				Name: "postgres",
				ThirdLevel: ThirdLevel{
					Something: "somevalue",
				},
			},
		},
	}
}

// UnstructuredNestedDatabaseCRMock returns a unstructured object from NestedDatabaseCRMock.
func UnstructuredNestedDatabaseCRMock(ns, name string) (*unstructured.Unstructured, error) {
	db := NestedDatabaseCRMock(ns, name)
	return converter.ToUnstructured(&db)
}

// ConfigMapDatabaseSpec ...
type ConfigMapDatabaseSpec struct {
	DBConfigMap string `json:"dbConfigMap"`
	ImageName   string
	Image       string
}

// ConfigMapDatabase ...
type ConfigMapDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigMapDatabaseSpec `json:"spec,omitempty"`
}

// DatabaseConfigMapMock returns a local ConfigMapDatabase object.
func DatabaseConfigMapMock(ns, name, configMapName string) *ConfigMapDatabase {
	return &ConfigMapDatabase{
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: ConfigMapDatabaseSpec{
			DBConfigMap: configMapName,
			Image:       "docker.io/postgres",
			ImageName:   "postgres",
		},
	}
}

// UnstructuredDatabaseConfigMapMock returns a unstructured version of DatabaseConfigMapMock.
func UnstructuredDatabaseConfigMapMock(ns, name, configMapName string) (*unstructured.Unstructured, error) {
	db := DatabaseConfigMapMock(ns, name, configMapName)
	return converter.ToUnstructured(&db)
}
