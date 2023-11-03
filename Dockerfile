# Build the manager binary
ARG builder_image

# Build architecture
ARG ARCH

FROM ${builder_image} as builder
WORKDIR /workspace

# Run this with docker build --build-arg goproxy=$(go env GOPROXY) to override the goproxy
ARG goproxy=https://proxy.golang.org
# Run this with docker build --build-arg package=./controlplane or --build-arg package=./bootstrap
ENV GOPROXY=$goproxy

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY ./ ./

ARG package=.
ARG ARCH
ARG ldflags

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} \
    go build -trimpath -ldflags "${ldflags} -extldflags '-static'" \
    -o aws-janitor .

FROM gcr.io/distroless/static:nonroot-${ARCH}
LABEL org.opencontainers.image.source=https://github.com/rancher-sandbox/aws-janitor
WORKDIR /
COPY --from=builder /workspace/aws-janitor .
# Use uid of nonroot user (65532) because kubernetes expects numeric user when applying pod security policies
USER 65532
ENTRYPOINT ["/aws-janitor"]