package binding

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetValueByJSONPath(t *testing.T) {
	json := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "StatefulSet",
		"metadata": {
			"name": "db1",
			"namespace": "prj1"
		},
		"spec": {
			"selector": {
				"matchLabels": {
					"app": "db1"
				}
			},
			"serviceName": "db1-svc",
			"template": {
				"metadata": {
					"labels": {
						"app": "db1"
					}
				},
				"spec": {
					"containers": [
						{
							"env": [
								{
									"name": "POSTGRESQL_USER",
									"value": "user1"
								},
								{
									"name": "POSTGRESQL_PASSWORD",
									"value": "k33p5ecret"
								},
								{
									"name": "POSTGRESQL_DATABASE",
									"value": "mydb"
								}
							],
							"image": "centos/postgresql-96-centos7",
							"name": "db1"
						}
					]
				}
			}
		}
	}`)
	tests := []struct {
		Path    string
		Want    interface{}
		WantErr bool
	}{
		{
			Path: ".spec.serviceName",
			Want: "db1-svc",
		},
		{
			Path: ".spec.template.spec.containers[0].name",
			Want: "db1",
		},
		{
			Path: ".spec.template.spec.containers[0].env[2].value",
			Want: "mydb",
		},
		{
			Path: ".spec.template.spec.containers[?(@.name==\"db1\")].env[?(@.name==\"POSTGRESQL_USER\")].value",
			Want: "user1",
		},
		{
			Path: ".spec.template.metadata.labels",
			Want: map[string]interface{}{
				"app": "db1",
			},
		},
		{
			Path:    ".foo",
			WantErr: true,
		},
	}

	var u unstructured.Unstructured
	err := u.UnmarshalJSON(json)
	if err != nil {
		t.Errorf("Error unmarshaling json input\n")
	}

	for _, test := range tests {
		t.Run(test.Path, func(t *testing.T) {
			result, err := getValuesByJSONPath(u.Object, "{"+test.Path+"}")
			if (err != nil) != test.WantErr {
				t.Errorf("Expecting err %v, got %v\n", test.WantErr, err)
			}
			if !test.WantErr && fmt.Sprintf("%v", result[0].Interface()) != fmt.Sprintf("%v", test.Want) {
				t.Errorf("Expecting %v, got %v\n", test.Want, result[0].Interface())
			}
		})
	}
}
