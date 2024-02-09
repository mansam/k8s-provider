# k8s-provider
Kubernetes provider for Konveyor analyzer.

## Provider Specific Configuration

The Kubernetes provider requires some details to be provided in the `providerSpecificConfig` portion of its entry in the
analyzer's provider configuration file. The rulesets are written assuming the provider is registered with the analyzer as `k8s`.

* `kubeconfig`: base64-encoded kubeconfig for the cluster that is to be analyzed
* `namespaces`: an array of namespaces that should be analyzed
* `groupVersionKinds`: an array of objects representing GroupVersionKinds to be collected from the cluster and processed.
  * Each object in the array must have the fields `group`, `version` and `kind`.

## Development

1. Build the provider gRPC server with `make serve` and start it with `./bin/serve --port <port>`.
2. Add an entry for the `k8s` provider to the analyzer's provider config file. (see the above section on provider specific config)
3. Run `konveyor-analyzer` with the provider config you wrote and with `--rules` pointing at the `rules/bestpractices` ruleset in this repository.

## Code of Conduct
Refer to Konveyor's Code of Conduct [here](https://github.com/konveyor/community/blob/main/CODE_OF_CONDUCT.md).
