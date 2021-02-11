package controllers

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
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

	log.SetLogger(zap.New(zap.UseDevMode((true))))

	ns := "testing"
	backingServiceNs := "backing-service-ns"
	crName := "db-testing"
	crId := "db_testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNs, nil)

	cr := mocks.UnstructuredDatabaseCRMock(backingServiceNs, crName)

	fakeDynClient := f.FakeDynClient()

	type testCase struct {
		dataMapping []v1alpha1.Mapping
		namePrefix  string
		expected    map[string][]byte
		name        string
		svcCtxs     serviceContextList
	}

	testCases := []testCase{
		{
			name:       "access with index should return correct value",
			namePrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{service: cr},
			},
			dataMapping: []v1alpha1.Mapping{
				{Name: "SAME_NAMESPACE", Value: toIndexTemplate(cr, "metadata.name")},
			},
			expected: map[string][]byte{
				"SAME_NAMESPACE": []byte(cr.GetName()),
			},
		},
		{
			name:       "direct access with apiVersion and kind should return correct value",
			namePrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{service: cr},
			},
			dataMapping: []v1alpha1.Mapping{
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
			name:       "direct access with declared id should return correct value",
			namePrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{
					service: cr,
					id:      &crId,
				},
			},
			dataMapping: []v1alpha1.Mapping{
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
			name:       "direct access without declared id should return <no value>",
			namePrefix: "SERVICE_BINDING",
			svcCtxs: serviceContextList{
				{
					service: cr,
				},
			},
			dataMapping: []v1alpha1.Mapping{
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
				tc.namePrefix, tc.svcCtxs, tc.dataMapping)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildServiceEnvVars(t *testing.T) {

	type testCase struct {
		ctx      *serviceContext
		expected map[string]string
	}

	cr := mocks.UnstructuredDatabaseCRMock("namespace", "name")

	noneNamingStrategy := "{{ .name }}"
	bindAsFilesNaming := "{{ .name | lower}}"
	customNamingStrategy := "CUSTOM_{{ .service.kind | upper }}_{{ .name | upper }}"

	testCases := []testCase{
		{
			ctx: &serviceContext{
				service:        cr,
				namingTemplate: "{{ .name }}",
				envVars: map[string]interface{}{
					"apiKey": map[string]interface{}{
						"key": "my-secret-key",
					},
				},
			},
			expected: map[string]string{
				"apiKey_key": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				service:        cr,
				namingTemplate: noneNamingStrategy,
				envVars: map[string]interface{}{
					"apiKey": map[string]interface{}{
						"key": "my-secret-key",
					},
				},
			},
			expected: map[string]string{
				"apiKey_key": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				service:        cr,
				namingTemplate: noneNamingStrategy,
				envVars: map[string]interface{}{
					"apiKey": map[string]interface{}{
						"key": "my-secret-key",
					},
				},
			},
			expected: map[string]string{
				"apiKey_key": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				namingTemplate: noneNamingStrategy,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"apiKey": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				service:        cr,
				namingTemplate: noneNamingStrategy,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"apiKey": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				namingTemplate: noneNamingStrategy,
				service:        cr,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"apiKey": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				namingTemplate: bindAsFilesNaming,
				service:        cr,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"apikey": "my-secret-key",
			},
		},
		{
			ctx: &serviceContext{
				namingTemplate: customNamingStrategy,
				service:        cr,
				envVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"CUSTOM_DATABASE_APIKEY": "my-secret-key",
			},
		},
	}

	for _, tc := range testCases {
		actual, err := buildServiceEnvVars(tc.ctx, tc.ctx.namingTemplate)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
