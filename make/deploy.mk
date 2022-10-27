.PHONY: run
## Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet install
	$(GO) run ./main.go

.PHONY: install
## Install CRDs into a cluster
install: manifests kustomize kubectl
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
## Uninstall CRDs from a cluster
uninstall: manifests kustomize kubectl
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

.PHONY: deploy-cert-manager
deploy-cert-manager: kubectl
	$(KUBECTL) apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
	$(KUBECTL) rollout status -n cert-manager deploy/cert-manager-webhook -w --timeout=120s

.PHONY: deploy
## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize kubectl image deploy-cert-manager
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_IMAGE_REF)
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
## UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy: kustomize kubectl
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete -f -
