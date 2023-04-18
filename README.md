# kustomization-auto-deployer
Automatically deploy **all** commits to a Flux [`GitRepository`](https://fluxcd.io/flux/components/source/gitrepositories/).

## Description
This controller provides a CRD that allows a user to configure the automated deployment of every commit made to a git repository using a Flux [`Kustomization`](https://fluxcd.io/flux/components/kustomize/kustomization/).

Normally, the `GitRepository` polls the `HEAD` of the configured repository with the configured `Interval` and then `Kustomization` are notified of the new commit available and update.

The `GitRepository` will send a [notification](https://fluxcd.io/flux/guides/notifications/) indicating that it has been updated to the `HEAD` commit ID.

If there are multiple commits within the timeframe of the `Interval` when the `GitRepository` polls the repository, it will get all the commits.

The `Kustomization` will send a notification indicating that it has (successfully) deployed the `HEAD` commit ID.

The `Kustomization` doesn't deploy the interim commits separately, there's no way to report that these have been deployed, yes, they are included in the `HEAD` but there's no way to notify on the success or failure of a deployment of an intermediary commit because it's not deployed in a way that we can determine whether or not it would be successful.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/kustomization-auto-deployer:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/kustomization-auto-deployer:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

