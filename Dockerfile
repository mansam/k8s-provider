# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset:latest as builder
ENV GOPATH=$APP_ROOT

RUN wget https://github.com/eksctl-io/eksctl/releases/download/v0.176.0/eksctl_Linux_amd64.tar.gz
RUN tar xvf eksctl_Linux_amd64.tar.gz
RUN wget https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v0.6.14/aws-iam-authenticator_0.6.14_linux_amd64
RUN chmod +x aws-iam-authenticator_0.6.14_linux_amd64

COPY --chown=1001:0 . .
RUN make serve

FROM registry.access.redhat.com/ubi9/ubi-minimal

COPY --from=builder /opt/app-root/src/bin/serve /usr/local/bin/k8s-provider
COPY --from=builder /opt/app-root/src/eksctl /usr/local/bin/eksctl
COPY --from=builder /opt/app-root/src/aws-iam-authenticator_0.6.14_linux_amd64 /usr/local/bin/aws-iam-authenticator
ENTRYPOINT k8s-provider --port ${PORT}