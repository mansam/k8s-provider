# Base inventory module.
# Contains shorthands to simplify rule development.
package lib.konveyor
import future.keywords

deployments[deployment] {
    some item in input
    item.kind == "Deployment"
    deployment := item
}

pods[pod] {
    some item in input
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
