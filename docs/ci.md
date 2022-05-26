# Continuous Integration

Service binding operator uses [OpenShift CI][openshift-ci-docs] for
continuous integration along with [GitHub Workflows][github-workflows].  The Openshift CI is built using
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
|  root  +---->+  src   +---->+  test  |
|        |     |        |     |        |
+--------+     +--------+     +--------+
```

All the steps mentioned below are taking place inside a temporary work
namespace.  When you click on the job details from your pull request, you can
see the name of the work namespace in the dashboard UI. The name starts with
`ci-op-`.

## root

As part of the CI pipeline, the first step is to create the `root` image.  In
fact, `root` is a tag created for the `pipeline` image.  This image contains all
the tools including but not limited to Go compiler, git, kubectl, oc, and
Operator SDK.

The `root` image tag is created using a [Dockerfile](https://github.com/openshift/release/blob/master/ci-operator/config/redhat-developer/service-binding-operator/redhat-developer-service-binding-operator-master__4.11.yaml#L3).

## src

This step clones the pull request branch with the latest changes.  The cloning
is taking place inside a container created from the `root` image.  As a result
of this step, an image named `src` is going to be created.

If your pull request depends on any package installed through `yum install`
inside the `root` image, those changes should be sent through a different PR and
merged first.  As you can see from the above diagrams, the `src` image gets
built after the `root` image, whereas the pull request branch gets merged in the
`src` image.

## test
### acceptance

The `4.x-acceptance` run acceptance tests against an operator installedin the freshly created temporary Openshift 4.x cluster.
It runs `test-acceptance-with-bundle` Makefile target.

## Debugging acceptance tests in OpenShift CI

1. Click on the `Details` link at respective PR check
2. Log in with `RedHat_Internal_SSO` identity provider
3. Note down the namespace from log.  As mentioned before, when you click on the
   job details from your pull request, you can see the name of the work
   namespace in the dashboard UI.  The name starts with `ci-op-`.
4. Now you can check each container logs or rsh into shell.
5. In the container named `test` in `acceptance-ipi-install-install` pod, you will get kubeconfig for the
   temporary cluster in the `$KUBECONFIG` environment variable as well as the `kubeadmin` user password for accessing the OpenShift Console stored in a file mentioned in the `KUBEADMIN_PASSWORD_FILE`.  You can use it to access the temporary cluster and
   perform further debugging.

[openshift-ci-docs]: https://docs.ci.openshift.org
[ci-operator]: https://github.com/openshift/release/tree/master/ci-operator
[golangci]: https://github.com/golangci/golangci-lint
[yaml-lint]: https://github.com/adrienverge/yamllint
[operator-courier]: https://github.com/operator-framework/operator-courier
[namespace]: https://docs.okd.io/latest/dev_guide/builds/build_output.html#output-image-environment-variables
[github-workflows]: https://github.com/redhat-developer/service-binding-operator/actions
