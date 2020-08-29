# Community Release

1. Ensure all necessary PRs merged to master
2. Based on the changes from last release, decide the new version number
3. Prepare release notes and announcement mail
4. Update `OPERATOR_VERSION` in Makefile and send PR
5. After the PR created in step 4 merged, copy files in 
   https://github.com/redhat-developer/service-binding-operator-manifests
   and send PR to https://github.com/operator-framework/community-operators
6. After the PR created in step 5 merged, create release with `OPERATOR_VERSION` + COMMIT COUNT
   This should be performed directly in GitHub (TODO: need exploration)
7. Send announcement mail to service-binding-support@redhat.com and openshift-dev-services@redhat.com

## Release Notes

Relase Title: vX.Y.Z

```
# Changes since vX.Y.Z-1

(For pre-releases use a message like this:)
:rotating_light: This is a RELEASE CANDIDATE. Use it only for testing purposes, if you find any bugs file an issue.
:rotating_light: This is a BETA release. Use it only for testing purposes, if you find any bugs file an issue.
:rotating_light: This is an ALPHA release. Use it only for testing purposes, if you find any bugs file an issue.

## :loudspeaker: Major announcement title

## :warning: Breaking Changes

## :sparkles: New Features

## :bug: Bug Fixes

## :book: Documentation

## :seedling: Others

## Install/Upgrade

(TODO: Instruction for install or upgrade)

## Uninstall

(TODO: Instruction for uninstall)

Thanks to all our contributors! :blush:

```

Remove any section that is empty.


## Announcement Mail

Subject: Service Binding Operator X.Y.Z released!

```
Service Binding Operator X.Y.Z  is now available for download.
For more details please see the release notes at:

https://github.com/redhat-developer/service-binding-operator/releases/tag/vX.Y.Z

To get started with using the operator, see the docs here:

https://github.com/redhat-developer/service-binding-operator

What is Service Binding Operator?
---------------------------------

The goal of the Service Binding Operator is to enable application authors to
import an application and run it on Kubernetes with services such as databases
represented as Kubernetes objects including Operator-backed and chart-based
backing services, without having to perform manual configuration of Secrets,
ConfigMaps, etc.

Notable changes since X.Y.Z-1
-----------------------------

- Major features
- and bugfixes
- with highlighted backward incompatible changes

Thanks to everyone who has contributed to this release.
```
