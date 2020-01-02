package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

// TestServiceBinder_Bind exercises scenarios regarding binding SBR and its related resources.
func TestServiceBinder_Bind(t *testing.T) {
	// wantedAction represents an action issued by the component that is required to exist after it
	// finished the operation
	type wantedAction struct {
		verb     string
		resource string
		name     string
	}

	// args are the test arguments
	type args struct {
		// options inform the test how to build the ServiceBinder.
		options *ServiceBinderOptions
		// wantBuildErr informs the test an error is wanted at build phase.
		wantBuildErr error
		// wantErr informs the test an error is wanted at ServiceBinder's bind phase.
		wantErr error
		// wantedActions informs the test all the actions that should have been issued by
		// ServiceBinder.
		wantedActions []wantedAction
	}

	// assertBind exercises the bind functionality
	assertBind := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			sb, err := BuildServiceBinder(args.options)
			if args.wantBuildErr != nil {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			res, err := sb.Bind()

			if args.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, args.wantErr, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
			}

			// extract actions from the dynamic client, regardless of the bind status; it is expected
			// that failures also issue updates for ServiceBindingRequest objects
			dynClient, ok := sb.DynClient.(*fake.FakeDynamicClient)
			require.True(t, ok)
			actions := dynClient.Actions()
			require.NotNil(t, actions)

			// regardless of the result, verify the actions expected by the reconciliation
			// process have been issued if user has specified wanted actions
			if len(args.wantedActions) > 0 {
				// proceed to find whether actions match wanted actions
				for _, w := range args.wantedActions {
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
								match = w.name == u.GetName()
							}
						}

						// short circuit to the end of collected actions if the action has matched.
						if match {
							break
						}
					}
					require.True(t, match, "expected action %+v not found", w)
				}
			}
		}
	}

	matchLabels := map[string]string{
		"connects-to": "database",
	}

	f := mocks.NewFake(t, reconcilerName)
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	f.S.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})

	d := f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredConfigMap(reconcilerName)

	// create and munge a Database CR since there's no "Status" field in
	// databases.postgres.baiju.dev, requiring us to add the field directly in the unstructured
	// object
	cr := f.AddMockedUnstructuredPostgresDatabaseCR(reconcilerName)
	runtimeStatus := map[string]interface{}{
		"dbConfigMap": reconcilerName,
	}
	err := unstructured.SetNestedMap(cr.Object, runtimeStatus, "status")
	require.NoError(t, err)

	// create the ServiceBindingRequest
	sbr := &v1alpha1.ServiceBindingRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1alpha1",
			Kind:       "ServiceBindingRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: reconcilerName,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			ApplicationSelector: v1alpha1.ApplicationSelector{
				MatchLabels: matchLabels,
				Group:       d.GetObjectKind().GroupVersionKind().Group,
				Version:     d.GetObjectKind().GroupVersionKind().Version,
				Resource:    "deployments",
				ResourceRef: reconcilerName,
			},
			BackingServiceSelectors: []v1alpha1.BackingServiceSelector{
				{
					GroupVersionKind: cr.GetObjectKind().GroupVersionKind(),
					ResourceRef:      reconcilerName,
				},
			},
		},
		Status: v1alpha1.ServiceBindingRequestStatus{},
	}
	f.AddMockResource(sbr)

	logger := log.NewLog("service-binder")
	t.Run("bind golden path", assertBind(args{
		options: &ServiceBinderOptions{
			Logger:                 logger,
			DynClient:              f.FakeDynClient(),
			DetectBindingResources: false,
			EnvVarPrefix:           "",
			SBR:                    sbr,
			Client:                 f.FakeClient(),
		},
		wantedActions: []wantedAction{
			{
				resource: "servicebindingrequests",
				verb:     "update",
				name:     reconcilerName,
			},
			{
				resource: "secrets",
				verb:     "update",
				name:     reconcilerName,
			},
			{
				resource: "databases",
				verb:     "update",
				name:     reconcilerName,
			},
		},
	}))

	t.Run("bind with binding resource detection", assertBind(args{
		options: &ServiceBinderOptions{
			Logger:                 logger,
			DynClient:              f.FakeDynClient(),
			DetectBindingResources: true,
			EnvVarPrefix:           "",
			SBR:                    sbr,
			Client:                 f.FakeClient(),
		},
	}))

	// Missing SBR returns an InvalidOptionsErr
	t.Run("bind missing SBR", assertBind(args{
		options: &ServiceBinderOptions{
			Logger:                 logger,
			DynClient:              f.FakeDynClient(),
			DetectBindingResources: false,
			EnvVarPrefix:           "",
			SBR:                    nil,
			Client:                 f.FakeClient(),
		},
		wantBuildErr: InvalidOptionsErr,
	}))
}
