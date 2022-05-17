.PHONY: run
## Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet install
	$(GO) run ./main.go

.PHONY: install
## Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
## Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy-cert-manager
deploy-cert-manager:
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml
	kubectl rollout status -n cert-manager deploy/cert-manager-webhook -w --timeout=120s

.PHONY: deploy
## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize image deploy-cert-manager
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_IMAGE_REF)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
## UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -