FROM quay.io/redhat-developer/go-toolset:builder-golang-1.17 as builder

WORKDIR /workspace
COPY / /workspace/

# Build
RUN make build

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager", "--zap-encoder=json"]
