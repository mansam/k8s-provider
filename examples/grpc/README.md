# gRPC Client Demo

This gRPC client demonstrates the functionality of the provider running in gRPC server mode, and it can be used to run (a feature-limited subset of) analyzer rulesets.

## Usage

The gRPC client will construct a provider config from its command line arguments and use that to initialize the provider.

The flags are:

* `--server`: address of provider server
* `--kubeconfig`: path to the kubeconfig with the coordinates to the cluster to be analyzed
* `--namespaces`: comma-delimited list of namespaces to analyze
* `--rules`: comma-delimited list of paths to rulesets
* `--stop`: indicates that the provider should be sent the signal to shut down after results are returned.

Example command invocations:

```sh
./bin/serve --port 8888
```

```sh
./bin/grpc --server 127.0.0.1:8888 --kubeconfig .kube/config --namespaces konveyor-tackle --rules rules/bestpractices
```

## Example output

```json
[
  {
    "name": "k8s Best Practices",
    "violations": {
      "lonely-pod": {
        "description": "Pod does not have an OwnerReference. Pod lifecycle should be managed by a resource such as a Deployment to ensure availability.",
        "category": "optional",
        "incidents": [
          {
            "uri": "https://api.crc.testing:6443/api/v1/namespaces/konveyor-tackle/pods/busybox",
            "message": "Pod does not have an OwnerReference. Pod lifecycle should be managed by a resource such as a Deployment to ensure availability.",
            "variables": {
              "apiVersion": "v1",
              "kind": "Pod",
              "name": "busybox",
              "namespace": "konveyor-tackle"
            }
          }
        ],
        "effort": 1
      },
      "low-replica-count": {
        "description": "Deployment has a low replica count",
        "category": "optional",
        "incidents": [
          {
            <snip>
```