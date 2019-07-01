package servicebindingrequest

import (
	"context"
	errs "errors"
	"strings"

	osappsv1 "github.com/openshift/api/apps/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_servicebindingrequest")

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServiceBindingRequest{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("servicebindingrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ServiceBindingRequest
	err = c.Watch(&source.Kind{Type: &v1alpha1.ServiceBindingRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ServiceBindingRequest
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.ServiceBindingRequest{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServiceBindingRequest{}

// ReconcileServiceBindingRequest reconciles a ServiceBindingRequest object
type ReconcileServiceBindingRequest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ServiceBindingRequest object and makes changes based on the state read
// and what is in the ServiceBindingRequest.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceBindingRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ServiceBindingRequest")

	// Fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	crdName := instance.Spec.BackingSelector.ResourceType
	crdVersion := instance.Spec.BackingSelector.ResourceVersion

	clo := &client.ListOptions{
		Namespace: request.Namespace,
	}
	csvl := &olmv1alpha1.ClusterServiceVersionList{}
	err = r.client.List(context.TODO(), clo, csvl)
	if err != nil {
		return reconcile.Result{}, err
	}

	operatorName := ""
outerLoop:
	for _, csv := range csvl.Items {
		for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
			if crdName == crd.Name {
				if crdVersion != "" {
					if crdVersion != crd.Version {
						return reconcile.Result{}, errs.New("Version not matching")
					}
				}
				operatorName = csv.Name
				break outerLoop
			}
		}
	}

	nn := types.NamespacedName{Namespace: request.Namespace,
		Name: operatorName}
	csv := &olmv1alpha1.ClusterServiceVersion{}
	err = r.client.Get(context.TODO(), nn, csv)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{Requeue: true}, err
	}

	evList := []corev1.EnvVar{}

	for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
		for _, spec := range crd.SpecDescriptors {
			pt := spec.Path
			for _, xd := range spec.XDescriptors {
				if strings.HasPrefix(xd, "urn:alm:descriptor:servicebindingrequest:secret:") {
					key := strings.Split(xd, ":")[5]
					sks := &corev1.SecretKeySelector{
						Key: key,
					}
					sks.Name = pt
					evs := &corev1.EnvVarSource{
						SecretKeyRef: sks,
					}
					evn := strings.ToUpper(strings.ReplaceAll(instance.Name, "-", "_")) + "_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
					ev := corev1.EnvVar{
						Name:      evn,
						ValueFrom: evs,
					}
					evList = append(evList, ev)
				}
				if strings.HasPrefix(xd, "urn:alm:descriptor:servicebindingrequest:configmap:") {
					key := strings.Split(xd, ":")[5]
					cmks := &corev1.ConfigMapKeySelector{
						Key: key,
					}
					cmks.Name = pt
					evs := &corev1.EnvVarSource{
						ConfigMapKeyRef: cmks,
					}
					evn := strings.ToUpper(strings.ReplaceAll(instance.Name, "-", "_")) + "_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
					ev := corev1.EnvVar{
						Name:      evn,
						ValueFrom: evs,
					}
					evList = append(evList, ev)
				}

			}

		}
	}

	lo := &client.ListOptions{
		Namespace:     request.Namespace,
		LabelSelector: labels.SelectorFromSet(instance.Spec.ApplicationSelector.MatchLabels),
	}

	switch strings.ToLower(instance.Spec.ApplicationSelector.ResourceKind) {
	case "deploymentconfig":
		dcl := &osappsv1.DeploymentConfigList{}
		err = r.client.List(context.TODO(), lo, dcl)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, d := range dcl.Items {
			for i, c := range d.Spec.Template.Spec.Containers {
				c.Env = evList
				d.Spec.Template.Spec.Containers[i] = c
			}
			err = r.client.Update(context.TODO(), &d)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case "statefulset":
		ssl := &appsv1.StatefulSetList{}
		err = r.client.List(context.TODO(), lo, ssl)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, d := range ssl.Items {
			for i, c := range d.Spec.Template.Spec.Containers {
				c.Env = evList
				d.Spec.Template.Spec.Containers[i] = c
			}
			err = r.client.Update(context.TODO(), &d)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	case "daemonset":
		ssl := &appsv1.DaemonSetList{}
		err = r.client.List(context.TODO(), lo, ssl)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, d := range ssl.Items {
			for i, c := range d.Spec.Template.Spec.Containers {
				c.Env = evList
				d.Spec.Template.Spec.Containers[i] = c
			}
			err = r.client.Update(context.TODO(), &d)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

	default:
		dpl := &appsv1.DeploymentList{}
		err = r.client.List(context.TODO(), lo, dpl)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, d := range dpl.Items {
			for i, c := range d.Spec.Template.Spec.Containers {
				c.Env = evList
				d.Spec.Template.Spec.Containers[i] = c
			}
			err = r.client.Update(context.TODO(), &d)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{Requeue: true}, nil

}
