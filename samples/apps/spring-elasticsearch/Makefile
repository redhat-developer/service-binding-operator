APP_REPO ?= quay.io/service-binding/spring-productsearch
APP_IMAGE ?= $(APP_REPO):latest

NAMESPACE ?= product-search

MULTIARCH_PLATFORMS ?= linux/amd64,linux/arm64,linux/s390x,linux/ppc64le

# one of "openshift", "minikube"
KUBERNETES_RUNTIME ?= minikube

.PHONY: namespace
namespace:
ifeq ($(KUBERNETES_RUNTIME), openshift)
	-oc new-project $(NAMESPACE)
else ifeq ($(KUBERNETES_RUNTIME), minikube)
	-kubectl create namespace $(NAMESPACE)
	kubectl config set-context --current --namespace=$(NAMESPACE)
endif

.PHONY: build-app
build-app:
	docker build -f Dockerfile.app -t $(APP_IMAGE) .

.PHONY: push-app
push-app:
	docker push $(APP_IMAGE)

.PHONY: build-app-multiarch
build-app-multiarch:
	docker buildx build --no-cache --platform "$(MULTIARCH_PLATFORMS)" -f Dockerfile.app -t $(APP_IMAGE) .

.PHONY: push-app-multiarch
push-app-multiarch:
	docker buildx build --push --platform "$(MULTIARCH_PLATFORMS)" -f Dockerfile.app -t $(APP_IMAGE) .

.PHONY: deploy-app
deploy-app:
	sed -e 's,quay.io/service-binding/spring-productsearch:latest,'$(APP_IMAGE)',g' productsearch-deployment.yaml | \
	kubectl apply -f - -n $(NAMESPACE) --wait

.PHONY: undeploy-app
undeploy-app:
	-kubectl delete -f productsearch-deployment.yaml -n $(NAMESPACE) --wait

# Elasticsearch ECK
.PHONY: deploy-elasticsearch
deploy-elasticsearch:
	kubectl apply -f elasticsearch-deployment.yaml -n $(NAMESPACE)
	kubectl wait --for=condition=ElasticsearchIsReachable=True elasticsearches.elasticsearch.k8s.elastic.co/productsearch-es --timeout=5m

.PHONY: undeploy-elasticsearch
undeploy-elasticsearch:
	-kubectl delete -f elasticsearch-deployment.yaml -n $(NAMESPACE) --wait

.PHONY: bind-elasticsearch
bind-elasticsearch:
	kubectl apply -f productsearch-elasticsearch-binding.yaml -n $(NAMESPACE)

.PHONY: unbind-elasticsearch
unbind-elasticsearch:
	-kubectl delete -f productsearch-elasticsearch-binding.yaml -n $(NAMESPACE)

# elastic/elasticsearch Helm chart
.PHONY: deploy-elastic-es
deploy-elastic-es:
	-kubectl delete secret elastic-config-credentials elastic-config-secret elastic-config-slack elastic-config-custom-path -n $(NAMESPACE)
	kubectl create secret generic elastic-config-credentials --from-literal=password=passwhat? --from-literal=username=elastic -n $(NAMESPACE)
	kubectl create secret generic elastic-config-slack --from-literal=xpack.notification.slack.account.monitoring.secure_url='https://hooks.slack.com/services/asdasdasd/asdasdas/asdasd' -n $(NAMESPACE)
	echo -n "supersecret" > watcher_encryption_key
	kubectl create secret generic elastic-config-secret --from-file=xpack.watcher.encryption_key=./watcher_encryption_key -n $(NAMESPACE)
	kubectl create secret generic elastic-config-custom-path --from-literal=slack_url='https://hooks.slack.com/services/asdasdasd/asdasdas/asdasd' --from-literal=thing_i_don_tcare_about=test -n $(NAMESPACE)
	helm upgrade --install elastic-es --version 7.10.0 elastic/elasticsearch -f elastic-es.$(KUBERNETES_RUNTIME).values.yaml -n $(NAMESPACE)
	kubectl wait --for=condition=Ready=True Pod/elastic-es-master-0 --timeout=5m -n $(NAMESPACE)

.PHONY: undeploy-elastic-es
undeploy-elastic-es:
	-kubectl delete secret elastic-config-credentials elastic-config-secret elastic-config-slack elastic-config-custom-path -n $(NAMESPACE)
	-helm delete elastic-es

.PHONY: deploy-elastic-es-sed
deploy-elastic-es-sed:
	helm upgrade --install elastic-es-sed openshift-helm-charts/redhat-elasticsearch-sed -f elastic-es-sed.values.yaml -n $(NAMESPACE)

.PHONY: undeploy-elastic-es-sed
undeploy-elastic-es-sed:
	-helm delete elastic-es-sed -n $(NAMESPACE)

.PHONY: bind-elastic-es-sed
bind-elastic-es-sed:
	kubectl apply -f productsearch-elastic-es-sed-binding.yaml -n $(NAMESPACE)

.PHONY: unbind-elastic-es-sed
unbind-elastic-es-sed:
	-kubectl delete -f productsearch-elastic-es-sed-binding.yaml -n $(NAMESPACE)
