# Build the manager binary
FROM golang:1.21 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/ cmd/
COPY k8s/ k8s/
COPY provider/ provider/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o k8s-provider cmd/server/main.go

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /workspace/k8s-provider /usr/local/bin/k8s-provider
ENTRYPOINT k8s-provider --port ${PORT}