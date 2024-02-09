# Base inventory module.
# Contains shorthands to simplify rule development.
package lib.konveyor
import future.keywords

deployments[deployment] {
    some list in input.namespaces[_]
    some item in list.items
    item.kind == "Deployment"
    deployment := item
}

pods[pod] {
    some list in input.namespaces[_]
    some item in list.items
    item.kind == "Pod"
    pod := item
}

containers[container] {
    some deployment in deployments
    some item in deployment.spec.template.spec.containers
    container := item
}

containers[container] {
    some pod in pods
    some item in pod.spec.containers
    container := item
}