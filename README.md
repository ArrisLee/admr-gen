- `--file` - mandatory. Path to the input YAML file, e.g., `./deployment.yaml` or `./pod.yaml`
- `--operation` - optional. Expect operation value in admission review output, available values: `create`, `update` and `delete`. 'create' operation will be applied by default if this param is missing


## Usage

Pass `file` and `operation` params to generate different types of kube admission review outputs:

```sh
admr-gen --file=pod.yaml --operation=create
```
Save output to a yaml file if needed:

```sh
admr-gen --file=pod.yaml --operation=create > example.yaml
```


## Example

Command

```sh
admr-gen --file=./sample_yaml/pod.yaml --operation=update
```

Input

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

Output

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
