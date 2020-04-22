package util

import (
	"encoding/json"
	"strings"
)

//SbrResponse is a struct based on the output of sbr request sent
type SbrResponse struct {
	APIVersion string
	Kind       string
	Metadata   struct {
		Annotations struct{}
		Name        string
		Namespace   string
	}
	Spec struct {
		ApplicationSelector struct {
			Group       string
			Resource    string
			ResourceRef string
			Version     string
		}
		BackingServiceSelector struct {
			Group       string
			Kind        string
			ResourceRef string
			Version     string
		}
	}
}

//UnmarshalJSONData unmarshall the data in form of json to a struct
func UnmarshalJSONData(jsonData string) SbrResponse {
	res := SbrResponse{}
	if strings.Contains(jsonData, "'") {
		jsonData = strings.Trim(jsonData, "'")
	}
	json.Unmarshal([]byte(jsonData), &res)
	return res
}

//GetSbrResponse returns struct for SBR response
func GetSbrResponse() SbrResponse {
	res := SbrResponse{
		APIVersion: "apps.openshift.io/v1alpha1",
		Kind:       "ServiceBindingRequest",
	}
	res.Spec.ApplicationSelector.Group = "apps"
	res.Spec.ApplicationSelector.Version = "v1"
	res.Spec.BackingServiceSelector.Version = "v1alpha1"
	return res
}
