package servicebinding

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func toIndexTemplate(obj *unstructured.Unstructured, fieldPath string) string {
	gvk := obj.GetObjectKind().GroupVersionKind()
	name := obj.GetName()
	parts := strings.Split(fieldPath, ".")
	var newParts []string
	for _, part := range parts {
		newParts = append(newParts, fmt.Sprintf(`%q`, part))
	}
	indexArg := strings.Join(newParts, " ")
	return fmt.Sprintf(
		`{{ index . %q %q %q %q %s }}`, gvk.Version, gvk.Group, gvk.Kind, name, indexArg)
}

func TestRetrieverProcessServiceContexts(t *testing.T) {

	logf.SetLogger(logf.ZapLogger(true))

	ns := "testing"
	backingServiceNs := "backing-service-ns"
	crName := "db-testing"
	crId := "db_testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNs, nil)

	cr, err := mocks.UnstructuredDatabaseCRMock(backingServiceNs, crName)
	require.NoError(t, err)

	fakeDynClient := f.FakeDynClient()

	type testCase struct {
		dataMapping  []corev1.EnvVar
		envVarPrefix string
		expected     map[string][]byte
		name         string
		svcCtxs      serviceContextList
	}

	testCases := []testCase{
		{
			name:         "access with index should return correct value",
			envVarPrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{service: cr},
			},
			dataMapping: []corev1.EnvVar{
				{Name: "SAME_NAMESPACE", Value: toIndexTemplate(cr, "metadata.name")},
			},
			expected: map[string][]byte{
				"SAME_NAMESPACE": []byte(cr.GetName()),
			},
		},
		{
			name:         "direct access with apiVersion and kind should return correct value",
			envVarPrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{service: cr},
			},
			dataMapping: []corev1.EnvVar{
				{
					Name:  "DIRECT_ACCESS",
					Value: `{{ .v1alpha1.postgresql_baiju_dev.Database.db_testing.metadata.name }}`,
				},
			},
			expected: map[string][]byte{
				"DIRECT_ACCESS": []byte(cr.GetName()),
			},
		},
		{
			name:         "direct access with declared id should return correct value",
			envVarPrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{
					service: cr,
					id:      &crId,
				},
			},
			dataMapping: []corev1.EnvVar{
				{
					Name:  "ID_ACCESS",
					Value: `{{ .db_testing.metadata.name }}`,
				},
			},
			expected: map[string][]byte{
				"ID_ACCESS": []byte(cr.GetName()),
			},
		},
		{
			name:         "direct access without declared id should return <no value>",
			envVarPrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{
					service: cr,
				},
			},
			dataMapping: []corev1.EnvVar{
				{
					Name:  "ID_ACCESS",
					Value: `{{ .db_testing.metadata.name }}`,
				},
			},
			expected: map[string][]byte{
				"ID_ACCESS": []byte("<no value>"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewRetriever(fakeDynClient).ProcessServiceContexts(
				tc.envVarPrefix, tc.svcCtxs, tc.dataMapping)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildServiceEnvVars(t *testing.T) {

	type testCase struct {
		ctx                *serviceContext
		globalEnvVarPrefix string
		expected           map[string]string
	}

	cr, err := mocks.UnstructuredDatabaseCRMock("namespace", "name")
	require.NoError(t, err)

	serviceEnvVarPrefix := "serviceprefix"
	emptyString := ""

	testCases := []testCase{
		{
			globalEnvVarPrefix: "",
			ctx: &serviceContext{
				envVarPrefix: &emptyString,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "globalprefix",
			ctx: &serviceContext{
				envVarPrefix: &emptyString,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"GLOBALPREFIX_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "globalprefix",
			ctx: &serviceContext{
				envVarPrefix: &serviceEnvVarPrefix,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"GLOBALPREFIX_SERVICEPREFIX_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &serviceContext{
				service:      cr,
				envVarPrefix: nil,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"DATABASE_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &serviceContext{
				envVarPrefix: &serviceEnvVarPrefix,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"SERVICEPREFIX_APIKEY": "my-secret-key",
			},
		},
	}

	for _, tc := range testCases {
		actual, err := buildServiceEnvVars(tc.ctx, tc.globalEnvVarPrefix)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
