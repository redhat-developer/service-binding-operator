#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
HTPASSWD_FILE=$SCRIPT_DIR/openshift-htpass
HTPASSWD_SECRET=acceptance-tests-htpass

USER=acceptance-tests-dev
USER_TOKEN=acceptance-tests-dev

CURRENT_CTX=$(oc config current-context)
oc login -u $USER -p $USER_TOKEN --loglevel=10 --insecure-skip-tls-verify=true

if [ $? -eq 0 ]; then
  echo "User $USER is already set up, skipping..."
else
  oc get secret $HTPASSWD_SECRET -n openshift-config >/dev/null 2>&1 \
    || oc create secret generic $HTPASSWD_SECRET --from-file=htpasswd=$HTPASSWD_FILE -n openshift-config

  oc patch oauth cluster --type merge -p '{"spec":{"identityProviders":[{"htpasswd":{"fileData":{"name":"'$HTPASSWD_SECRET'"}},"mappingMethod":"claim","name":"acceptance-tests","challenge":"true","login":true,"type":"HTPasswd"}]}}'

  retries=50
  until [[ $retries == 0 ]]; do
    oc login -u $USER -p $USER_TOKEN --insecure-skip-tls-verify=true && break
    sleep 5
    retries=$(($retries - 1))
  done;
fi

oc config set-credentials $USER --token=$(oc whoami --show-token)

oc config use-context $CURRENT_CTX