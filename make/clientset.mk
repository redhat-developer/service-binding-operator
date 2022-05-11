CLIENT_GEN := $(shell go env GOPATH)/bin/client-gen
CLIENT_GO_VERSION := $(shell grep client-go go.mod | sed 's/^.*client-go //')


.PHONY: client-gen
client-gen: $(CLIENT_GEN)

$(CLIENT_GEN):
	go install k8s.io/code-generator/cmd/client-gen@$(CLIENT_GO_VERSION)

.PHONY: clientset
## Generate sources for clientset
clientset: client-gen
	rm -rf clientset
	client-gen \
		--input "binding/v1alpha1,spec/v1alpha3" \
		--clientset-name=servicebinding \
		--fake-clientset=false \
		--go-header-file hack/boilerplate.go.txt \
		--input-base github.com/redhat-developer/service-binding-operator/apis \
		--output-package github.com/redhat-developer/service-binding-operator/clientset \
		--plural-exceptions BindableKinds:BindableKinds
	mv github.com/redhat-developer/service-binding-operator/clientset clientset
	rmdir -p github.com/redhat-developer/service-binding-operator
