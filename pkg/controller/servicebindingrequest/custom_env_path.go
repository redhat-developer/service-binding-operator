package servicebindingrequest

import (
	"bytes"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"text/template"
)

type CustomEnvPath struct {
	EnvMap []v1alpha1.EnvMap
	Cache map[string]interface{}
}

func NewCustomEnvPath(envMap []v1alpha1.EnvMap, cache map[string]interface{}) *CustomEnvPath {
	return &CustomEnvPath{
		EnvMap: envMap,
		Cache:  cache,
	}
}

func (c *CustomEnvPath) Parse() (map[string][]byte,error) {
	data := make(map[string][]byte)
	for _  , v := range c.EnvMap {
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
		data[v.Name] = buf.Bytes()
	}
	return data, nil
}

