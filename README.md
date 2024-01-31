# k8s-provider
Kubernetes provider for Konveyor analyzer.

## Running from the CLI

The CLI version of the tool can be run to evaluate a single rule at a time. It accepts three parameters: `capability`, `initConfig`, and `conditionInfo`.

* `capability` is the name of the capability needed to evaluate the rule. Currently, the implemented capabilities are `rego.expr` and `rego.module`.
* `initConfig` is a path to the provider init config stored as json.
* `conditionInfo` is a path to a file containing parameters to the capability stored as json.

Examples of the input files can be found in the `examples` directory. An example command invocation would be as follows:

```shell
./bin/cli -initConfig initconfig.json -conditionInfo expression.json -capability rego.expr | jq
```

## Code of Conduct
Refer to Konveyor's Code of Conduct [here](https://github.com/konveyor/community/blob/main/CODE_OF_CONDUCT.md).
