package servicebindingrequest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client    client.Client     // kubernetes api client
	dynClient dynamic.Interface // kubernetes dynamic api client
	scheme    *runtime.Scheme   // api scheme
}

const (
	// binding is in progress
	bindingInProgress = "inProgress"
	// binding has succeeded
	bindingSuccess = "success"
	// binding has failed
	bindingFail = "fail"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

// setSecretName update the CR status field to "in progress", and setting secret name.
func (r *Reconciler) setSecretName(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	name string,
) error {
	instance.Status.BindingStatus = bindingInProgress
	instance.Status.Secret = name
	return r.client.Status().Update(ctx, instance)

}

// setStatus update the CR status field.
func (r *Reconciler) setStatus(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	status string,
) error {
	instance.Status.BindingStatus = status
	return r.client.Status().Update(ctx, instance)
}

// setApplicationObjects set the ApplicationObject status field, and also set the overall status as
// success, since it was able to bind applications.
func (r *Reconciler) setApplicationObjects(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	objs []string,
) error {
	instance.Status.BindingStatus = bindingSuccess
	instance.Status.ApplicationObjects = objs
	return r.client.Status().Update(ctx, instance)
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value;
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in "spec" level.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	logger := logf.Log.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)
	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		logger.Error(err, "On retrieving service-binding-request instance.")
		return RequeueOnNotFound(err, 0)
	}

	logger = logger.WithValues("ServiceBindingRequest.Name", instance.Name)
	logger.Info("Found service binding request to inspect")

	if err = r.setStatus(ctx, instance, bindingInProgress); err != nil {
		logger.Error(err, "On updating service-binding-request status.")
		return RequeueError(err)
	}

	//
	// Planing changes
	//

	logger.Info("Creating a plan based on OLM and CRD.")
	planner := NewPlanner(ctx, r.dynClient, instance)
	plan, err := planner.Plan()
	if err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On creating a plan to bind applications.")
		return RequeueOnNotFound(err, requeueAfter)
	}

	//
	// Retrieving data
	//

	logger.Info("Retrieving data to create intermediate secret.")
	retriever := NewRetriever(ctx, r.client, plan, instance.Spec.EnvVarPrefix)
	if err = retriever.Retrieve(); err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On retrieving binding data.")
		return RequeueOnNotFound(err, requeueAfter)
	}

	if err = r.setSecretName(ctx, instance, plan.Name); err != nil {
		logger.Error(err, "On updating service-binding-request status.")
		return RequeueError(err)
	}

	//
	// Updating applications to use intermediary secret
	//

	logger.Info("Binding applications with intermediary secret.")
	binder := NewBinder(ctx, r.client, r.dynClient, instance, retriever.volumeKeys)
	if updatedObjectNames, err := binder.Bind(); err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On binding application.")
		return RequeueOnNotFound(err, requeueAfter)
	} else if err = r.setApplicationObjects(ctx, instance, updatedObjectNames); err != nil {
		logger.Error(err, "On updating application objects status field.")
		return RequeueError(err)
	}

<<<<<<< HEAD
	// FIXME: find a way to DRY this block, and then add statefulsets and other kinds back again;
	switch resourceKind {
	case "deploymentconfig":
		logger.Info("Searching DeploymentConfig objects matching labels")

		deploymentConfigListObj := &osappsv1.DeploymentConfigList{}
		err = resourcepoll.WaitUntilResourcesFound(r.client, &searchByLabelsOpts, deploymentConfigListObj)
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
			if len(instance.Status.ApplicationObjects) >= 1 {
				for _, v := range instance.Status.ApplicationObjects {
					if v == deploymentConfigObj.GetName() {
						break
					}
					r.setApplicationObjectsStatus(instance, deploymentConfigObj.GetName())
					err = r.client.Status().Update(context.TODO(), instance)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			} else {
				r.setApplicationObjectsStatus(instance, deploymentConfigObj.GetName())
				err = r.client.Status().Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
			for i, c := range deploymentConfigObj.Spec.Template.Spec.Containers {
				if len(retriever.data) > 0 {
					logger.Info("Adding EnvFrom to container")
					deploymentConfigObj.Spec.Template.Spec.Containers[i].EnvFrom = r.appendEnvFrom(
						c.EnvFrom, instance.GetName())

					existingEnvVars := deploymentConfigObj.Spec.Template.Spec.Containers[i].Env
					if len(existingEnvVars) == 0 {
						existingEnvVars = []corev1.EnvVar{}
					}

					// 	Whenever this env variable's value is updated
					// a new deployment is triggered.

					existingEnvVars = append(existingEnvVars, corev1.EnvVar{
						Name:  "lastboundtime", // TODO: Move out constant.
						Value: retriever.BindableDataHash(),
					})
					deploymentConfigObj.Spec.Template.Spec.Containers[i].Env = existingEnvVars

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
		err = resourcepoll.WaitUntilResourcesFound(r.client, &searchByLabelsOpts, deploymentListObj)
		if err != nil {
			return RequeueOnNotFound(err)
		}
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
			if len(instance.Status.ApplicationObjects) >= 1 {
				for _, v := range instance.Status.ApplicationObjects {
					if v == deploymentObj.GetName() {
						break
					}
					r.setApplicationObjectsStatus(instance, deploymentObj.GetName())
					err = r.client.Status().Update(context.TODO(), instance)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			} else {
				r.setApplicationObjectsStatus(instance, deploymentObj.GetName())
				err = r.client.Status().Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
				}
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
=======
>>>>>>> 6c4ae002f4dfb3f5f27ef75e5a2ed0d9f5522fbb
	logger.Info("All done!")
	return Done()
}
