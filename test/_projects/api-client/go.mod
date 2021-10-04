module api-client

go 1.15

require (
	github.com/redhat-developer/service-binding-operator v0.0.0
	k8s.io/apimachinery v0.19.2
)

replace github.com/redhat-developer/service-binding-operator => ../../..
