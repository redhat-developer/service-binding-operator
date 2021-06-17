package naming

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNamingStrategy(t *testing.T) {
	data := map[string]interface{}{
		"name": "database",
		"kind": "Service",
	}

	dataProvider := []struct {
		namingTemplate      string
		bindingName         string
		expectedBindingName string
		expectedError       error
	}{
		{
			namingTemplate:      "{{ .service.kind | upper }}",
			bindingName:         "db",
			expectedBindingName: "SERVICE",
		},
		{
			namingTemplate:      "{{ .service.kind | upper }}_{{ .name }}",
			bindingName:         "db",
			expectedBindingName: "SERVICE_db",
		},
		{
			namingTemplate:      "{{ .service.kind | upper }}_{{ .name | upper }}",
			bindingName:         "db",
			expectedBindingName: "SERVICE_DB",
		},
		{
			namingTemplate:      "{{ .wrongfield | upper }}",
			bindingName:         "db",
			expectedBindingName: "",
			expectedError:       errors.New("please check the namingStrategy template provided"),
		},
	}

	for _, tt := range dataProvider {
		t.Run(fmt.Sprintf("Process %s with %s gives %s", tt.bindingName, tt.namingTemplate, tt.expectedBindingName), func(t *testing.T) {
			template, err := NewTemplate(tt.namingTemplate, data)
			assert.NoError(t, err)
			bindingName, err := template.GetBindingName(tt.bindingName)
			if err != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			}
			assert.EqualValues(t, tt.expectedBindingName, bindingName)
		})
	}

}
