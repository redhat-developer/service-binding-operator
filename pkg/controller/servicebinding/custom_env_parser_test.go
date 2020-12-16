package servicebinding

import (
	"errors"
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestCustomEnvParser(t *testing.T) {
	type wantedVar struct {
		name       string
		value      string
		errMessage string
	}

	type args struct {
		in           map[string]interface{}
		wanted       []wantedVar
		varTemplates []v1alpha1.Mapping
	}

	testCase := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			parser := newMappingsParser(args.varTemplates, args.in)
			values, err := parser.Parse()
			require.NoError(t, err)

			for _, w := range args.wanted {
				require.Equal(t, values[w.name], w.value, w.errMessage)
			}
		}
	}

	t.Run("spec and status only", testCase(args{
		in: map[string]interface{}{
			"spec": map[string]interface{}{
				"dbName": "database-name",
			},
			"status": map[string]interface{}{
				"creds": map[string]interface{}{
					"user": "database-user",
					"pass": "database-pass",
				},
			},
		},
		wanted: []wantedVar{
			{name: "JDBC_CONNECTION_STRING", value: "database-name:database-user@database-pass"},
			{name: "ANOTHER_STRING", value: "database-name_database-user"},
		},
		varTemplates: []v1alpha1.Mapping{
			{
				Name:  "JDBC_CONNECTION_STRING",
				Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
			},
			{
				Name:  "ANOTHER_STRING",
				Value: `{{ .spec.dbName }}_{{ .status.creds.user }}`,
			},
		},
	}))
}

func TestCustomEnvPath_Parse(t *testing.T) {
	type args struct {
		envVarCtx map[string]interface{}
		templates []v1alpha1.Mapping
		expected  map[string]interface{}
		wantErr   error
	}

	assertParse := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			customEnvParser := newMappingsParser(args.templates, args.envVarCtx)
			actual, err := customEnvParser.Parse()
			if args.wantErr != nil {
				require.Error(t, args.wantErr)
				require.Equal(t, args.wantErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, args.expected, actual)
			}
		}
	}

	envVarCtx := map[string]interface{}{
		"spec": map[string]interface{}{
			"dbName": "database-name",
		},
		"status": map[string]interface{}{
			"creds": map[string]interface{}{
				"user": "database-user",
				"pass": "database-pass",
			},
		},
	}

	t.Run("JDBC connection string template", assertParse(args{
		envVarCtx: envVarCtx,
		templates: []v1alpha1.Mapping{
			{
				Name:  "JDBC_CONNECTION_STRING",
				Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
			},
		},
		expected: map[string]interface{}{
			"JDBC_CONNECTION_STRING": "database-name:database-user@database-pass",
		},
	}))

	t.Run("incomplete template", assertParse(args{
		envVarCtx: envVarCtx,
		templates: []v1alpha1.Mapping{
			{
				Name:  "INCOMPLETE_TEMPLATE",
				Value: `{{ .spec.dbName `,
			},
		},
		wantErr: errors.New("template: set:1: unclosed action"),
	}))
}

func TestCustomEnvPath_Parse_exampleCase(t *testing.T) {
	cache := map[string]interface{}{
		"status": map[string]interface{}{
			"dbConfigMap": map[string]interface{}{
				"db.user":     "database-user",
				"db.password": "database-pass",
			},
		},
	}

	envMap := []v1alpha1.Mapping{
		{
			Name:  "JDBC_USERNAME",
			Value: `{{ index .status.dbConfigMap "db.user" }}`,
		},
		{
			Name:  "JDBC_PASSWORD",
			Value: `{{ index .status.dbConfigMap "db.password" }}`,
		},
	}

	customEnvPath := newMappingsParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["JDBC_USERNAME"]
	require.Equal(t, "database-user", str, "Connection string is not matching")
	str2 := values["JDBC_PASSWORD"]
	require.Equal(t, "database-pass", str2, "Connection string is not matching")
}

func TestCustomEnvPath_Parse_ToJson(t *testing.T) {
	cache := map[string]interface{}{
		"spec": map[string]interface{}{
			"dbName": "database-name",
		},
		"status": map[string]interface{}{
			"creds": map[string]interface{}{
				"user": "database-user",
				"pass": "database-pass",
			},
		},
	}

	envMap := []v1alpha1.Mapping{
		{
			Name:  "root",
			Value: `{{ json . }}`,
		},
		{
			Name:  "spec",
			Value: `{{ json .spec }}`,
		},
		{
			Name:  "status",
			Value: `{{ json .status }}`,
		},
		{
			Name:  "creds",
			Value: `{{ json .status.creds }}`,
		},
		{
			Name:  "dbName",
			Value: `{{ json .spec.dbName }}`,
		},
		{
			Name:  "notExist",
			Value: `{{ json .notExist }}`,
		},
	}
	customEnvPath := newMappingsParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["root"]
	require.Equal(t, `{"spec":{"dbName":"database-name"},"status":{"creds":{"pass":"database-pass","user":"database-user"}}}`, str, "root path json string is not matching")
	str2 := values["spec"]
	require.Equal(t, `{"dbName":"database-name"}`, str2, "spec json string is not matching")
	str3 := values["status"]
	require.Equal(t, `{"creds":{"pass":"database-pass","user":"database-user"}}`, str3, "status json string is not matching")
	str4 := values["creds"]
	require.Equal(t, `{"pass":"database-pass","user":"database-user"}`, str4, "creds json string is not matching")
	str5 := values["dbName"]
	require.Equal(t, `"database-name"`, str5, "dbName json string is not matching")
	str6 := values["notExist"]
	require.Equal(t, "null", str6, "notExist json string is not matching")
}
