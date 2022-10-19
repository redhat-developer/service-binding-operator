module api-client

go 1.18

require (
	github.com/redhat-developer/service-binding-operator v0.0.0
	k8s.io/apimachinery v0.22.1
)

replace github.com/redhat-developer/service-binding-operator => ../../..
