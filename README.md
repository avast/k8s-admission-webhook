# k8s-admission-webhook [![](https://images.microbadger.com/badges/version/avastsoftware/k8s-admission-webhook.svg)](https://microbadger.com/images/avastsoftware/k8s-admission-webhook "avastsoftware/k8s-admission-webhook image") [![](https://images.microbadger.com/badges/image/avastsoftware/k8s-admission-webhook.svg)](https://microbadger.com/images/avastsoftware/k8s-admission-webhook "avastsoftware/k8s-admission-webhook image")

| Kubernetes version | v1.9                  | v1.10                  | v1.11                  | v1.12                  | v1.13                  |
| ------------------ |-----------------------| -----------------------| -----------------------| -----------------------| -----------------------|
| Build status       | [![Build v1.9][1]][6] | [![Build v1.10][2]][6] | [![Build v1.11][3]][6] | [![Build v1.12][4]][6] | [![Build v1.13][5]][6] |

[1]: https://travis-matrix-badges.herokuapp.com/repos/avast/k8s-admission-webhook/branches/master/1
[2]: https://travis-matrix-badges.herokuapp.com/repos/avast/k8s-admission-webhook/branches/master/2
[3]: https://travis-matrix-badges.herokuapp.com/repos/avast/k8s-admission-webhook/branches/master/3
[4]: https://travis-matrix-badges.herokuapp.com/repos/avast/k8s-admission-webhook/branches/master/4
[5]: https://travis-matrix-badges.herokuapp.com/repos/avast/k8s-admission-webhook/branches/master/5
[6]: https://travis-ci.org/avast/k8s-admission-webhook

A general-purpose Kubernetes [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) to aid with enforcing best practices within your cluster.

Apart from webhook validation, this tool can also do [one-time outside of cluster scan](#on-demand-outside-of-cluster-scan) for objects that are violating validation rules.

Validation features:
* Container setup validation
    * Containers have their resource limits specified (`memory`, `cpu`)
    * Containers have their resource requests specified (`memory`, `cpu`)
    * Containers have readonly root filesystem
* Ingress setup validation
    * Ingress rules hosts and paths match regex
    * Ingress rules hosts and paths are not in collision with definitions already in the cluster
    * Ingress tls hosts match regex
    * Ingress tls hosts are not in collision with definitions already in the cluster
* Annotation validation
    * Annotations match give rule set (See [annotation-rules-example.yaml](annotation-rules-example.yaml))


Validation                 | Operates on
---------------------------| -------------
Container setup            | `Pod`s, `Deployment`s, `ReplicaSet`s, `DaemonSet`s, `Job`s, `CronJob`s, `StatefulSet`s
Ingress setup              | `Ingress`es
Annotation                 | `Pod`s, `Ingress`es

## Introduction
As per [Kubernetes docs](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/), admission webhooks are
> HTTP callbacks that receive admission requests and do something with them. You can define two types of admission webhooks, validating admission webhook and mutating admission webhook. With validating admission Webhooks, you may reject requests to enforce custom admission policies. With mutating admission Webhooks, you may change requests to enforce custom defaults.

This implementation currently acts as a **validating admission webhook** to validate resources against a predefined opinionated set of rules that have to be individually enabled.

## Configuration options
The webhook application can be configured using the flags outlined below.

Note that every option can also be specified via an environment variable. Environment variables should be in uppercase, using `_` instead of `-` as seen in the flag name. E.g.: `--rule-resource-limit-cpu-required` can be alternatively set via an environment variable `RULE_RESOURCE_LIMIT_CPU_REQUIRED=1`.

```
--listen-port int32                                                  Port to listen on. (default 443)
--no-tls                                                             Do not use TLS.
--rule-resource-limit-cpu-must-be-nonzero                            Whether 'cpu' limit in resource specifications must be a nonzero value.
--rule-resource-limit-cpu-required                                   Whether 'cpu' limit in resource specifications is required.
--rule-resource-limit-memory-must-be-nonzero                         Whether 'memory' limit in resource specifications must be a nonzero value.
--rule-resource-limit-memory-required                                Whether 'memory' limit in resource specifications is required.
--rule-resource-request-cpu-must-be-nonzero                          Whether 'cpu' request in resource specifications must be a nonzero value.
--rule-resource-request-cpu-required                                 Whether 'cpu' request in resource specifications is required.
--rule-resource-request-memory-must-be-nonzero                       Whether 'memory' request in resource specifications must be a nonzero value.
--rule-resource-request-memory-required                              Whether 'memory' request in resource specifications is required.
--rule-security-readonly-rootfs-required                             Whether 'readOnlyRootFilesystem' in security context specifications is required.
--rule-security-readonly-rootfs-required-whitelist-enabled           Whether rule 'readOnlyRootFilesystem' in security context can be ignored by container whitelisting.
--rule-resource-violation-message                                    Additional message to be included whenever any of the resource-related rules are violated.
--rule-ingress-collision                                             Whether ingress tls and host collision should be checked 
--rule-ingress-violation-message                                     Additional message to be included whenever any of the ingress-related rules are violated.
--tls-cert-file string                                               Path to the certificate file. Required, unless --no-tls is set.
--tls-private-key-file string                                        Path to the certificate key file. Required, unless --no-tls is set.
--annotations-prefix                                                 What prefix should be used for admission validation annotations.
--annotation-rules-file                                              File with rules to be used for annotations validation (everything is considered valid if not used).
```

In case you want to check `readOnlyRootFilesystem` property globally but also allow some containers requiring writable root filesystem, they can be whitelisted by using annotations.
In order to do that you need to:
* Allow whitelisting of containers for this option by `--rule-security-readonly-rootfs-required-whitelist-enabled` option
* Provide annotation `admission.validation.avast.com/readonly-rootfs-containers-whitelist` at Pod level (prefix can be changed by `--annotations-prefix` option)

Example usage:
```
apiVersion: v1
kind: Pod
metadata:
  name: pod-name
  namespace: test
  annotations:
    admission.validation.avast.com/readonly-rootfs-containers-whitelist: "container-name-1, container-name-2"
spec:
  containers:
    - name: container-name-1
      image: .....
      command: .....
      resources:
        .....
    - name: container-name-2
      image: .....
      command: .....
      resources:
        .....
```


## Installation
The following instructions assume that you will want to deploy the admission webhook inside the cluster, running as a Kubernetes service.

The following Kubernetes resources need to be set up:
* A `Deployment` with the actual webhook container. It runs the HTTP server, which must use TLS with a certificate signed by the Kubernetes cluster CA.
* A `Service`.
* A `ValidatingWebhookConfiguration`, specifying which resource types are to be validated, pointing to the webhook's `Service`. Its configuration must include the cluster's CA bundle.

Because of the above, the necessary Kubernetes manifests will vary based on your cluster configuration.

### Example configuration
See [test/webhook.template.yaml](test/webhook.template.yaml), which contains an example definition of the various Kubernetes resources that might be typically involved when configuring the webhook.

Certain places of the example configuration YAML contain references to variables:
* `${WEBHOOK_IMAGE_NAME}` (in `Deployment`): the Docker image name, e.g. `avastsoftware/k8s-admission-webhook:v0.0.1`
* `${WEBHOOK_TLS_CERT}` (in the `ConfigMap`) and `${WEBHOOK_TLS_PRIVATE_KEY_B64}` in the (`Secret`): webhook server's TLS certificate
  - You can generate an appropriate certificate by running `./test/create-signed-cert.sh --namespace default --service k8s-admission-webhook.default.svc`. You'll need to have `kubectl` on your `PATH`, with its current context pointing at the target cluster.
  - `--namespace` and `--service` in the above command need to reflect the namespace and name of webhook's `Service`, and also need to match the related configuration in the `ValidatingWebhookConfiguration` and the `Service` itself.
  - Running the above command will drop `test/server-cert.pem` and `test/server-key.pem`, which correspond to the server certificate and its private key, respectively.
  - `${WEBHOOK_TLS_CERT}` (in the `ConfigMap`) is the contents of the generated `test/server-cert.pem`.
  - `${WEBHOOK_TLS_PRIVATE_KEY_B64}` (in the `Secret`) is the base-64 encoded contents of the generated `test/server-key.pem`. (E.g. `WEBHOOK_TLS_PRIVATE_KEY_B64=$(cat test/server-key.pem | base64 | tr -d '\n')`)
* `${WEBHOOK_CA_BUNDLE}` (in `ValidatingWebhookConfiguration`) is the cluster's CA bundle.
  - This can be retrieved by running `$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')`. You'll need to have `kubectl` on your `PATH`, with its current context pointing at the target cluster.

## On-demand outside of cluster scan
This tool provides also function to scan cluster (based on **current context** from `~/.kube/config`) for objects that are violating given rules.

Usage:
* Create binary via `make build-binary`
* Run `./k8s-admission-webhook scanner [options]`

Configuration options for cluster scanner:
```
--namespace                                                          Whether specific namespace should be scanned. If omitted, all namespaces are scanned.
--rule-resource-limit-cpu-must-be-nonzero                            Whether 'cpu' limit in resource specifications must be a nonzero value.
--rule-resource-limit-cpu-required                                   Whether 'cpu' limit in resource specifications is required.
--rule-resource-limit-memory-must-be-nonzero                         Whether 'memory' limit in resource specifications must be a nonzero value.
--rule-resource-limit-memory-required                                Whether 'memory' limit in resource specifications is required.
--rule-resource-request-cpu-must-be-nonzero                          Whether 'cpu' request in resource specifications must be a nonzero value.
--rule-resource-request-cpu-required                                 Whether 'cpu' request in resource specifications is required.
--rule-resource-request-memory-must-be-nonzero                       Whether 'memory' request in resource specifications must be a nonzero value.
--rule-resource-request-memory-required                              Whether 'memory' request in resource specifications is required.
--rule-security-readonly-rootfs-required                             Whether 'readOnlyRootFilesystem' in security context specifications is required.
--rule-security-readonly-rootfs-required-whitelist-enabled           Whether rule 'readOnlyRootFilesystem' in security context can be ignored by container whitelisting.
--rule-resource-violation-message                                    Additional message to be included whenever any of the resource-related rules are violated.
--rule-ingress-collision                                             Whether ingress tls and host collision should be checked 
--rule-ingress-violation-message                                     Additional message to be included whenever any of the ingress-related rules are violated.
--annotations-prefix                                                 What prefix should be used for admission validation annotations.
--annotation-rules-file                                              File with rules to be used for annotations validation (everything is considered valid if not used).
```
Note that every option can also be specified via an environment variable. Environment variables should be in uppercase, using `_` instead of `-` as seen in the flag name. E.g.: `--rule-resource-limit-cpu-required` can be alternatively set via an environment variable `RULE_RESOURCE_LIMIT_CPU_REQUIRED=1`.

## Development
The webhook is written in Go and uses [Glide](https://glide.sh/) for dependency management.

### Development cluster & webhook deployment automation
One of the more convenient ways to spin up a dev Kubernetes cluster for local testing is [kubeadm-dind-cluster](https://github.com/kubernetes-sigs/kubeadm-dind-cluster), which runs the cluster using Docker-in-Docker. Makefile targets in this project make use of it to aid with most common development tasks and also enable running integration tests in Travis. You can use the more traditional `minikube` if you want, but you'll need to set up the webhook manually or write your own automation scripts.

These are some `make` targets intended to be used during local development. Note that they require `docker`, `wget`, `openssl` and `envsubst` on your `PATH` and have only been tested on Linux.
* `make dev-start`: Starts the Docker-in-Docker Kubernetes cluster.
* `make deploy-webhook-for-test`: Builds the webhook Docker image and configures the local cluster to use it. This is one of the more involved scripts because it includes generation of an appropriate TLS certificate, retrieving cluster's CA bundle and injecting those into the webhook config. Also, pushes the image to a temporary Docker registry reachable from within the local cluster.
* `make dev-stop`: Stops the local cluster and performs additional cleanup.

#### Running the end-to-end tests
On the CI server, end-to-end tests are run via:
```
make ci-e2e-test KUBERNETES_VERSION=1.13
```

Tests can currently run against versions `1.9` to `1.13` (see [.travis.yml](.travis.yml) for more details).

While `ci-e2e-test` spins up the whole cluster on every run, a more convenient workflow to run end-to-end tests during
local development is available via:
1. `make dev-start`
2. make changes to code + `make dev-e2e-test`, rinse and repeat
3. `make dev-stop` (cleans up the development cluster)

### Why Glide for Go dependency management?
This project makes use of the [official Go client for Kubernetes](https://github.com/kubernetes/client-go), which as of this writing does not support the now-standard [dep](https://github.com/golang/dep) for dependency management. This is discussed in more detail in the ["Dep (Not supported yet!)"](https://github.com/kubernetes/client-go/blob/master/INSTALL.md#dep-not-supported-yet) section of `client-go`'s installation docs, which also [suggest Glide](https://github.com/kubernetes/client-go/blob/master/INSTALL.md#glide) as a viable alternative.

## License

Copyright © 2018, [Petr Krebs](https://github.com/petr-k), [Avast Software](https://avast.github.io/).
Released under the [MIT License](LICENSE).
