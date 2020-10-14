package servicebinding

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/PaesslerAG/jsonpath"

	corev1 "k8s.io/api/core/v1"
)

// customEnvParser is responsible to interpolate a given EnvVar containing templates.
type customEnvParser struct {
	EnvMap []corev1.EnvVar
	Cache  map[string]interface{}
}

// newCustomEnvParser returns a new CustomEnvParser.
func newCustomEnvParser(envMap []corev1.EnvVar, cache map[string]interface{}) *customEnvParser {
	return &customEnvParser{
		EnvMap: envMap,
		Cache:  cache,
	}
}

// Parse interpolates and caches the templates in EnvMap.
func (c *customEnvParser) Parse() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for _, v := range c.EnvMap {
		tmpl, err := template.New("set").Funcs(template.FuncMap{"json": marshalToJSON, "jsonp": jsonPath}).Parse(v.Value)
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

func jsonPath(m interface{}, q string) (string, error) {
	v := interface{}(nil)
	err := json.Unmarshal([]byte(m.(string)), &v)
	if err != nil {
		return "", err
	}

	out, err := jsonpath.Get(q, v)
	if err != nil {
		return "", err
	}
	return out.(string), nil
}

func marshalToJSON(m interface{}) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
