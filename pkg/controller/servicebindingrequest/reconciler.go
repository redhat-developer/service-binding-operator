package servicebindingrequest

import (
	"context"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// ReconcileServiceBindingRequest reconciles a ServiceBindingRequest object
type ReconcileServiceBindingRequest struct {
	client client.Client
	scheme *runtime.Scheme
}

// selectClusterServiceVersion based on ServiceBindingRequest and a list of CSV (Cluster Service Version)
// picking the one that matches backing-selector rule.
func (r *ReconcileServiceBindingRequest) selectClusterServiceVersion(
	logger logr.Logger,
	instance *v1alpha1.ServiceBindingRequest,
	csvList *olmv1alpha1.ClusterServiceVersionList,
) *olmv1alpha1.ClusterServiceVersion {
	// based on backing-selector, looking for custom resource definition
	backingSelector := instance.Spec.BackingSelector

	logger.WithValues(
		"BackingSelector.ResourceName", backingSelector.ResourceName,
		"BackingSelector.ResourceVersion", backingSelector.ResourceVersion,
	).Info("Looking for a CSV based on backing-selector")

	for _, csv := range csvList.Items {
		logger.WithValues("ClusterServiceVersion.Name", csv.Name).Info("Inspecting CSV...")
		for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
			logger.WithValues("CRD.Name", crd.Name, "CRD.Version", crd.Version).Info("Inspecting CRD...")
			if backingSelector.ResourceName != crd.Name {
				continue
			}
			if crd.Version != "" && backingSelector.ResourceVersion != crd.Version {
				continue
			}
			return &csv
		}
	}

	return nil
}

func (r *ReconcileServiceBindingRequest) retrieveData() {

}

// intermediarySecret create a secret to be used as a intermediary place beteween operator descriptor
// fields and applications interested to have them.
func (r *ReconcileServiceBindingRequest) intermediarySecret(csv *olmv1alpha1.ClusterServiceVersion) *corev1.Secret {
	/*
		for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
			for _, spec := range crd.SpecDescriptors {

			}
		}
	*/

	return nil
}

// Reconcile reads that state of the cluster for a ServiceBindingRequest object and makes changes
// based on the state read and what is in the ServiceBindingRequest.Spec
//
// Note:
// 	The Controller will requeue the Request to be processed again if the returned error is non-nil or
// 	Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceBindingRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logf.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ServiceBindingRequest")

	// Fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// binding-request is not found, empty result means requeue
			return reconcile.Result{}, nil
		}
		// error on executing the request, requeue informing the error
		return reconcile.Result{}, err
	}

	reqLogger.WithValues("ServiceBindingRequest.Name", instance.Name).
		Info("Found service binding request to inspect")

	// list of cluster service version in the namespace
	csvList := &olmv1alpha1.ClusterServiceVersionList{}
	err = r.client.List(context.TODO(), &client.ListOptions{Namespace: request.Namespace}, csvList)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Empty CSV list, requeueing the request")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Error on retrieving CSV list")
		return reconcile.Result{Requeue: true}, err
	}

	// selecting a CSV that matches backing-selector
	csv := r.selectClusterServiceVersion(reqLogger, instance, csvList)
	if csv == nil {
		// unable to obtain a CSV, requeueing
		reqLogger.Info("Warning: Unable to select a CSV object, requeueing!")
		return reconcile.Result{}, nil
	}
	reqLogger.WithValues("ClusterServiceVersion.Name", csv.Name).
		Info("Found cluster-service-version to inspect")

	/*
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
	*/

	return reconcile.Result{Requeue: true}, nil
}
