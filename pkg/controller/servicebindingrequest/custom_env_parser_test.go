package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestCustomEnvPath_Parse(t *testing.T) {
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

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "JDBC_CONNECTION_STRING",
			Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
		},
		corev1.EnvVar{
			Name:  "ANOTHER_STRING",
			Value: `{{ .spec.dbName }}_{{ .status.creds.user }}`,
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
	cache := map[string]interface{}{
		"status": map[string]interface{}{
			"dbConfigMap": map[string]interface{}{
				"db.user":     "database-user",
				"db.password": "database-pass",
			},
		},
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "JDBC_USERNAME",
			Value: `{{ index .status.dbConfigMap "db.user" }}`,
		},
		corev1.EnvVar{
			Name:  "JDBC_PASSWORD",
			Value: `{{ index .status.dbConfigMap "db.password" }}`,
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
