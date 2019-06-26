FROM centos:7 as build-tools
ENV LANG=en_US.utf8
ENV GOPATH /tmp/go
ARG GO_PACKAGE_PATH=github.com/redhat-developer/service-binding-operator

ENV GIT_COMMITTER_NAME devtools
ENV GIT_COMMITTER_EMAIL devtools@redhat.com

RUN yum install epel-release -y \
    && yum install --enablerepo=centosplus install -y --quiet \
    findutils \
    git \
    make \
    gcc \
    procps-ng \
    tar \
    wget \
    which \
    bc \
    yamllint \
    python36-virtualenv \
    && yum clean all

ENV PATH=:$GOPATH/bin:/tmp/goroot/go/bin:$PATH

WORKDIR /tmp

RUN mkdir -p $GOPATH/bin
RUN mkdir -p /tmp/goroot

RUN curl -Lo go1.12.6.linux-amd64.tar.gz https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz && tar -C /tmp/goroot -xzf go1.12.6.linux-amd64.tar.gz
RUN curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.14.3/bin/linux/amd64/kubectl && chmod +x kubectl && mv kubectl $GOPATH/bin/

RUN curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.8.1/operator-sdk-v0.8.1-x86_64-linux-gnu && chmod +x operator-sdk && mv operator-sdk $GOPATH/bin/

RUN mkdir -p ${GOPATH}/src/${GO_PACKAGE_PATH}/

WORKDIR ${GOPATH}/src/${GO_PACKAGE_PATH}

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