package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestCustomEnvPath_Parse(t *testing.T) {
	cache := map[string]map[string]interface{}{
		"testCrId": map[string]interface{}{
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
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "JDBC_CONNECTION_STRING",
			Value: `{{ .testCrId.spec.dbName }}:{{ .testCrId.status.creds.user }}@{{ .testCrId.status.creds.pass }}`,
		},
		corev1.EnvVar{
			Name:  "ANOTHER_STRING",
			Value: `{{ .testCrId.spec.dbName }}_{{ .testCrId.status.creds.user }}`,
		},
	}
	customEnvPath := NewCustomEnvParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["JDBC_CONNECTION_STRING"]
	require.Equal(t, "database-name:database-user@database-pass", str, "Connection string is not matching")
	str2 := values["ANOTHER_STRING"]
	require.Equal(t, "database-name_database-user", str2, "Connection string is not matching")
}

func TestCustomEnvPath_Parse_exampleCase(t *testing.T) {
	cache := map[string]map[string]interface{}{
		"testCrId": map[string]interface{}{
			"status": map[string]interface{}{
				"dbConfigMap": map[string]interface{}{
					"db.user":     "database-user",
					"db.password": "database-pass",
				},
			},
		},
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "JDBC_USERNAME",
			Value: `{{ index .testCrId.status.dbConfigMap "db.user" }}`,
		},
		corev1.EnvVar{
			Name:  "JDBC_PASSWORD",
			Value: `{{ index .testCrId.status.dbConfigMap "db.password" }}`,
		},
	}

	customEnvPath := NewCustomEnvParser(envMap, cache)
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
	cache := map[string]map[string]interface{}{
		"testCrId": map[string]interface{}{
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
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "root",
			Value: `{{ json . }}`,
		},
		corev1.EnvVar{
			Name:  "spec",
			Value: `{{ json .testCrId.spec }}`,
		},
		corev1.EnvVar{
			Name:  "status",
			Value: `{{ json .testCrId.status }}`,
		},
		corev1.EnvVar{
			Name:  "creds",
			Value: `{{ json .testCrId.status.creds }}`,
		},
		corev1.EnvVar{
			Name:  "dbName",
			Value: `{{ json .testCrId.spec.dbName }}`,
		},
		corev1.EnvVar{
			Name:  "notExist",
			Value: `{{ json .notExist }}`,
		},
	}
	customEnvPath := NewCustomEnvParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["root"]
	require.Equal(t, `{"testCrId":{"spec":{"dbName":"database-name"},"status":{"creds":{"pass":"database-pass","user":"database-user"}}}}`, str, "root path json string is not matching")
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
