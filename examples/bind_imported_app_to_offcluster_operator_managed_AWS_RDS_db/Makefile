## Cluster Admin

.PHONY: install-service-binding-operator-source
install-service-binding-operator-source:
	@oc apply -f operator-source.service-binding-operator.yaml

.PHONY: install-aws-rds-operator-source
install-aws-rds-operator-source:
	@oc apply -f operator-source.aws-rds-operator.yaml

.PHONY: install-operator-secrets
install-operator-secrets: install-aws-rds-operator-secrets

.PHONY: install-aws-rds-operator-secrets
install-aws-rds-operator-secrets:
	$(eval AWS_ACCESS_KEY_ID := $(shell cat ~/.aws/credentials | grep aws_access_key_id | cut -d " " -f3))
	$(eval AWS_SECRET_ACCESS_KEY := $(shell cat ~/.aws/credentials | grep aws_secret_access_key | cut -d " " -f3))
	@sed -e 's,REPLACE_AWS_ACCESS_KEY_ID,$(shell echo -n "$(AWS_ACCESS_KEY_ID)" | base64),g' secret.aws-rds-operator.yaml | \
	 sed -e 's,REPLACE_AWS_SECRET_ACCESS_KEY,$(shell echo -n "$(AWS_SECRET_ACCESS_KEY)" | base64),g' | oc apply -f -

## Application Developer

.PHONY: create-project
create-project:
	@oc new-project service-binding-demo

.PHONY: deploy-shell-app
deploy-shell-app:
	@cat image-stream.shell-app.yaml build-config.shell-app.yaml deployment-config.shell-app.yaml | oc apply -f -

.PHONY: deploy-mydb
deploy-mydb:
	@cat secret.mydb.yaml rds.mydb.yaml | oc apply -f -

.PHONY: create-service-binding-request-shell-app
create-service-binding-request-shell-app:
	@oc apply -f service-binding-request.shell-app.yaml

.PHONY: set-labels-on-nodejs-app
set-labels-on-nodejs-app:
	@oc patch dc nodejs-app -p '{"metadata": {"labels": {"connects-to": "postgres", "environment": "nodejs"}}}'

.PHONY: create-service-binding-request-nodejs-app
create-service-binding-request-nodejs-app:
	@oc apply -f service-binding-request.nodejs-app.yaml

.PHONY: delete-project
delete-project:
	@-oc delete project service-binding-demo
