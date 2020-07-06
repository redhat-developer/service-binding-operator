# Continuous Integration

Service binding operator primarily uses [OpenShift CI][openshift-ci] for
continuous integration.  The Openshift CI is built using
[CI-Operator][ci-operator].  The Service Binding Operator specific configuration
is located here:
https://github.com/openshift/release/tree/master/ci-operator/config/redhat-developer/service-binding-operator

As part of the continuous integration, there are several jobs configured to run
against pull requests in GitHub.  The CI jobs are triggered whenever there is a
new pull request from the team members.  It is also triggered when there is a
new commit in the current pull request.  If a non-member creates the pull
request, there should be a comment with text as `/ok-to-test` from any member to
run the CI job.

Here is a high-level schematic diagram of the CI pipeline for an end-to-end
test:

```
+--------+     +--------+     +--------+
|        |     |        |     |        |
|  root  +---->+  src   +---->+  bin   |
|        |     |        |     |        |
+--------+     +--------+     +---+----+
                                   |
    ,------------------------------'
    |
    v
+---+----+     +--------+
|        |     |        |
| images +---->+ tests  |
|        |     |        |
+--------+     +--------+
```

For lint and unit test, the schematic diagram is as follows:

```
+--------+     +--------+     +----------------+
|        |     |        |     |                |
|  root  +---->+  src   +---->+ lint/unit test |
|        |     |        |     |                |
+--------+     +--------+     +----------------+
```


All the steps mentioned below are taking place inside a temporary work
namespace.  When you click on the job details from your pull request, you can
see the name of the work namespace in the dashboard UI.  The name starts with
`ci-op-`.  The images created goes into this temporary work namespace.  At the
end of every image build, the [work namespace name has set as an environment
variable][namespace] called `OPENSHIFT_BUILD_NAMESPACE`.  This environment
variable is used to refer the image URLs inside the configuration files.

## root

As part of the CI pipeline, the first step is to create the `root` image.  In
fact, `root` is a tag created for the pipeline image.  This image contains all
the tools including but not limited to Go compiler, git, kubectl, oc, and
Operator SDK.

The `root` image tag is created using this Dockerfile:
`openshift-ci/Dockerfile.tools`.

## src

This step clones the pull request branch with the latest changes.  The cloning
is taking place inside a container created from the `root` image.  As a result
of this step, an image named `src` is going to be created.

If your pull request depends on any package installed through `yum install`
inside the `root` image, those changes should be sent through a different PR and
merged first.  As you can see from the above diagrams, the `src` image gets
built after the `root` image, whereas the pull request branch gets merged in the
`src` image.

## bin

This step runs the `build` Makefile target.  This step is taking place
inside a container created from the `src` image created in the
previous step.

The `make build` produces an operator binary image available under `./out`
directory.  As a result of this step, a container image named `bin` is going to
be created.

## tests

### lint

The lint runs the [GolangCI Lint][golangci], [YAML Lint][yaml-lint], and
[Operator Courier][operator-courier].  GolangCI is a Go program, whereas the
other two are Python based.  So, Python 3 is a prerequisite to run lint.

The GolangCI Lint program runs multiple Go lint tools against the repository.
GolangCI Lint runs lint concurrently and completes execution in a few seconds.
As of now, there is no configuration provided to run GolangCI Lint.

The YAML Lint tools validate all the YAML configuration files.  It
excludes the `vendor` directory while running.  There is a
configuration file at the top-level directory of the source code:
`.yamllint`.

### unit

The `unit` target runs the unit tests.  Some of the tests make use of mock
objects. The unit tests don't require a dedicated OpenShift cluster, unlike
end-to-end tests. It runs `test-unit` Makefile target.

### e2e

The `e2e` run an end-to-end test against an operator running inside the CI
cluster pod but connected to a freshly created temporary Openshift 4 cluster.
It makes use of the `--up-local` option provided by the Operator SDK testing
tool.  It runs `test-e2e` Makefile target.

## Debugging e2e test

1. Login to https://api.ci.openshift.org
2. Copy login command from top-right corner
3. Note down the namespace from log.  As mentioned before, when you click on the
   job details from your pull request, you can see the name of the work
   namespace in the dashboard UI.  The name starts with `ci-op-`.
4. Now you can check each container log or rsh into shell.
5. In the container named `setup` in `e2e` pod, you will get kubeconfig for the
   temporary cluster.  You can use it to access the temporary cluster and
   perform further debugging.

[openshift-ci]: https://github.com/openshift/release
[ci-operator]: https://github.com/openshift/release/tree/master/ci-operator
[golangci]: https://github.com/golangci/golangci-lint
[yaml-lint]: https://github.com/adrienverge/yamllint
[operator-courier]: https://github.com/operator-framework/operator-courier
[namespace]: https://docs.okd.io/latest/dev_guide/builds/build_output.html#output-image-environment-variables
