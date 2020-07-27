# orderedhttp-operator-v19

Walkthrough of using operator-sdk v0.19.0+, see "ordered-operator" for versions prior.

Example k8s operator using operator-sdk

## What the example looks to accomplish

The goal for this example is to demonstrate a method of launching PODs serially.  Waiting to launch
additional PODs until the currently running POD has reached a **live** state.

Clearly with a basic structure of this single docker image, the same functionality can be accomplished
using Service and StatefulSet features.  This project is meant to be a learning tool to allow
iteration and discovery.

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

## Clone and Run

If you have a kubernetes cluster available and you want to see this operator in action, follow these
steps.

### Docker Hub Repositories

Logon to your Docker Hub and create yourself two repositories:

- orderedhttp-operator
- nginx-delay

```bash
# change into a directory in your $GOPATH/src/
git clone https://github.com/splicemaahs/orderedhttp-operator-v19.git
cd orderedhttp-operator/nginx-delay
vi Makefile
# change the `splicemaahs` reference to YOUR docker id, so it will push to your newly created
# repository
make build
make push
cd ..
operator-sdk build YOURDOCKERID/orderedhttp-operator:latest
docker push YOURDOCKERID/orderedhttp-operator:latest
```

### Create Kubernetes Resources

```bash
kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_crd.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml
kubectl get pods
kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_cr.yaml
```

### Check on Resources

```bash
kubectl get pods
kubectl describe OrderedHttp
kubectl logs $(kubectl get pods | grep orderedhttp-operator | tr -s ' ' | cut -d' ' -f1)
```

### Delete Resources

```bash
kubectl delete -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_cr.yaml
kubectl delete -f deploy/operator.yaml

kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_crd.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml

kubectl get pods
```

## Create Operator using operator-sdk

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
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator/master/patches/apicode.patch -o patches/apicode.patch
```

```bash
git apply patches/apicode.patch
# the 'generate k8s' process builds code based on the properties added to the Spec and Status
# sections of ./pkg/apis/orderedhttp/v1alpha1/orderedhttp_types.go'
operator-sdk generate k8s
```

### Add a controller

```bash
operator-sdk add controller --api-version=orderedhttp.splicemachine.io/v1alpha1 --kind=OrderedHttp
```

### Add reconciler code to the controller

```bash
mkdir -p patches
curl -s https://raw.githubusercontent.com/splicemaahs/orderedhttp-operator/master/patches/controllercode.patch -o patches/controllercode.patch
```

```bash
# ./pkg/controller/orderedhttp/orderedhttp_controller.go
git apply patches/controllercode.patch
```

### Update operator deploy for docker image name

```bash
vi deploy/operator.yaml
# change the 'image: REPLACE_IMAGE' reference to 'image: YOURDOCKERID/orderedhttp-operator:latest'
```

### Build the operator docker image

```bash
go mod vendor # <- you need only run this once, and can rebuild with the 'build' command
# this process will fail on go syntax errors as it builds the code as part of the docker image build.
operator-sdk build YOURDOCKERID/orderedhttp-operator:latest
# push our image to Docker Hub
docker push YOURDOCKERID/orderedhttp-operator:latest
```

### Create Kubernetes Resources (same as above)

```bash
# this installs/defines the Custom Resource Definition
kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_crd.yaml
# these create the ability for the operator to interact with the k8s controller
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
# this deploys the operator pod itself
kubectl create -f deploy/operator.yaml
kubectl get pods
# this creates an instance of our custom 'Kind'
kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_cr.yaml
```

### Check on Resources (same as above)

```bash
kubectl get pods
kubectl describe OrderedHttp
kubectl logs $(kubectl get pods | grep orderedhttp-operator | tr -s ' ' | cut -d' ' -f1)
```

### Delete Resources (same as above)

```bash
kubectl delete -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_cr.yaml
kubectl delete -f deploy/operator.yaml

kubectl create -f deploy/crds/orderedhttp_v1alpha1_orderedhttp_crd.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml

kubectl get pods
```
