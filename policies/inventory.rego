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