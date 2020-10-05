package servicebinding

import (
	"context"
	"encoding/base64"
	"testing"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

// objAssertionFunc implements an assertion of given obj in the context of given test t.
type objAssertionFunc func(t *testing.T, obj *unstructured.Unstructured)

func base64StringEqual(expected string, fields ...string) objAssertionFunc {
	return func(t *testing.T, obj *unstructured.Unstructured) {
		raw, found, err := unstructured.NestedString(obj.Object, fields...)
		require.NoError(t, err)
		require.True(t, found, "path %+v not found in %+v", fields, obj)
		decoded, err := base64.StdEncoding.DecodeString(raw)
		require.NoError(t, err)
		require.Equal(t, expected, string(decoded))
	}
}

// TestServiceBinder_Bind exercises scenarios regarding binding SBR and its related resources.
func TestServiceBinder_Bind(t *testing.T) {
	// wantedAction represents an action issued by the component that is required to exist after it
	// finished the operation
	type wantedAction struct {
		verb          string
		resource      string
		name          string
		objAssertions []objAssertionFunc
	}

	type wantedCondition struct {
		Type    conditionsv1.ConditionType
		Status  corev1.ConditionStatus
		Reason  string
		Message string
	}

	// args are the test arguments
	type args struct {
		// options inform the test how to build the ServiceBinder.
		options *serviceBinderOptions
		// wantBuildErr informs the test an error is wanted at build phase.
		wantBuildErr error
		// wantErr informs the test an error is wanted at ServiceBinder's bind phase.
		wantErr error
		// wantActions informs the test all the actions that should have been issued by
		// ServiceBinder.
		wantActions []wantedAction
		// wantConditions informs the test the conditions that should have been issued
		// by ServiceBinder.
		wantConditions []wantedCondition

		wantResult *reconcile.Result
	}

	// assertBind exercises the bind functionality
	assertBind := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			ctx := context.TODO()
			sb, err := buildServiceBinder(ctx, args.options)
			if args.wantBuildErr != nil {
				require.EqualError(t, err, args.wantBuildErr.Error())
				return
			} else {
				require.NoError(t, err)
			}

			res, err := sb.bind()

			if args.wantErr != nil {
				require.EqualError(t, err, args.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
			if args.wantResult != nil {
				require.Equal(t, &args.wantResult, res)
			}

			// extract actions from the dynamic client, regardless of the bind status; it is expected
			// that failures also issue updates for ServiceBinding objects
			dynClient, ok := sb.dynClient.(*fake.FakeDynamicClient)
			require.True(t, ok)
			actions := dynClient.Actions()
			require.NotNil(t, actions)

			if len(args.wantConditions) > 0 {
				// proceed to find whether conditions match wanted conditions
				for _, c := range args.wantConditions {
					if c.Status == corev1.ConditionTrue {
						requireConditionPresentAndTrue(t, c.Type, sb.sbr.Status.Conditions)
					}
					if c.Status == corev1.ConditionFalse {
						requireConditionPresentAndFalse(t, c.Type, sb.sbr.Status.Conditions)
					}
				}
			}
			// regardless of the result, verify the actions expected by the reconciliation
			// process have been issued if user has specified wanted actions
			if len(args.wantActions) > 0 {
				// proceed to find whether actions match wanted actions
				for _, w := range args.wantActions {
					var match bool
					// search for each wanted action in the slice of actions issued by ServiceBinder
					for _, a := range actions {
						// match will be updated in the switch branches below
						if match {
							break
						}

						if a.Matches(w.verb, w.resource) {
							// there are several action types; here it is required to 'type
							// switch' it and perform the right check.
							switch v := a.(type) {
							case k8stesting.GetAction:
								match = v.GetName() == w.name
							case k8stesting.UpdateAction:
								obj := v.GetObject()
								uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
								require.NoError(t, err)
								u := &unstructured.Unstructured{Object: uObj}
								if w.name == u.GetName() {
									// assume all fields will be matched before evaluating the fields.
									match = true

									// in the case a field is not found or the value isn't the expected, break.
									for _, wantedField := range w.objAssertions {
										wantedField(t, u)
									}
								}
							}
						}

						// short circuit to the end of collected actions if the action has matched.
						if match {
							break
						}
					}
				}
			}
		}
	}

	matchLabels := map[string]string{
		"connects-to": "database",
	}

	reconcilerName := "service-binder"
	f := mocks.NewFake(t, reconcilerName)
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBinding{})
	f.S.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})

	d := f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredConfigMap("db1")
	f.AddMockedUnstructuredConfigMap("db2")

	// create and munge a Database CR since there's no "Status" field in
	// databases.postgresql.baiju.dev, requiring us to add the field directly in the unstructured
	// object
	db1 := f.AddMockedUnstructuredPostgresDatabaseCR("db1")
	{
		runtimeStatus := map[string]interface{}{
			"dbConfigMap":   "db1",
			"dbCredentials": "db1",
			"dbName":        "db1",
		}
		err := unstructured.SetNestedMap(db1.Object, runtimeStatus, "status")
		require.NoError(t, err)
	}
	f.AddMockedUnstructuredSecret("db1")

	db2 := f.AddMockedUnstructuredPostgresDatabaseCR("db2")
	{
		runtimeStatus := map[string]interface{}{
			"dbConfigMap":   "db2",
			"dbCredentials": "db2",
			"dbName":        "db2",
		}
		err := unstructured.SetNestedMap(db2.Object, runtimeStatus, "status")
		require.NoError(t, err)
	}
	f.AddMockedUnstructuredSecret("db2")

	// create the ServiceBinding
	sbrSingleService := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "single-sbr",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: matchLabels,
				},
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    d.GetObjectKind().GroupVersionKind().Group,
					Version:  d.GetObjectKind().GroupVersionKind().Version,
					Resource: "deployments",
				},
				LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
			},
			Services: &[]v1alpha1.Service{
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   db1.GetObjectKind().GroupVersionKind().Group,
						Version: db1.GetObjectKind().GroupVersionKind().Version,
						Kind:    db1.GetObjectKind().GroupVersionKind().Kind,
					},
					LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
				},
			},
		},
		Status: v1alpha1.ServiceBindingStatus{},
	}
	f.AddMockResource(sbrSingleService)

	sbrSingleServiceWithCustomEnvVar := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "single-sbr-with-customenvvar",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: matchLabels,
				},
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    d.GetObjectKind().GroupVersionKind().Group,
					Version:  d.GetObjectKind().GroupVersionKind().Version,
					Resource: "deployments",
				},
				LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
			},
			Services: &[]v1alpha1.Service{
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   db1.GetObjectKind().GroupVersionKind().Group,
						Version: db1.GetObjectKind().GroupVersionKind().Version,
						Kind:    db1.GetObjectKind().GroupVersionKind().Kind,
					},
					LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
				},
			},
			CustomEnvVar: []corev1.EnvVar{
				{
					Name:  "MY_DB_NAME",
					Value: `{{ index . "v1alpha1" "postgresql.baiju.dev" "Database" "db1" "status" "dbName" }}`,
				},
				{
					Name:  "MY_DB_CONNECTIONIP",
					Value: `{{ index . "v1alpha1" "postgresql.baiju.dev" "Database" "db1" "status" "dbConnectionIP" }}`,
				},
			},
		},
		Status: v1alpha1.ServiceBindingStatus{},
	}
	f.AddMockResource(sbrSingleServiceWithCustomEnvVar)

	// create the ServiceBinding
	sbrMultipleServices := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "multiple-sbr",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: matchLabels,
				},
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    d.GetObjectKind().GroupVersionKind().Group,
					Version:  d.GetObjectKind().GroupVersionKind().Version,
					Resource: "deployments",
				},
				LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
			},
			Services: &[]v1alpha1.Service{
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   db1.GetObjectKind().GroupVersionKind().Group,
						Version: db1.GetObjectKind().GroupVersionKind().Version,
						Kind:    db1.GetObjectKind().GroupVersionKind().Kind,
					},
					LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
				},
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   db2.GetObjectKind().GroupVersionKind().Group,
						Version: db2.GetObjectKind().GroupVersionKind().Version,
						Kind:    db2.GetObjectKind().GroupVersionKind().Kind,
					},
					LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
				},
			},
		},
		Status: v1alpha1.ServiceBindingStatus{},
	}
	f.AddMockResource(sbrMultipleServices)

	sbrEmptyAppSelector := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "empty-app-selector",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				LabelSelector: &metav1.LabelSelector{},
			},
			Services: &[]v1alpha1.Service{
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   db1.GetObjectKind().GroupVersionKind().Group,
						Version: db1.GetObjectKind().GroupVersionKind().Version,
						Kind:    db1.GetObjectKind().GroupVersionKind().Kind,
					},
					LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
				},
			},
		},
		Status: v1alpha1.ServiceBindingStatus{},
	}
	f.AddMockResource(sbrEmptyAppSelector)

	sbrEmptyServices := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "operators.coreos.com/v1alpha1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "empty-bss",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: matchLabels,
				},
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    d.GetObjectKind().GroupVersionKind().Group,
					Version:  d.GetObjectKind().GroupVersionKind().Version,
					Resource: "deployments",
				},
				LocalObjectReference: corev1.LocalObjectReference{Name: d.GetName()},
			},
			Services: &[]v1alpha1.Service{},
		},
		Status: v1alpha1.ServiceBindingStatus{},
	}
	f.AddMockResource(sbrEmptyServices)

	logger := log.NewLog("service-binder")

	t.Run("single bind golden path", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: false,
			sbr:                    sbrSingleService,
			binding: &internalBinding{
				envVars:    map[string][]byte{},
				volumeKeys: []string{},
			},
			restMapper: testutils.BuildTestRESTMapper(),
		},
		wantConditions: []wantedCondition{
			{
				Type:   CollectionReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:   InjectionReady,
				Status: corev1.ConditionTrue,
			},
		},
		wantActions: []wantedAction{
			{
				resource: "servicebindings",
				verb:     "update",
				name:     sbrSingleService.GetName(),
			},
			{
				resource: "secrets",
				verb:     "update",
				name:     sbrSingleService.GetName(),
			},
			{
				resource: "databases",
				verb:     "update",
				name:     db1.GetName(),
			},
		},
	}))

	t.Run("single bind golden path and custom env vars", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: false,
			sbr:                    sbrSingleServiceWithCustomEnvVar,
			binding: &internalBinding{
				envVars: map[string][]byte{
					"MY_DB_NAME": []byte("db1"),
				},
				volumeKeys: []string{},
			},
			restMapper: testutils.BuildTestRESTMapper(),
		},
		wantConditions: []wantedCondition{
			{
				Type:   CollectionReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:   InjectionReady,
				Status: corev1.ConditionTrue,
			},
		},
		wantActions: []wantedAction{
			{
				resource: "servicebindings",
				verb:     "update",
				name:     sbrSingleServiceWithCustomEnvVar.GetName(),
			},
			{
				resource: "secrets",
				verb:     "update",
				name:     sbrSingleServiceWithCustomEnvVar.GetName(),
				objAssertions: []objAssertionFunc{
					base64StringEqual("db1", "data", "MY_DB_NAME"),
				},
			},
			{
				resource: "databases",
				verb:     "update",
				name:     db1.GetName(),
			},
		},
	}))

	t.Run("bind with binding resource detection", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: true,
			sbr:                    sbrSingleService,
			binding: &internalBinding{
				envVars:    map[string][]byte{},
				volumeKeys: []string{},
			},
			restMapper: testutils.BuildTestRESTMapper(),
		},
		wantConditions: []wantedCondition{
			{
				Type:   CollectionReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:   InjectionReady,
				Status: corev1.ConditionTrue,
			},
		},
	}))

	t.Run("empty application", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: true,
			sbr:                    sbrEmptyAppSelector,
			binding: &internalBinding{
				envVars:    map[string][]byte{},
				volumeKeys: []string{},
			},
			restMapper: testutils.BuildTestRESTMapper(),
		},
		wantConditions: []wantedCondition{
			{
				Type:   CollectionReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:    InjectionReady,
				Status:  corev1.ConditionFalse,
				Reason:  EmptyApplicationReason,
				Message: errEmptyApplication.Error(),
			},
		},
	}))

	// Missing SBR returns an InvalidOptionsErr
	t.Run("bind missing SBR", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: false,
			sbr:                    nil,
			restMapper:             testutils.BuildTestRESTMapper(),
		},
		wantBuildErr: errInvalidServiceBinderOptions("SBR"),
	}))

	t.Run("multiple services bind golden path", assertBind(args{
		options: &serviceBinderOptions{
			logger:                 logger,
			dynClient:              f.FakeDynClient(),
			detectBindingResources: false,
			sbr:                    sbrMultipleServices,
			binding: &internalBinding{
				envVars:    map[string][]byte{},
				volumeKeys: []string{},
			},
			restMapper: testutils.BuildTestRESTMapper(),
		},
		wantConditions: []wantedCondition{
			{
				Type:   CollectionReady,
				Status: corev1.ConditionTrue,
			},
			{
				Type:   InjectionReady,
				Status: corev1.ConditionTrue,
			},
		},
		wantActions: []wantedAction{
			{
				resource: "servicebindings",
				verb:     "update",
				name:     sbrMultipleServices.GetName(),
			},
			{
				resource: "secrets",
				verb:     "update",
				name:     sbrMultipleServices.GetName(),
			},
			{
				resource: "databases",
				verb:     "update",
				name:     db1.GetName(),
			},
			{
				resource: "databases",
				verb:     "update",
				name:     db2.GetName(),
			},
		},
	}))
}

func TestEnsureDefaults(t *testing.T) {
	t.Run("label selector with non nil", func(t *testing.T) {
		applicationSelector := &v1alpha1.Application{}
		ensureDefaults(applicationSelector)
		require.NotNil(t, applicationSelector.LabelSelector)
	})

	t.Run("empty label selector", func(t *testing.T) {
		applicationSelector := &v1alpha1.Application{}
		require.Nil(t, applicationSelector.LabelSelector)
		ensureDefaults(applicationSelector)
		require.NotNil(t, applicationSelector.LabelSelector)
	})

	t.Run("default pod spec path", func(t *testing.T) {
		applicationSelector := &v1alpha1.Application{}
		ensureDefaults(applicationSelector)
		containersPath := getContainersPath(applicationSelector)
		expectedContainersPath := []string{"spec", "template", "spec", "containers"}
		require.Equal(t, expectedContainersPath, containersPath)
	})

	t.Run("container path with value", func(t *testing.T) {
		applicationSelector := &v1alpha1.Application{}
		applicationSelector.BindingPath = &v1alpha1.BindingPath{
			ContainersPath: "spec.some.path",
		}
		ensureDefaults(applicationSelector)
		containersPath := getContainersPath(applicationSelector)
		expectedContainersPath := []string{"spec", "some", "path"}
		require.Equal(t, expectedContainersPath, containersPath)
	})
	t.Run("container path with secret value", func(t *testing.T) {
		applicationSelector := &v1alpha1.Application{}
		applicationSelector.BindingPath = &v1alpha1.BindingPath{
			SecretPath: "spec.some.path",
		}
		ensureDefaults(applicationSelector)
		containersPath := getContainersPath(applicationSelector)
		expectedContainersPath := []string{""}
		require.Equal(t, expectedContainersPath, containersPath)
		secretPath := getSecretFieldPath(applicationSelector)
		expectedSecretPath := []string{"spec", "some", "path"}
		require.Equal(t, expectedSecretPath, secretPath)
	})

}
