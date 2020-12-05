#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
HTPASSWD_FILE=$SCRIPT_DIR/openshift-htpass
HTPASSWD_SECRET=acceptance-tests-htpass

oc get secret $HTPASSWD_SECRET -n openshift-config >/dev/null 2>&1 \
  || oc create secret generic $HTPASSWD_SECRET --from-file=htpasswd=$HTPASSWD_FILE -n openshift-config

oc apply -f - <<EOF
apiVersion: config.openshift.io/v1
kind: OAuth
metadata:
  name: cluster
spec:
  identityProviders:
  - name: acceptance-tests
    challenge: true
    login: true
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: $HTPASSWD_SECRET
EOF

CURRENT_CTX=$(oc config current-context)

oc login -u acceptance-tests-dev -p acceptance-tests-dev

oc config use-context $CURRENT_CTX