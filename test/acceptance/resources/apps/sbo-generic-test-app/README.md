Follow instructions on https://github.com/docker/buildx#building-multi-platform-images
and https://docs.docker.com/buildx/working-with-buildx/

Build and push with

```shell
docker buildx build --push \
--platform "linux/amd64,linux/ppc64le,linux/arm64,linux/s390x" \
-t quay.io/service-binding/generic-test-app:YYYYMMDD .
```