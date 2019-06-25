FROM centos:7 as build-tools
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Shoubhik Bose <shbose@redhat.com>"
ENV LANG=en_US.utf8
ENV GOPATH /tmp/go
ARG GO_PACKAGE_PATH=github.com/redhat-developer/service-binding-operator

RUN yum install epel-release -y \
    && yum install --enablerepo=centosplus install -y --quiet \
    findutils \
    git \
    golang \
    make \
    procps-ng \
    tar \
    wget \
    which \
    bc \
    kubectl \
    yamllint \
    python36-virtualenv \
    httpd-tools \
    && yum clean all

# install dep
RUN mkdir -p $GOPATH/bin && chmod a+rwx $GOPATH \
    && curl -L -s https://github.com/golang/dep/releases/download/v0.5.1/dep-linux-amd64 -o dep \
    && echo "7479cca72da0596bb3c23094d363ea32b7336daa5473fa785a2099be28ecd0e3  dep" > dep-linux-amd64.sha256 \
    && sha256sum -c dep-linux-amd64.sha256 \
    && rm dep-linux-amd64.sha256 \
    && chmod +x ./dep \
    && mv dep $GOPATH/bin/dep

ENV PATH=$PATH:$GOPATH/bin

# download, verify and install openshift client tools (oc and kubectl)
WORKDIR /tmp
RUN curl -L -s https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz -o openshift-origin-client-tools.tar.gz \
    && echo "4b0f07428ba854174c58d2e38287e5402964c9a9355f6c359d1242efd0990da3 openshift-origin-client-tools.tar.gz" > openshift-origin-client-tools.sha256 \
    && sha256sum -c openshift-origin-client-tools.sha256 \
    && tar xzf openshift-origin-client-tools.tar.gz \
    && mv /tmp/openshift-origin-client-tools-*/oc /usr/bin/oc \
    && mv /tmp/openshift-origin-client-tools-*/kubectl /usr/bin/kubectl \
    && rm -rf ./openshift* \
    && oc version

# install operator-sdk (from git with no history and only the tag)
RUN mkdir -p $GOPATH/src/github.com/operator-framework \
    && cd $GOPATH/src/github.com/operator-framework \
    && git clone --depth 1 -b v0.7.0 https://github.com/operator-framework/operator-sdk \
    && cd operator-sdk \
    && make dep \
    && make install

RUN mkdir -p ${GOPATH}/src/${GO_PACKAGE_PATH}/

WORKDIR ${GOPATH}/src/${GO_PACKAGE_PATH}

ENTRYPOINT [ "/bin/bash" ]

#--------------------------------------------------------------------

FROM build-tools as builder
ARG VERBOSE=2
COPY . .
RUN make build
#--------------------------------------------------------------------

FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:latest
LABEL com.redhat.delivery.appregistry=true

LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Shoubhik Bose <shbose@redhat.com>"
ENV LANG=en_US.utf8

ENV GOPATH=/tmp/go
ARG GO_PACKAGE_PATH=github.com/redhat-developer/service-binding-operator

COPY --from=builder ${GOPATH}/src/${GO_PACKAGE_PATH}/out/operator /usr/local/bin/service-binding-operator

USER 10001

ENTRYPOINT [ "/usr/local/bin/service-binding-operator" ]