# orderedhttp-operator-v19

Walkthrough of using operator-sdk v0.19.0+, see "ordered-operator" for versions prior.

Example k8s operator using operator-sdk

## What the example looks to accomplish

The goal for this example is to demonstrate a method of launching PODs serially.  Waiting to launch
additional PODs until the currently running POD has reached a **live** state.

Clearly with a basic structure of this single docker image, the same functionality can be accomplished
using Service and StatefulSet features.  This project is meant to be a learning tool to allow
iteration and discovery.

This loosely follows the quickstart from the [Operator-SDK Docs](https://sdk.operatorframework.io/docs/golang/quickstart/)

## Requirements

```bash
brew install go
brew install operator-sdk

go version
# go version go1.14.5 darwin/amd64
operator-sdk version
# operator-sdk version: "v0.19.0", commit: "8e28aca60994c5cb1aec0251b85f0116cc4c9427", kubernetes version: "v1.18.2", go version: "go1.14.4 darwin/amd64"
```

These were the current brew versions as of July 27th, 2020.

---

## Clone and Run

---

If you have a kubernetes cluster available and you want to see this operator in action, follow these
steps.

### Docker Hub Repositories

Logon to your Docker Hub and create yourself two repositories:

- orderedhttp-operator
- nginx-delay

```bash
# Change into a temporary directory
mkdir -p ~/tmp
cd ~/tmp
git clone https://github.com/splicemaahs/orderedhttp-operator-v19.git
cd orderedhttp-operator-v19/nginx-delay
vi Makefile
# change the `splicemaahs` reference to YOUR docker id, so it will push to your newly created
# repository
make build
make push
cd ..
export DOCKER_USERNAME=<docker hub user>
make docker-build IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
make docker-push IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
```

### Install the CustomResourceDefinition to Kubernetes

Ensure you are connected to your K8s cluster and run:

```bash
make install
```

### Create Kubernetes Resources

```bash
# the next commands edits this file: config/default/kustomization.yaml
# these two NEED to match with the second having a suffix of "-"
# An oddity with "kustomize" the "set nameprefix" is CASE sensitive in a very ODD way.
# it is because "nameprefix" is a parameter to "kustomize" and not an arbitrary "path" of yaml.
cd config/default/ && kustomize edit set namespace "orderedhttp-system" && cd ../..
cd config/default/ && kustomize edit set nameprefix "orderedhttp-" && cd ../..
export DOCKER_USERNAME=<docker hub user>
make deploy IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
```

### Validate that the Controller/Operator is running

```bash
kubectl -n orderedhttp-system get deployment -l control-plane=controller-manager
kubectl -n orderedhttp-system get pod -l control-plane=controller-manager
```

### Create the OrderedHttp Resource

Edit the file `config/samples/orderedhttp_v1alpha1_orderedhttp.yaml` to create our
custom CR

```yaml
apiVersion: orderedhttp.splicemachine.io/v1alpha1
kind: OrderedHttp
metadata:
  name: orderedhttp-sample
spec:
  # Add fields here
  replicas: 3
```

```bash
kubectl -n orderedhttp-system create -f config/samples/orderedhttp_v1alpha1_orderedhttp.yaml
```

### Check on Resources

```bash
kubectl -n orderedhttp-system get pods
kubectl -n orderedhttp-system describe OrderedHttp
kubectl logs -f $(kubectl get pods | grep orderedhttp-controller | tr -s ' ' | cut -d' ' -f1) -c manager
```

### Delete Resources

```bash
kubectl delete namespace orderedhttp-system
for i in $(kubectl get clusterrole,clusterrolebinding --no-headers | grep orderedhttp | awk '{ print $1 }'); do kubectl delete ${i}; done
kubectl delete customresourcedefinition orderedhttps.orderedhttp.splicemachine.io
```

---

## Create Operator using operator-sdk

---

### Create Docker Hub Repositories, if not already existing

Logon to your Docker Hub and create yourself two repositories:

- orderedhttp-operator
- nginx-delay

### Build our custom nginx-delay docker image

```bash
# create an ordered-http directory somewhere and change into that directory.
mkdir -p nginx-delay
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/nginx-delay/Dockerfile -o nginx-delay/Dockerfile
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/nginx-delay/Makefile -o nginx-delay/Makefile
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/nginx-delay/hello-plain-text.conf -o nginx-delay/hello-plain-text.conf
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/nginx-delay/nginx-foreground -o nginx-delay/nginx-foreground
chmod 755 nginx-delay/nginx-foreground
cd nginx-delay
vi Makefile
# change the `splicemaahs` reference to YOUR docker id, so it will push to your newly created
# repository
make build
make push
cd ..
```

### Create new operator

```bash
operator-sdk init --domain=splicemachine.io --repo=github.com/splicemaahs/orderedhttp-operator
```

### Add an api / controller

This output produces some warning lines, this appears to be normal as the resulting operator works without issue.

```bash
operator-sdk create api --controller --resource --group=orderedhttp --version=v1alpha1 --kind=OrderedHttp
```

### Add Properties to the api

Create this patch file, then apply to the current sources

```bash
mkdir -p patches
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/patches/apicode.patch -o patches/apicode.patch
```

```bash
git apply patches/apicode.patch
# the 'make' process builds code based on the properties added to the Spec and Status
# sections of ./api/orderedhttp/v1alpha1/orderedhttp_types.go'
make generate
```

### Add reconciler code to the controller

```bash
mkdir -p patches
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/patches/controllercode.patch -o patches/controllercode.patch
```

```bash
# ./pkg/controller/orderedhttp/orderedhttp_controller.go
git apply patches/controllercode.patch
```

### Add Descriptions to the types fields

This is part of the "structured" CRD requirement from Kubernetes 1.16+, this process builds compliant CRD manifests.

```bash
mkdir -p patches
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/patches/typesdescriptions.patch -o patches/typesdescriptions.patch
# ./api/v1alpha1/orderedhttp_types.go
git apply patches/typesdescriptions.patch
make manifests
```

### Add RBAC hints to the Controller code

```bash
mkdir -p patches
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator-v19/master/patches/rbac.patch -o patches/rbac.patch
# ./api/v1alpha1/orderedhttp_types.go
git apply patches/rbac.patch
make manifests
```

### Install the CustomResourceDefinition to Kubernetes (same as above)

Ensure you are connected to your K8s cluster and run:

```bash
make install
```

### Build and Push the Docker Image for the Operator/Controller

You will want to edit the Makefile, find 'docker-build:' and remove 'test' from the line.

```plaintext
# Build the docker image
# docker-build: test
docker-build:
        docker build . -t ${IMG}
```

```bash
export DOCKER_USERNAME=<docker hub user>
make docker-build IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
make docker-push IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
```

### Deploy Operator/Controller to Kubernetes

```bash
# the next commands edits this file: config/default/kustomization.yaml
# these two need to match up until the "system" part
# An oddity with "kustomize" the "set nameprefix" is CASE sensitive in a very ODD way.
# it is because "nameprefix" is a parameter to "kustomize" and not an arbitrary "path" of yaml.
cd config/default/ && kustomize edit set namespace "orderedhttp-system" && cd ../..
cd config/default/ && kustomize edit set nameprefix "orderedhttp-" && cd ../..
export DOCKER_USERNAME=<docker hub user>
make deploy IMG=${DOCKER_USERNAME}/orderedhttp-operator:v0.0.1
```

Output of the deploy will look similar to this:

```plaintext
/Users/cmaahs/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
cd config/manager && /usr/local/bin/kustomize edit set image controller=splicemaahs/orderedhttp-operator:v0.0.1
/usr/local/bin/kustomize build config/default | kubectl apply -f -
namespace/orderedhttp-operator-v19-system created
customresourcedefinition.apiextensions.k8s.io/orderedhttps.orderedhttp.splicemachine.io configured
role.rbac.authorization.k8s.io/orderedhttp-operator-v19-leader-election-role created
clusterrole.rbac.authorization.k8s.io/orderedhttp-operator-v19-manager-role created
clusterrole.rbac.authorization.k8s.io/orderedhttp-operator-v19-proxy-role created
clusterrole.rbac.authorization.k8s.io/orderedhttp-operator-v19-metrics-reader created
rolebinding.rbac.authorization.k8s.io/orderedhttp-operator-v19-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/orderedhttp-operator-v19-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/orderedhttp-operator-v19-proxy-rolebinding created
service/orderedhttp-operator-v19-controller-manager-metrics-service created
deployment.apps/orderedhttp-operator-v19-controller-manager created
```

### Validate that the Controller/Operator is running (same as above)

```bash
kubectl -n orderedhttp-system get deployment -l control-plane=controller-manager
kubectl -n orderedhttp-system get pod -l control-plane=controller-manager
```

### Create an instance of our Kubernetes Custom Resource (same as above)

Edit the file `config/samples/orderedhttp_v1alpha1_orderedhttp.yaml` to create our
custom CR

```yaml
apiVersion: orderedhttp.splicemachine.io/v1alpha1
kind: OrderedHttp
metadata:
  name: orderedhttp-sample
spec:
  # Add fields here
  replicas: 3
```

```bash
kubectl -n orderedhttp-system create -f config/samples/orderedhttp_v1alpha1_orderedhttp.yaml
```

### Check on Resources (same as above)

```bash
kubectl -n orderedhttp-system get pods
kubectl -n orderedhttp-system describe OrderedHttp
kubectl logs -f $(kubectl get pods | grep orderedhttp-controller | tr -s ' ' | cut -d' ' -f1) -c manager
```

### Adjust number of replicas UP

```bash
kubectl patch orderedhttp orderedhttp-sample -p '{"spec":{"replicas": 6}}' --type=merge
kubectl logs -f $(kubectl get pods | grep orderedhttp-controller | tr -s ' ' | cut -d' ' -f1) -c manager
```

### Adjust number of replicas DOWN

```bash
kubectl patch orderedhttp orderedhttp-sample -p '{"spec":{"replicas": 2}}' --type=merge
kubectl logs -f $(kubectl get pods | grep orderedhttp-controller | tr -s ' ' | cut -d' ' -f1) -c manager
```

### Delete Resources (same as above)

```bash
kubectl delete namespace orderedhttp-system
for i in $(kubectl get clusterrole,clusterrolebinding --no-headers | grep orderedhttp | awk '{ print $1 }'); do kubectl delete ${i}; done
kubectl delete customresourcedefinition orderedhttps.orderedhttp.splicemachine.io
```
