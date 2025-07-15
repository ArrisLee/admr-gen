# Kube Admission Review Generator

**Kube Admission Review Generator** (`admr-gen`) is a command-line tool for generating Kubernetes AdmissionReview requests. It is especially useful for testing Gatekeeper constraints and policies (e.g., with Gator), or for simulating admission webhook requests in CI/CD pipelines and policy development.

## Overview

Kubernetes Admission Reviews are a core part of the dynamic admission control system. They are HTTP callbacks that receive admission requests and process them, enabling custom policy enforcement and object mutation at the API server level.

This tool helps you generate realistic AdmissionReview objects from your resource YAMLs, supporting all major operations (`create`, `update`, `delete`) and output formats (`yaml`, `json`).

**References:**
- [Kubernetes Admission Controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [OPA Gatekeeper & Gator](https://open-policy-agent.github.io/gatekeeper/website/docs/gator/)

---

## Features

- Generate AdmissionReview requests for any Kubernetes resource YAML
- Supports `create`, `update`, and `delete` operations
- Output in either YAML or JSON format
- Simulate "oldObject" for update/delete operations
- Useful for policy testing, CI/CD, and admission webhook development

---

## Installation

### Pre-built Binary

Download the latest release from the [Releases page](https://github.com/ArrisLee/admr-gen/releases).

### Go Install

If you have Go installed, you can install via:

```sh
go install github.com/ArrisLee/admr-gen@latest
```

The binary will be placed in your `$GOPATH/bin` or `$HOME/go/bin` directory.

**Tip:** Add your Go bin directory to your `PATH` if you haven't already:

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

---

## Usage

### Command-line Parameters

| Parameter      | Required | Description                                                                                  |
|----------------|----------|----------------------------------------------------------------------------------------------|
| `--file`       | Yes      | Path to the input Kubernetes resource YAML file (e.g., `./deployment.yaml`, `./pod.yaml`)    |
| `--operation`  | No       | Admission operation: `create`, `update`, or `delete`. Defaults to `create`.                  |
| `--output`     | No       | Output format: `yaml` or `json`. Defaults to `yaml`.                                         |

### Basic Example

Generate a create AdmissionReview in YAML:

```sh
admr-gen --file=pod.yaml --operation=create --output=yaml
```

Generate an update AdmissionReview in JSON and save to a file:

```sh
admr-gen --file=pod.yaml --operation=update --output=json > admission_review.json
```

---

## Example

### Input Resource (`pod_sample.yaml`)

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

### Generate Update AdmissionReview

```sh
admr-gen --file=./pod_sample.yaml --operation=update --output=yaml
```

#### Output (YAML)

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
    resource: pods
    version: v1
  uid: <generated>
  userInfo:
    uid: <generated>
    username: fake-k8s-admin-review
```

---

## Advanced Usage

- **Update and Delete Operations:**  
  For `update` and `delete`, the generated AdmissionReview will include an `oldObject` field, simulating the previous state of the resource.
- **Resource Kind Mapping:**  
  The tool automatically maps Kubernetes Kind to the correct resource name (e.g., `Pod` → `pods`).

---

## Contributing

Contributions, issues, and feature requests are welcome!  
Feel free to open an issue or submit a pull request.

---

## License

This project is licensed under the MIT License.

---

If you have any questions or suggestions, please open an issue on the [GitHub repository](https://github.com/ArrisLee/admr-gen).