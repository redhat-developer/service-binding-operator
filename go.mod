module github.com/redhat-developer/service-binding-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0 // indirect
	github.com/google/go-cmp v0.5.2
	github.com/imdario/mergo v0.3.10
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/openshift/custom-resource-status v0.0.0-20190822192428-e62f2f3b79f3
	github.com/operator-framework/api v0.3.8
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.6.4
)
