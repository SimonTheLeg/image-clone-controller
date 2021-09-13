# image-clone-controller

Image-Clone-Controller automates backing up images used in your Kubernetes cluster. It watches Deployments and Daemonsets and will copy any images to a backup-registry of your choosing. Additionally it will replace all the Deployment and Daemonset manifests to exclusively use images from the backup-registry.

## Demo

[![asciicast](https://asciinema.org/a/435553.svg)](https://asciinema.org/a/435553)

## Installation

### A) Running Locally

You can run the controller locally. You can use the `dockerconf` flag to point to a local docker config and the `kubecontext` flag to select a kubeconfig. Keep in mind this method uses your local kubeconfig and should only be used for development purposes.

```sh
go run main.go
```

### B) Running Inside a cluster

1. Create a dockerconfig secret

    ```sh
    kubectl create secret generic image-clone-controller --from-file dockerconfig.json
    ```

    Example `dockerconfig.json`:

    ```json
    {
        "auths": {
            "dockerhub": {
                "auth": "base64 encoded auth here"
            }
        }
    }
    ```

2. Deploy the operator

    ```sh
    kubectl apply -f deployment
    ```

## Developing

### Running Unit Tests

You can run unit tests exclusively, by using the `-short` flag.

```sh
go test ./... -short
```

### Running Integration Tests

You can run integration tests exclusively, by using the `-run Integration` flag. Keep in mind that integration tests require a valid `dockerconfig.json` on your machine.

```sh
go test ./... -Integration
```

## Caveats

* Currently only support `appsv1` k8s-api-version of Deployment and DaemonSet. With the current design it can easily be extended by adding more apiVersions as separate components
* Dockerhub allows for certain images to not include a repo (e.g. `nginx:latest`). The operator supports these images as well. They will be escaped using the library repo (e.g. `{your-repo}/library_nginx:latest`). Functionality of the operator is not affected by this
