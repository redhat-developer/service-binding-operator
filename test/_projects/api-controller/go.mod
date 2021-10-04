module api-controller

go 1.15

require (
	github.com/redhat-developer/service-binding-operator v0.0.0
	k8s.io/apimachinery v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)

replace github.com/redhat-developer/service-binding-operator => ../../..
