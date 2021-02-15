# Service Binding Operator Acceptance Tests

## Pre-requisities

### OpenShift Client (`oc`) v4.5+

Get the `oc` binary from [here](https://mirror.openshift.com/pub/openshift-v4/clients/ocp/), ideally the matching version to your OpenShift or CRC version.

### OpenShift or CRC v4.5+

If you have an OpenShift 4.5+ cluster available, you can use it or get the CodeReady Containers binaries with an embedded OpenShift disk image from [this page](https://cloud.redhat.com/openshift/install/crc/installer-provisioned) and install the `crc` into your system.

Let the version of CRC is `4.6.15`.

Setup CRC:
```bash
crc setup

.
.
.
Setup is complete, you can now run 'crc start' to start the OpenShift cluster
```

Start local CRC instance:
```bash
crc start

.
.
INFO Creating CodeReady Containers VM for OpenShift 4.6.15...
.
.
INFO Starting OpenShift cluster ... [waiting 3m]
.
.
INFO To access the cluster, first set up your environment by following 'crc oc-env' instructions
INFO Then you can access it by running 'oc login -u developer -p developer https://api.crc.testing:6443'
INFO To login as an admin, run 'oc login -u kubeadmin -p ****-*****-*****-***** https://api.crc.testing:6443'
INFO
INFO You can now run 'crc console' and use these credentials to access the OpenShift web console
```

You can also figure out those credentials by running:
```bash
crc console --credentials

To login as a regular user, run 'oc login -u developer -p developer https://api.crc.testing:6443'.
To login as an admin, run 'oc login -u kubeadmin -p ****-*****-*****-***** https://api.crc.testing:6443'

```


## Set environment
`KUBECONFIG` env variable needs to be set for the `oc` to be able to communicate with the cluster

### OpenShift

```bash
export KUBECONFIG=<path-to-kubeconfig>
```

### CRC
```bash
export KUBECONFIG=$HOME/.crc/cache/crc_libvirt_4.6.15/kubeconfig
```

### Verify `oc` is working with your cluster
```bash
oc status

In project default on server https://api.crc.testing:6443
...
```

## Run acceptance tests

Let's use `WORKSPACE` as a placeholder for the path of your local Service Binding Operator repository.

```bash
export WORKSPACE=<path-to-sbo-repo>
cd $WORKSPACE
```

Every output artifact (such as SBO's logs or test reports) related to the acceptance tests can be found under `$WORKSPACE/out/acceptance-tests` directory during and after the test execution.

The acceptance test scenarios are defined in a [Gherkin Syntax](https://cucumber.io/docs/gherkin/), that is `Given -> When -> Then` while the individual steps are in a English-like, human readable, language (e.g. [bindAppToService.feature](https://github.com/redhat-developer/service-binding-operator/blob/master/test/acceptance/features/bindAppToService.feature)).

There are couple of `Makefile` targets that runs acceptance tests:

* `test-acceptance-smoke` - Execute `smoke` sub-set of acceptance tests
* `test-acceptance` - Execute all the acceptance tests

Let's take a look each of them.

### Run all acceptance tests with local instance of SBO

```bash
make test-acceptance
.
.
.
```

This command will prepare the environment so that:
* Deletes the old namespace from any previous run,
* Creates a new namespace with generated name like `test-namespace-xxxxxxxx`,
* Installs all the dependencies such as CRD, role, role-binding, service-account, etc. all for the service binding operator to be able to run,

and:
* Builds and start a local instance of the operator,
* Finally, executes the acceptance tests [features and scenarios](https://github.com/redhat-developer/service-binding-operator/tree/master/test/acceptance/features).

At the end of the acceptance tests execution there's a summary printed out:

```bash
.
.
.

4 features passed, 0 failed, 0 skipped
13 scenarios passed, 0 failed, 4 skipped
149 steps passed, 0 failed, 40 skipped, 0 undefined
Took 24m13.854s
```

### Run acceptance smoke tests with local instance of SBO

There is possibility to run a smaller sub-set of selected acceptance test scenarios as smoke test.

To run acceptance smoke tests with local instance of SBO:

```bash
make test-acceptance-smoke
.
.
.
```

This command will prepare the environment similarly to the previous one and:
* Builds and start a local instance of the operator
* Executes the `smoke` sub-set of acceptance tests [features and scenarios](https://github.com/redhat-developer/service-binding-operator/tree/master/test/acceptance/features) - marked with `@smoke` tag.

### Run acceptance tests with remote instance of SBO

It is possible to run the acceptance tests with SBO already running, that is without building and starting the local instance and without installing the dependent resources (such as CRD, role, role-binding, service-account, etc.) required when the local instance is used. That is particularly useful for testing SBO installed on a remote cluster or running the acceptance tests with various versions of the SBO.

For that there is `TEST_ACCEPTANCE_START_SBO` environment variable that needs to be set to `remote`.

The remote cluster where the SBO should be running is determined by the `KUBECONFIG`.

To run acceptance tests with remote instance of SBO:

```bash
TEST_ACCEPTANCE_START_SBO=remote make test-acceptance
```

Compared with the defaults this command only:
* Deletes the old namespace from any previous run,
* Creates a new namespace with generated name like `test-namespace-xxxxxxxx`,
* Executes the acceptance tests.


### Run acceptance tests with SBO installed from a given index image

This is similar to the running the acceptance tests with a remote instance of the SBO, however in this case the SBO is being installed from a given index image before the acceptance tests are executed.


For that there is `OPERATOR_INDEX_IMAGE_REF` environment variable that can be used to specify the given index image.

The remote cluster where the SBO should be installed is determined by the `KUBECONFIG`.

To run acceptance tests with the remote instance of SBO installed from [OperatorHub.io](https://operatorhub.io/operator/service-binding-operator)'s index image:

```bash
OPERATOR_INDEX_IMAGE_REF=quay.io/operatorhubio/catalog:latest make test-acceptance-with-bundle 
```

To run acceptance tests with the remote instance of SBO installed from the latest master's index image (`OPERATOR_INDEX_IMAGE_REF` variable is set to `quay.io/redhat-developer/servicebinding-operator:index` by default):

```bash
make test-acceptance-with-bundle 
```

### Run a sub-set of features or scenarios

It is possible to run a sub-set of features or even a single scenario (for example, when you are working on a new scenario and you need to run it over and over). The way how to do it is by "marking" that feature or scenario by a "tag" - directly in the particular `*.feature` file - in a form of `@<tag>` (e.g. `@wip`). Something like the following:

```gherkin
Feature: ...

   @wip
   Scenario: Bind service 1 to app 1
   ...

   Scenario: Bind service 1 to app 2
   ...
```

Now, when you want to execute just that scenario out of the whole suite, use the `TEST_ACCEPTANCE_TAGS` environmental variable to set the tag.

To run single scenario tagged with `@wip`:

```bash
TEST_ACCEPTANCE_TAGS=@wip make test-acceptance
.
.
.

```

Only the scenario or feature tagged with the particular tag (`@wip` in case of our example) will be executed, and all of the others will be simply skipped.

```bash
.
.
.

1 feature passed, 0 failed, 3 skipped
1 scenario passed, 0 failed, 16 skipped
10 steps passed, 0 failed, 179 skipped, 0 undefined
Took 0m49.231s
```

## Inspect test results

During the test execution the result of individual steps are shown as the execution goes. That is the most obvious way to see the results, but it is far from elegant - especially with a growing number of scenarios. However, the framework also records the test results in a standardized way of xUnit format, which can be found under `$WORKSPACE/out/acceptance-tests` directory.

As mentioned above those `*.xml` test results files are in xUnit format, which is standardized although it is not human readable. To make it readable a tool for visualizating is needed to help us.

There's one tool in particular that is meant for BDD-like results and that is [Allure](https://docs.qameta.io/allure/)

The following actions use this tool to visualize the results in a form of interactive HTML reports.

### Serve interactive HTML test result report

```bash
make test-acceptance-serve-report
```

This command generates an HTML report and runs a local HTTP server in a container called `test-acceptance-report`. This report can be access at `http://localhost:8088`.

### Generate interactive HTML test results report

```bash
make test-acceptance-generate-report
```

This command uses the `out/acceptance-tests/*.xml` xUnit results and generates a HTML report into `out/acceptance-tests-report` directory that can be served by any HTTP server.

:warning: However, openning the report in Chrome is blocked by a security feature of the Chrome browser that prevents cross-origin requests to the local file system. That results in an error similar to the following:

```txt
Failed to load file:///...: Cross origin requests are only supported for protocol schemes: http, data, chrome, chrome-extension, https.
```

So locally is better to use the above `test-acceptance-serve-report`.

That's all Folks! ...well, at least for now.