# Build the manager binary
FROM golang:1.16 as arm_builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN GOPROXY="https://goproxy.cn" GO111MODULE=on go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOPROXY="https://goproxy.cn" GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM alpine:latest@sha256:bd9137c3bb45dbc40cde0f0e19a8b9064c2bc485466221f5e95eb72b0d0cf82e
# FROM golang:1.13
WORKDIR /
COPY --from=arm_builder /workspace/manager .
# USER nonroot:nonroot
USER root:root

ENTRYPOINT ["/manager"]
