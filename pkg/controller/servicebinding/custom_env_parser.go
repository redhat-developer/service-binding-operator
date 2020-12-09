package servicebinding

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
)

// mappingsParser is responsible to interpolate a given EnvVar containing templates.
type mappingsParser struct {
	EnvMap []v1alpha1.Mapping
	Cache  map[string]interface{}
}

// newCustomEnvParser returns a new mappingsParser.
func newMappingsParser(envMap []v1alpha1.Mapping, cache map[string]interface{}) *mappingsParser {
	return &mappingsParser{
		EnvMap: envMap,
		Cache:  cache,
	}
}

// Parse interpolates and caches the templates in EnvMap.
func (c *mappingsParser) Parse() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for _, v := range c.EnvMap {
		tmpl, err := template.New("set").Funcs(template.FuncMap{"json": marshalToJSON}).Parse(v.Value)
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
func marshalToJSON(m interface{}) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
