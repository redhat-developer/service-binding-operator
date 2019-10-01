#!/bin/bash
export EXAMPLE_NAMESPACE="service-binding-demo"

function pathname() {
  DIR="${1%/*}"
  (cd "$DIR" && echo "$(pwd -P)")
}
export HACK_YAMLS=${HACK_YAMLS:-$(pathname $0)/yamls}

## OpenShift Project/Namespace
function create_project {
    oc new-project $EXAMPLE_NAMESPACE
}

function delete_project {
    oc delete project $EXAMPLE_NAMESPACE --ignore-not-found=true
}

## Generic OperatorSources
function print_operator_source {
    REGISTRY_NAMESPACE=$1
    NAME=$2
    sed -e 's,REPLACE_OPSRC_NAME,'$NAME',g' $HACK_YAMLS/operator-source.template.yaml \
    | sed -e 's,REPLACE_REGISTRY_NAMESPACE,'$REGISTRY_NAMESPACE',g'
}

function install_operator_source {
    print_operator_source $@ | oc apply --wait -f -
}

function uninstall_operator_source {
    print_operator_source $@ | oc delete --wait --ignore-not-found=true -f -
}

## Generic operator Subscriptions

function get_current_csv {
    PACKAGE_NAME=$1
    CHANNEL=$2
    oc get packagemanifest $PACKAGE_NAME -o jsonpath='{.status.channels[?(@.name == "'$CHANNEL'")].currentCSV}'
}

function print_operator_subscription {
    PACKAGE_NAME=$1
    OPSRC_NAME=$2
    CHANNEL=$3

    CSV_VERSION=$(get_current_csv $PACKAGE_NAME $CHANNEL)
    sed -e 's,REPLACE_CSV_VERSION,'$CSV_VERSION',g' $HACK_YAMLS/subscription.template.yaml \
    | sed -e 's,REPLACE_CHANNEL,'$CHANNEL',g' \
    | sed -e 's,REPLACE_OPSRC_NAME,'$OPSRC_NAME',g' \
    | sed -e 's,REPLACE_NAME,'$PACKAGE_NAME',g';
}

function install_operator_subscription {
    if [[ ! -z $(oc get packagemanifest | grep $1) ]]; then
        print_operator_subscription $1 $2 $3 | oc apply --wait -f -
    else
        echo "ERROR: packagemanifest $1 not found";
        exit 1;
    fi
}

function uninstall_operator_subscription {
    print_operator_subscription $1 $2 $3 | oc delete --ignore-not-found=true --wait -f -
}

function wait_for_packagemanifest {
    PACKAGE_NAME=$1
    i=1
    while [[ -z "$(oc get packagemanifest | grep $PACKAGE_NAME)" ]] && [ $i -le 5 ]; do
        sleep 5
        i=$(($i+1))
    done
}

function uninstall_current_csv {
    PACKAGE_NAME=$1
    CHANNEL=$2

    oc delete csv $(get_current_csv $PACKAGE_NAME $CHANNEL) -n openshift-operators --ignore-not-found=true
}

## Backing DB (PostgreSQL) Operator

function install_postgresql_operator_source {
    OPSRC_NAMESPACE=pmacik
    OPSRC_NAME=db-operators
    PACKAGE_NAME=db-operators

    install_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
    wait_for_packagemanifest $PACKAGE_NAME
}

function uninstall_postgresql_operator_source {
    OPSRC_NAMESPACE=pmacik
    OPSRC_NAME=db-operators

    uninstall_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
}

function install_postgresql_operator_subscription {
    NAME=db-operators
    OPSRC_NAME=db-operators
    CHANNEL=stable

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_postgresql_operator_subscription {
    NAME=db-operators
    OPSRC_NAME=db-operators
    CHANNEL=stable

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $CHANNEL
}

function install_postgresql_db_instance {
    oc apply -f $HACK_YAMLS/postgresql-database.yaml
}

function uninstall_postgresql_db_instance {
    oc delete -f $HACK_YAMLS/postgresql-database.yaml --ignore-not-found=true
}

## Service Binding Operator

function install_service_binding_operator_source {
    OPSRC_NAMESPACE=redhat-developer
    OPSRC_NAME=redhat-developer-operators
    PACKAGE_NAME=service-binding-operator

    install_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
    wait_for_packagemanifest $PACKAGE_NAME
}

function uninstall_service_binding_operator_source {
    OPSRC_NAMESPACE=redhat-developer
    OPSRC_NAME=redhat-developer-operators

    uninstall_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
}

function install_service_binding_operator_subscription {
    NAME=service-binding-operator
    OPSRC_NAME=redhat-developer-operators
    CHANNEL=stable

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_binding_operator_subscription {
    NAME=service-binding-operator
    OPSRC_NAME=redhat-developer-operators
    CHANNEL=stable

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $CHANNEL
}

## Serverless Operator

function install_serverless_operator_subscription {
    NAME=serverless-operator
    OPSRC_NAME=redhat-operators
    CHANNEL=techpreview

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_serverless_operator_subscription {
    NAME=serverless-operator
    OPSRC_NAME=redhat-operators
    CHANNEL=techpreview

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $CHANNEL
}

## Service Mesh Operator

function install_service_mesh_operator_subscription {
    NAME=servicemeshoperator
    OPSRC_NAME=redhat-operators
    CHANNEL='1.0'

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_mesh_operator_subscription {
    NAME=servicemeshoperator
    OPSRC_NAME=redhat-operators
    CHANNEL=1.0

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $CHANNEL
}


print_operator_subscription servicemeshoperator redhat-operators 1.0
