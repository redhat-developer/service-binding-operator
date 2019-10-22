package servicebindingrequest

import (
	"bytes"
	"text/template"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// CustomEnvParser is responsible to interpolate a given CustomEnvMap containing templates.
type CustomEnvParser struct {
	EnvMap []v1alpha1.CustomEnvMap
	Cache  map[string]interface{}
}

// NewCustomEnvParser returns a new CustomEnvParser.
func NewCustomEnvParser(envMap []v1alpha1.CustomEnvMap, cache map[string]interface{}) *CustomEnvParser {
	return &CustomEnvParser{
		EnvMap: envMap,
		Cache:  cache,
	}
}

// Parse interpolates and caches the templates in EnvMap.
func (c *CustomEnvParser) Parse() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for _, v := range c.EnvMap {
		tmpl, err := template.New("set").Parse(v.Value)
		if err != nil {
			return data, err
		}

		// evaluating template and storing value in a buffer
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, c.Cache)
		if err != nil {
			return data, err
		}

		// saving buffer in cache
		data[v.Name] = buf.String()
	}
	return data, nil
}
