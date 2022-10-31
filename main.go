/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	v1apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1beta1apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/redhat-developer/service-binding-operator/apis/webhooks"

	crdcontrollers "github.com/redhat-developer/service-binding-operator/controllers/crd"

	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"

	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/controllers"
	"github.com/redhat-developer/service-binding-operator/controllers/binding"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	specv1alpha3 "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	specv1beta1 "github.com/redhat-developer/service-binding-operator/apis/spec/v1beta1"
	speccontrollers "github.com/redhat-developer/service-binding-operator/controllers/spec"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(specv1alpha3.AddToScheme(scheme))
	utilruntime.Must(v1apiextensions.AddToScheme(scheme))
	utilruntime.Must(v1beta1apiextensions.AddToScheme(scheme))
	utilruntime.Must(specv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	controllers.RegisterFlags(flag.CommandLine)

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "8fa65150.coreos.com",
		Namespace:              watchNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	serviceAccountName, err := kubernetes.WhoAmI(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "cannot detect service account name")
		os.Exit(1)
	}
	setupLog.Info("Service account", "name", serviceAccountName)

	if err = binding.New(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("ServiceBinding"),
		mgr.GetScheme(),
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceBinding")
		os.Exit(1)
	}
	if err = (&v1alpha1.ServiceBinding{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ServiceBinding")
		os.Exit(1)
	}
	if err = speccontrollers.New(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("SPEC ServiceBinding"),
		mgr.GetScheme(),
		&specv1alpha3.ServiceBinding{},
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SPEC ServiceBinding")
		os.Exit(1)
	}
	if err = speccontrollers.New(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("SPEC ServiceBinding"),
		mgr.GetScheme(),
		&specv1beta1.ServiceBinding{},
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SPEC ServiceBinding")
		os.Exit(1)
	}

	webhooks.SetupWithManager(mgr, serviceAccountName)
	if err = (&specv1beta1.ServiceBinding{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "SPEC ServiceBinding")
		os.Exit(1)
	}
	mappingValidator, err := webhooks.NewMappingValidator(
		mgr.GetConfig(),
		mgr.GetRESTMapper(),
	)
	if err != nil {
		setupLog.Error(err, "unable to initialize webhook", "webhook", "ClusterWorkloadResourceMapping")
		os.Exit(1)
	}
	if err = mappingValidator.SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterWorkloadResourceMapping")
		os.Exit(1)
	}

	bindableKinds := &sync.Map{}
	if err = (&crdcontrollers.CrdReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CRD v1"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, bindableKinds); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CRD v1")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
