FROM registry.access.redhat.com/ubi8:8.5

ENV WORKSPACE /tmp/workspace
ENV KUBECONFIG_DIR /tmp
ENV KUBECONFIG ${KUBECONFIG_DIR}/kubeconfig
ENV TEST_NAMESPACE default
ENV KEEP_TESTS_RESOURCES false

RUN yum -y --nodocs install git python3 python3-pip && \
    yum clean all
RUN pip3 install --upgrade pip
RUN pip3 --no-cache-dir install --upgrade awscli
RUN yum clean all
RUN curl -SL -o oc.tar.gz https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/latest/openshift-client-linux.tar.gz && \
    tar -xvf oc.tar.gz && \
    chmod +x oc && \
    chmod +x kubectl && \
    mv -vf oc /usr/bin/oc && \
    mv -vf kubectl /usr/bin/kubectl && \
    rm -rf oc.tar.gz

WORKDIR ${WORKSPACE}

ENV PWD ${WORKSPACE}

COPY secret.yaml ${WORKSPACE}/secret.yaml
COPY application.yaml ${WORKSPACE}/application.yaml
COPY sbo.yaml ${WORKSPACE}/sbo.yaml
COPY test-entrypoint.sh ${WORKSPACE}/test-entrypoint.sh

ENTRYPOINT [ "/tmp/workspace/test-entrypoint.sh" ]