.PHONY: test
## Run unit and integration tests
test: generate fmt vet manifests
	$(GO) test ./... -covermode=atomic -coverprofile cover.out

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...
