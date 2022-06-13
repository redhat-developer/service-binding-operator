# The following image is a mirror of CI image, performed as:
# * Copy the oc login command from https://api.ci.openshift.org/oauth/token/request
# * oc login api.ci.openshift.org ... # command from the web page
# * oc registry login
# * skopeo copy docker://registry.svc.ci.openshift.org/ocp/builder:golang-1.15 docker:/quay.io/redhat-developer/servicebinding-operator:builder-golang-1.15
FROM quay.io/redhat-developer/servicebinding-operator:builder-golang-1.16 AS builder

USER root

WORKDIR /workspace
COPY / /workspace/

# Build
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager", "--zap-encoder=json"]
