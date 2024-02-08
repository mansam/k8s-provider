# gRPC Client Demo

This gRPC client demonstrates the functionality of the provider running in gRPC server mode, and it can be used to run (a feature-limited subset of) analyzer rulesets. 

## Usage

The gRPC client needs to be passed an entire Analyzer provider config, not just the provider specific conf. It will use this configuration to remotely initialize the provider gRPC server. For example:

```json
{
  "name": "k8s",
  "address": "127.0.0.1:8888",
  "proxyConfig": {},
  "initConfig": [{
    "providerSpecificConfig": {
      "groupVersionKinds": [{"group": "apps", "version": "v1", "kind": "Deployment"}, {"group": "", "version": "v1", "kind": "Pod"}, {"group": "", "version": "route.openshift.io/v1", "kind":"Route"}],
      "namespaces": ["konveyor-tackle"],
      "kubeConfigPath": "/home/username/.kube/config",
      "baseModulesPath": "modules/"
    },
    "proxyConfig": {
    }
  }]
}
```

The name must be `k8s`, and the address should be the address on which the gRPC provider is listening. `kubeConfigPath` and `baseModulesPath` are paths that need to be accessible to the the server binary where ever it happens to be running.

The `--rules` flag must be a path to a file containing a json object with an property named `ruleSets` containing an array of paths to analyzer rulesets compatible with the provider. Such a ruleset is included in this repository unde `rules/bestpractices`.

```json
{
  "ruleSets": ["rules/bestpractices"]
}
```

Example command invocations:

```sh
./bin/serve --port 8888
```

```sh
./bin/grpc --config provider.json --rules rules.json | jq > output.json
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