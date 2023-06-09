# Kube Admission Review Generator
A tiny tool to generate Kubernetes Admission Review Requests, which can be utilized for Gatekeeper constraint/policy testing purposes (e.g., Gator test).

Admission Reviews in Kubernetes are part of the dynamic admission control system which are HTTP callbacks that receive admission requests and process them. They are an integral part of the Kubernetes API and are used to govern and enforce custom policies or modifications on objects.

Relevant links:

https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/

https://open-policy-agent.github.io/gatekeeper/website/docs/gator/

## Installation

Binary file can be directly dowloaded in the release page [here](https://github.com/ArrisLee/admr-gen/releases)

OR：

Use `go get` to donwload and install the file if you have Golang env ready.

The tool will be installed in `$GOPATH/bin` directory.

```sh
go get -u "github.com/ArrisLee/admr-gen"
```

If you haven't add `bin` dir to your system `$PATH`, modify your bash profile by adding following lines:

```sh
export GOPATH=/your/own/go/path
export PATH=$PATH:$GOPATH/bin

```
This will allow you to use installed go binaries in terminal.

## Parameters

- `--file` - mandatory. Path to the input resource YAML file, e.g., `./deployment.yaml` or `./pod.yaml`.
- `--operation` - optional. Expect operation value in admission review output, available values: `create`, `update` and `delete`. There will be an extra section in the generated yaml file called `OldOBject` when using `update` or `delete` operations.`create` operation will be applied if this parameter is missing.
- `--output` - optional. Output format can be either `json` or `yaml`. `yaml` format will be applied if this parameter is missing.


## Usage

Pass `file` and `operation` params to generate different types of kube admission review outputs:

```sh
admr-gen --file=pod.yaml --operation=create --output=yaml
```
Save output into a file if needed:

```sh
admr-gen --file=pod.yaml --operation=create --output=json > example.json
```


## Example

Command

```sh
admr-gen --file=./pod_sample.yaml --operation=update
admr-gen --file=./pod_sample.yaml --operation=delete --output=json
```

Input file

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: allowed
  namespace: test
spec:
  serviceAccountName: test-user
  containers:
    - name: test
      image: openpolicyagent/opa:0.9.2
      args:
        - "run"
        - "--server"
        - "--addr=localhost:8080"
      resources:
        limits:
          cpu: "100m"
          memory: "30Mi"
```

Output YAML

```yaml
apiVersion: admission.k8s.io/v1
kind: AdmissionReview
request:
  dryRun: true
  kind:
    group: ""
    kind: Pod
    version: v1
  object:
    apiVersion: v1
    kind: Pod
    metadata:
      name: allowed
      namespace: test
    spec:
      containers:
      - args:
        - run
        - --server
        - --addr=localhost:8080
        image: openpolicyagent/opa:0.9.2
        name: test
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
      serviceAccountName: test-user
  oldObject:
    apiVersion: v1
    kind: Pod
    metadata:
      name: allowed-old
      namespace: test
    spec:
      containers:
      - args:
        - run
        - --server
        - --addr=localhost:8080
        image: openpolicyagent/opa:0.9.2
        name: test
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
      serviceAccountName: test-user
  operation: UPDATE
  options: null
  requestKind:
    group: ""
    kind: Pod
    version: v1
  resource:
    group: ""
    resource: Pod
    version: v1
  uid: 8f248c68-639d-452f-a28f-3f331b001821
  userInfo:
    uid: 502ab568-4acd-4776-8326-b10b5414eb6b
    username: fake-k8s-admin-review
```

Output JSON

```json
{
    "kind": "AdmissionReview",
    "apiVersion": "admission.k8s.io/v1",
    "request": {
        "uid": "d37115b6-3de2-4ad8-b58e-0ef6cc5c71cd",
        "kind": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        },
        "resource": {
            "group": "",
            "version": "v1",
            "resource": "Pod"
        },
        "requestKind": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        },
        "operation": "DELETE",
        "userInfo": {
            "username": "fake-k8s-admin-review",
            "uid": "e888069a-3103-454e-8ccd-e3bb745c43ed"
        },
        "object": null,
        "oldObject": {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "name": "allowed",
                "namespace": "test"
            },
            "spec": {
                "containers": [
                    {
                        "args": [
                            "run",
                            "--server",
                            "--addr=localhost:8080"
                        ],
                        "image": "openpolicyagent/opa:0.9.2",
                        "name": "test",
                        "resources": {
                            "limits": {
                                "cpu": "100m",
                                "memory": "30Mi"
                            }
                        }
                    }
                ],
                "serviceAccountName": "test-user"
            }
        },
        "dryRun": true,
        "options": null
    }
}
```