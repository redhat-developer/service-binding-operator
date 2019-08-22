package catchall

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/common"
)

type CatchAllReconciler struct {
	Client    client.Client
	DynClient dynamic.Interface
	Scheme    *runtime.Scheme
}

const bindingPending = "pending"

// FIXME: update the SBR object, encoded in the request name, return done if successful to the
// FIXME: make sure we can update SBR object to trigger re-reconciliation;
func (r *CatchAllReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := logf.Log.WithName("catchall").WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)

	ns := request.Namespace
	name := request.Name

	gv := v1alpha1.SchemeGroupVersion
	gvr := gv.WithResource("servicebindingrequests")
	resourceClient := r.DynClient.Resource(gvr).Namespace(ns)

	logger.Info("Reading SBR resource...")
	u, err := resourceClient.Get(name, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "On trying to read SBR resource.")
		return common.RequeueError(err)
	}

	logger.Info("Converting unstructured object into SBR...")
	sbr := &v1alpha1.ServiceBindingRequest{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		logger.Error(err, "On converting unstructured to SBR.")
		return common.RequeueError(err)
	}

	logger.Info("Updating SBR status to 'pending'...")
	sbr.Status.BindingStatus = bindingPending
	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(sbr)
	if err != nil {
		logger.Error(err, "On tranforming SBR back to unstructured.")
		return common.RequeueError(err)
	}

	_, err = resourceClient.UpdateStatus(u, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "On updating the status of SBR resource.")
		return common.RequeueError(err)
	}

	logger.Info("SBR resource updated!")
	return common.Done()
}
