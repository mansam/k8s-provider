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

images[image] {
    some deployment in deployments
    some container in deployment.spec.template.spec.containers
    image := {
        "name": container.image,
        "container": container.name,
        "resource": deployment,
    }
}

images[image] {
    some pod in pods
    some container in pod.spec.containers
    image := {
        "image": container.image,
        "container": container.name,
        "resource": pod,
    }
}