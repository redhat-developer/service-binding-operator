package annotations

import (
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const secretValue = "binding:env:object:secret"
const volumeMountSecretValue = "binding:volumemount:secret"

// IsSecret returns true if the annotation value should trigger the secret handler.
func isSecret(s string) bool {
	return secretValue == s || volumeMountSecretValue == s
}

// decodeBase64String asserts whether val is a string and returns its decoded value.
func base64StringValue(v interface{}) (string, error) {
	stringVal, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("should be a string")
	}
	decodedVal, err := base64.StdEncoding.DecodeString(stringVal)
	if err != nil {
		return "", err
	}
	return string(decodedVal), nil
}

// NewSecretHandler constructs a SecretHandler.
func newSecretHandler(
	client dynamic.Interface,
	bi *bindingInfo,
	resource unstructured.Unstructured,
	restMapper meta.RESTMapper,
) (handler, error) {
	h, err := NewResourceHandler(
		client,
		bi,
		resource,
		schema.GroupVersionResource{
			Version:  "v1",
			Resource: "secrets",
		},
		&dataPath,
		restMapper,
	)
	if err != nil {
		return nil, err
	}

	h.stringValue = base64StringValue
	return h, nil
}
