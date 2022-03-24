# How to Contribute

:+1::tada: First off, thanks for taking the time to contribute!

You can look at the issues [with help wanted label][help-wanted] for items that
you can work on.

If you need help, please feel free to reach out to our [slack channel](https://app.slack.com/client/T09NY5SBT/C019LQYGC5C)!

When contributing to this repository, please first discuss the change you wish
to make via issue, email, or any other method with the owners of this repository
before making a change.  Small pull requests are easy to review and merge.  So,
please send small pull requests.

Please note we have a [code of conduct][conduct], please follow it in all your
interactions with the project.

Contributions to this project should conform to the [Developer Certificate of
Origin][dco].  See the [next section](#sign-your-work) for more details.

## Sign Your Work

Contributions to this project should conform to the [Developer Certificate of
Origin][dco].  You need to sign-off your git commits before sending the pull
requests.  The sign-off is a single line of text at the end of the commit
message.  The signature consists of your official name and email address.  These
two details should match with the name and email address used in the Git commit.
All your commits needs to be signed.  Your signature certifies that you wrote
the patch or otherwise have the right to contribute the material.  The rules are
pretty simple, if you can certify the below (from
[developercertificate.org][dco]):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@example.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

Note: If your git config information is set properly then viewing the `git log`
information for your commit will look something like this:

```
Author: Joe Smith <joe.smith@example.com>
Date:   Thu Feb 2 11:41:15 2018 -0800

    Update README

    Signed-off-by: Joe Smith <joe.smith@example.com>
```

Notice the `Author` and `Signed-off-by` lines match. If they don't
your PR will be rejected by the automated DCO check.


## Getting Started
* Fork the repository on GitHub
* Read the [README](README.md) file for build and test instructions
* Run the examples
* Explore the the project
* Submit issues and feature requests, submit changes, bug fixes, new features

## How Can I Contribute?

### Reporting Bugs

Bugs are tracked as
[GitHub issues](https://github.com/redhat-developer/service-binding-operator/issues).
Before you log a new bug, review the existing bugs to determine if the problem
that you are seeing has already been reported. If the problem has not already
been reported, then you may log a new bug.

Please describe the problem fully and provide information so that the bug
can be reproduced. Document all the steps that were performed, the
environment used, include screenshots, logs, and any other information
that will be useful in reproducing the bug.

### Suggesting Enhancements

Enhancements and requests for new features are also tracked as [GitHub issues](https://github.com/redhat-developer/service-binding-operator/issues) or [Enhancement Proposals](docs/proposals/README.md)

As is the case with bugs, review the existing feature requests before logging
a new request.

## Pull Requests

All submitted code and document changes are reviewed by the project
maintainers through pull requests.

To submit a bug fix or enmhancement, log an issue in github, create a new
branch in your fork of the repo and include the issue number as a prefix in
the name of the branch. Include new tests to provide coverage for all new
or changed code. Create a pull request when you have completed code changes.
Include an informative title and full details on the code changed/added in
the git commit message and pull request description.

Before submitting the pull request, verify that all existing tests run
cleanly.

Be sure to run yamllint on all yaml files included in pull requests. Ensure
that all text in files in pull requests is compliant with:
[.editorconfig](.editorconfig)

Each Pull Request is expected to meet the following expectations around:

* [Pull Request Description](#pull-request-description)
* [Commits](#commits)
* [Docs](#docs)
* [Functionality](#functionality)
* [Code](#code)
* [Tests](#tests)

## Pull request description

 Include a link to the issue being addressed, but describe the context for the reviewer
  * If there is no issue, consider whether there should be one:
    * New functionality must be designed and approved, may require a TEP
    * Bugs should be reported in detail
  * If the template contains a checklist, it should be checked off
  * Release notes filled in for user visible changes (bugs + features),
    or removed if not applicable (refactoring, updating tests) (may be enforced
    via the [release-note Prow plugin](https://github.com/tektoncd/plumbing/blob/main/prow/plugins.yaml))

### Commits

* Use the body to explain [what and why vs. how](https://chris.beams.io/posts/git-commit/#why-not-how).
  Link to an issue whenever possible and [aim for 2 paragraphs](https://www.youtube.com/watch?v=PJjmw9TRB7s),
  e.g.:
  * What is the problem being solved?
  * Why is this the best approach?
  * What other approaches did you consider?
  * What side effects will this approach have?
  * What future work remains to be done?
* Prefer one commit per PR. For multiple commits ensure each makes sense without the context of the others.
* As much as possible try to stick to these general formatting guidelines:
  * Separate subject line from message body.
  * Write the subject line using the "imperative mood" ([see examples](https://chris.beams.io/posts/git-commit/#imperative)).
  * Keep the subject to 50 characters or less.
  * Try to keep the message wrapped at 72 characters.
  * Check [these seven best practices](https://chris.beams.io/posts/git-commit/#seven-rules) for more detail.

### Example Commit Message

Here's a commit message example to work from that sticks to the spirit
of the guidance outlined above:

```
Add example commit message to demo our guidance

Prior to this message being included in our standards there was no
canonical example of an "ideal" commit message for devs to quickly copy.

Providing a decent example helps clarify the intended outcome of our
commit message rules and offers a template for people to work from. We
could alternatively link to good commit messages in our repos but that
requires developers to follow more links rather than just showing
what we want.
```

### Docs

* Include AsciiDoc [doc updates](docs/userguide) for user visible features
* We use [Antora framework](https://docs.antora.org/antora/latest/) for AsciiDoc rendering 
* Spelling and grammar should be correct
* Try to make formatting look as good as possible (use preview mode to check, i.e. render the content locally using `make site`)
* Should explain thoroughly how the new feature works
* If possible, in addition to code snippets, include a reference to an end to end example
* Ensure that all links and references are valid

### Functionality

It should be safe to cut a release at any time, i.e. merging this PR should not
put us into an unreleasable state

## Code

* Reviewers are expected to understand the changes well enough that they would feel confident
  saying the understand what is changing and why:
  * Read through all the code changes
  * Read through linked issues and pull requests, including the discussions
* Prefer small well factored packages with unit tests
* [Go Code Review comments](https://github.com/golang/go/wiki/CodeReviewComments)
  * All public functions and attributes have docstrings
  * Don’t panic
  * Error strings are not capitalized
  * Handle all errors ([gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully))
    * When returning errors, add more context with `fmt.Errorf` and `%v`
  * Use meaningful package names (avoid util, helper, lib)
  * Prefer short variable names

### Tests

* New features (and often/whenever possible bug fixes) have one or all of:
  * Unit tests
  * End to end, i.e. acceptance tests
* Unit tests:
  * Coverage should remain the same or increase
 
## Running Acceptance Tests

1. Set KUBECONFIG for both minikube and acceptance tests (it will be generated at minikube's start if it does not exist):

```
export KUBECONFIG=/tmp/minikubeconfig
```

2. Start minikube:

```
minikube start
```

3. Enable olm on minikube:


```
minikube addons enable olm
```

4. Deploy operator to the minikube cluster

```
eval $(minikube docker-env)
make deploy OPERATOR_REPO_REF=$(minikube ip):5000/sbo
```

5. Execute all acceptance tests tagged with `@dev` using `kubectl` CLI:

```
make test-acceptance TEST_ACCEPTANCE_TAGS="@dev" TEST_ACCEPTANCE_START_SBO=remote TEST_ACCEPTANCE_CLI=kubectl
```

To run a specific test:

```
make test-acceptance TEST_ACCEPTANCE_START_SBO=remote TEST_ACCEPTANCE_CLI=kubectl EXTRA_BEHAVE_ARGS='-n "Specify path of secret in the Service Binding"'
```

# Pull Request Workflow

- Fork the repository and clone it your work directory
- Create a topic branch from where you want to base your work
  - This is usually the `master` branch.
  - Only target release branches if you are certain your fix must be on that
    branch.
  - To quickly create a topic branch based on `master`; `git checkout -b
    my-bug-fix upstream/master` (Here `upstream` is alias for the remote repo)
- Make commits of logical units
- Make sure your commit messages are in [the proper format][commit-message].
- Push your changes to a topic branch in your fork of the repository
- Submit a pull request

Example:

```shell
git remote add upstream https://github.com/kubepreset/kubepreset.git
git fetch upstream
git checkout -b my-bug-fix upstream/master
git commit -s
git push origin my-bug-fix
```

### Staying in sync with upstream

When your branch gets out of sync with the `upstream/master` branch, use the
following to update:

``` shell
git checkout my-bug-fix
git fetch upstream
git rebase upstream/master
git push --force-with-lease origin my-bug-fix
```

### Updating pull requests

If your PR fails to pass CI or needs changes based on code review, you'll most
likely want to squash these changes into existing commits.

If your pull request contains a single commit or your changes are related to the
most recent commit, you can simply amend the commit.

```
git add .
git commit --amend
git push --force-with-lease origin my-bug-fix
```

If you need to squash changes into an earlier commit, you can use:

```
git add .
git commit --fixup <commit>
git rebase -i --autosquash main
git push --force-with-lease origin my-bug-fix
```

Please add a comment in the PR indicating your new changes are ready to review.

[help-wanted]: https://github.com/redhat-developer/service-binding-operator/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22
[conduct]: https://github.com/redhat-developer/service-binding-operator/blob/master/CODE_OF_CONDUCT.md
[dco]: http://developercertificate.org
[commit-message]: https://chris.beams.io/posts/git-commit/
