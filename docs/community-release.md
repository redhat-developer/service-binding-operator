# Community Release

1. Ensure all necessary PRs merged to master
2. Based on the changes from last release, decide the new version number 
3. Prepare release notes
4. Update `VERSION` variable in Makefile and send PR
5. After the PR is merged, run `make prepare-operatorhub-pr` and follow the reported instructions on how to submit PR for [OperatorHub](https://operatorhub.io)
6. Tag the released commit

## GitHub Release Notes

After the operator is published on OperatorHub, create [the new release
on GitHub](https://github.com/redhat-developer/service-binding-operator/releases/new)

Release Title: vX.Y.Z

```
# Improvements since vX.Y.Z-1

# Bugfixes

# Other Changes
```

Generate `release.yaml` file with `make release-manifests` and attach it to the GitHub release.