package servicebindingrequest

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
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

	envMap := []v1alpha1.EnvMap{
		v1alpha1.EnvMap{
			Name:  "JDBC_CONNECTION_STRING",
			Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
		},
		v1alpha1.EnvMap{
			Name:  "ANOTHER_STRING",
			Value: `{{ .spec.dbName }}_{{ .status.creds.user }}`,
		},
	}
	customEnvPath := NewCustomEnvPath(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["JDBC_CONNECTION_STRING"]
	assert.Equal(t, "database-name:database-user@database-pass", str, "Connection string is not matching")
	str2 := values["ANOTHER_STRING"]
	assert.Equal(t, "database-name_database-user", str2, "Connection string is not matching")
}
