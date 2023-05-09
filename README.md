# Kube Admission Review Generator
A tiny tool to generate kube admission review content, which can be utilized for Gator/Gatekeeper testing purposes.

## Install

Use `go get` to donwload bin file, which will be saved in `$GOPATH/bin` directory.

```sh
go get -u "github.com/ArrisLee/admr-gen"
```

If you haven't add `bin` dir to your system `$PATH`, modify your bash profile or zshrc file by adding following lines:

```sh
export GOPATH=/your/own/go/path
export PATH=$PATH:$GOPATH/bin

```
This will allow you to use go binaries in terminal.


## Usage

Pass `file` and `operation` params to generate different types of kube admission review outputs:

```sh
admr-gen --file=pod.yaml --operation=create
admr-gen --file=pod.yaml --operation=update
admr-gen --file=pod.yaml --operation=delete
```
save output to a yaml file if needed:

```sh
admr-gen --file=pod.yaml --operation=create > example.yaml
```

## Sample output

```yaml
apiVersion: admission.k8s.io/v1
kind: AdmissionReview
request:
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
  oldObject: null
  operation: CREATE
  options: null
  requestKind:
    group: ""
    kind: Pod
    version: v1
  resource:
    group: ""
    resource: Pod
    version: v1
  uid: 7055b44d-d66e-4a3e-a0e7-02329c52b1e0
  userInfo:
    uid: c717f477-06ec-4a47-93e0-0591f421ad06
    username: fake-k8s-admin-review
```
