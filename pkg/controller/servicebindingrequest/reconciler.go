package servicebindingrequest

import (
	"context"
	"github.com/redhat-developer/service-binding-operator/pkg/utils"
	"strings"

	osappsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/planner"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client client.Client   // kubernetes api client
	scheme *runtime.Scheme // api scheme
}

// appendEnvFrom based on secret name and list of EnvFromSource instances, making sure secret is
// part of the list or appended.
func (r *Reconciler) appendEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	for _, env := range envList {
		if env.SecretRef.Name == secret {
			// secret name is already referenced
			return envList
		}
	}

	return append(envList, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret,
			},
		},
	})
}

// appendVolumeMounts append volume mounts pointing to volumes created using secret
func (r *Reconciler) appendVolumeMounts(vmList []corev1.VolumeMount, volumeName, mountPath string) []corev1.VolumeMount {
	for _, vm := range vmList {
		if vm.Name == volumeName {
			// volume name already referenced
			return vmList
		}
	}

	return append(vmList, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	})
}

// appendVolumes append volumes
func (r *Reconciler) appendVolumes(volumeList []corev1.Volume, data map[string][]byte, volumeKeys []string, volumeName, secretName string) []corev1.Volume {
	for _, vm := range volumeList {
		if vm.Name == volumeName {
			// volume name already referenced
			return volumeList
		}
	}

	items := []corev1.KeyToPath{}
	for _, k := range volumeKeys {
		items = append(items, corev1.KeyToPath{
			Key:  k,
			Path: k,
		})
	}

	return append(volumeList, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Items:      items,
			},
		},
	})
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value.
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in PodTeamplate level updating `envFrom` entry
// 	  to load intermediary secret;
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	logger := logf.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling ServiceBindingRequest")

	// Fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		// Update Status
		r.setBindingInProgressStatus(instance)
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return RequeueOnNotFound(err)
	}

	logger.WithValues("ServiceBindingRequest.Name", instance.Name).
		Info("Found service binding request to inspect")

	// Set secret name
	r.setSecretStatus(instance)
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	plnr := planner.NewPlanner(ctx, r.client, request.Namespace, instance)
	plan, err := plnr.Plan()
	if err != nil {
		// Update Status
		r.setBindingInProgressStatus(instance)
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return RequeueOnNotFound(err)
	}

	retriever := NewRetriever(ctx, r.client, plan, instance.Spec.EnvVarPrefix)
	if err = retriever.Retrieve(); err != nil {
		// Update Status
		r.setBindingInProgressStatus(instance)
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return RequeueOnNotFound(err)
	}

	//
	// Updating applications to use intermediary secret
	//

	// TODO: very long block that needs to be extracted;
	logger = logger.WithValues("MatchLabels", instance.Spec.ApplicationSelector.MatchLabels)
	logger.Info("Searching applications to receive intermediary secret bind...")

	resourceKind := strings.ToLower(instance.Spec.ApplicationSelector.ResourceKind)
	searchByLabelsOpts := client.ListOptions{
		Namespace:     request.Namespace,
		LabelSelector: labels.SelectorFromSet(instance.Spec.ApplicationSelector.MatchLabels),
	}

	// FIXME: find a way to DRY this block, and then add statefulsets and other kinds back again;
		switch resourceKind {
	case "deploymentconfig":
		logger.Info("Searching DeploymentConfig objects matching labels")

		deploymentConfigListObj := &osappsv1.DeploymentConfigList{}
		err = utils.WaitUntilResourcesFound(r.client, &searchByLabelsOpts, deploymentConfigListObj)
		if err != nil {
			return RequeueOnNotFound(err)
		}
		err = r.client.List(ctx, &searchByLabelsOpts, deploymentConfigListObj)
		if err != nil {
			// Update Status
			r.setBindingInProgressStatus(instance)
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			return RequeueOnNotFound(err)
		}

		if len(deploymentConfigListObj.Items) == 0 {
			logger.Info("No DeploymentConfig objects found, requeueing request!")
			// Update Status
			r.setBindingInProgressStatus(instance)
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			return Requeue()
		}

		for _, deploymentConfigObj := range deploymentConfigListObj.Items {
			logger.WithValues("DeploymentConfig.Name", deploymentConfigObj.GetName()).
				Info("Inspecting DeploymentConfig object...")

			// Update ApplicationObjects Status
			r.setApplicationObjectsStatus(instance, deploymentConfigObj.GetName())
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

			for i, c := range deploymentConfigObj.Spec.Template.Spec.Containers {
				if len(retriever.data) > 0 {
					logger.Info("Adding EnvFrom to container")
					deploymentConfigObj.Spec.Template.Spec.Containers[i].EnvFrom = r.appendEnvFrom(
						c.EnvFrom, instance.GetName())
				}
				if len(retriever.volumeKeys) > 0 {
					logger.Info("Adding VolumeMounts to container")
					mountPath := "/var/data"
					if instance.Spec.MountPathPrefix != "" {
						mountPath = instance.Spec.MountPathPrefix
					}
					deploymentConfigObj.Spec.Template.Spec.Containers[i].VolumeMounts = r.appendVolumeMounts(
						c.VolumeMounts, instance.GetName(), mountPath)
					logger.Info("Adding Volumes to pod")
					deploymentConfigObj.Spec.Template.Spec.Volumes = r.appendVolumes(
						deploymentConfigObj.Spec.Template.Spec.Volumes, retriever.data, retriever.volumeKeys, instance.GetName(), instance.GetName())
				}
			}
			logger.Info("Updating DeploymentConfig object")
			err = r.client.Update(ctx, &deploymentConfigObj)
			if err != nil {
				logger.Error(err, "Error on updating object!")
				// Update Status
				r.setBindingFailStatus(instance)
				err = r.client.Status().Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, err
			}
		}
	default:
		logger.Info("Searching Deployment objects matching labels")

		deploymentListObj := &extv1beta1.DeploymentList{}
		err = r.client.List(ctx, &searchByLabelsOpts, deploymentListObj)
		if err != nil {
			// Update Status
			r.setBindingInProgressStatus(instance)
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			return RequeueOnNotFound(err)
		}

		if len(deploymentListObj.Items) == 0 {
			logger.Info("No Deployment objects found, requeueing request!")
			// Update Status
			r.setBindingInProgressStatus(instance)
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			return Requeue()
		}

		for _, deploymentObj := range deploymentListObj.Items {
			logger = logger.WithValues("Deployment.Name", deploymentObj.GetName())
			logger.Info("Inspecting Deploymen object...")

			// Update ApplicationObjects Status
			r.setApplicationObjectsStatus(instance, deploymentObj.GetName())
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

			for i, c := range deploymentObj.Spec.Template.Spec.Containers {
				if len(retriever.data) > 0 {
					logger.Info("Adding EnvFrom to container")
					deploymentObj.Spec.Template.Spec.Containers[i].EnvFrom = r.appendEnvFrom(
						c.EnvFrom, instance.GetName())
				}
				if len(retriever.volumeKeys) > 0 {
					logger.Info("Adding VolumeMounts to container")
					mountPath := "/var/data"
					if instance.Spec.MountPathPrefix != "" {
						mountPath = instance.Spec.MountPathPrefix
					}
					deploymentObj.Spec.Template.Spec.Containers[i].VolumeMounts = r.appendVolumeMounts(
						c.VolumeMounts, instance.GetName(), mountPath)
					logger.Info("Adding Volumes to pod")
					deploymentObj.Spec.Template.Spec.Volumes = r.appendVolumes(
						deploymentObj.Spec.Template.Spec.Volumes, retriever.data, retriever.volumeKeys, instance.GetName(), instance.GetName())
				}

			}

			logger.Info("Updating Deployment object")
			err = r.client.Update(ctx, &deploymentObj)
			if err != nil {
				// Update Status
				r.setBindingFailStatus(instance)
				err = r.client.Status().Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
				}
				logger.Error(err, "Error on updating object!")
				return reconcile.Result{}, err
			}
		}
	}

	// Update Status
	r.setBindingSuccessStatus(instance)
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	logger.Info("All done!")
	return Done()
}
