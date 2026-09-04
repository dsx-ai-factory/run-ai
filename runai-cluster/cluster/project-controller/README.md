# project-controller
Lifecycle management of cluster scheduling resources based on Projects

![image](https://user-images.githubusercontent.com/10087174/126780612-d6c0139a-246e-4bb1-bfcd-ffba3f970a18.png)

Built using the [Operator SDK](https://sdk.operatorframework.io/) this controller listens for events on our Project crd and reconciles the following based on the crd:
* Namespace
* Queue
* Limit Range
* Image pull secret
* RBAC (TBD)

This controller listens for events on all aforementioned resources, and responds to each such event as if a change was made to the related Project.

It reconciles each resource by comparing it to the Project spec and making changes as required. 


### Prerequisites
- `brew install operator-sdk`
- `go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1` (or `brew install brew install golangci/tap/golangci-lint` if you messed up your `${GOPATH}/bin` 
for some reason
- `make` - in most cases you already have it in `/usr/bin` (use `which make`) but its also available as `brew install make` if not
- Clone this repo under `$GOPATH/src/github.com/run-ai`

### Build targets
* `make` - Grabs dependencies, and runs `build` and `test` 
* `make build` - runs codegen, static checks and compiles
* `make test` - runs codegen, static checks and tests
* `make dependencies` - retrieves project dependencies
* `build-debug` - compiles the binary with debug flags so it can run with `delve`
* `docker-debug` - packages the debug binary into a debug Docker image that runs the binary with `delve`
* `docker` - builds the Docker image
* `docker-push` - pushes the built Docker image

#####A note about the Docker targets
They are mainly for usage by the CI but you can levarage them as well, if `${TAG}` is defined it will be used, if not the image tag is constructed from the current 
commit hash `HEAD` is at

`build-debug` and `docker-debug` can be used with the debug configuration (more below) 

### Build and Deploy
* Apart from metrics which can't be exposed from your local machine to the cluster Prometheus instance, the controller will run with you current `kubectl` 
context if you simply 
build and run it (i.e. via an intellij `go build` target)
* Branches and master commits are built on the CI server, tagged by the current commit hash `HEAD` is at.
* To release a version you need to `git tag` the relevant commit, any tag that begins with `v.` will invoke the release build.
* The release build finds the already-built Docker image that corresponds to the tagged commit and promotes it to the production gcr


### Debugging
There are 2 ways to debug this project to simulate a run in an actual cluster:
* Using Google's `Cloud Code` plugin for Intellij:
    * Uses `skaffold` under the hood
    * All required configuration is under the `debug` dir, it contains the minimal required resources for the controller's deployment (`deployment.yaml`), 
    the debug Dockerfile (which the `make docker-debug` target uses), and a `skaffold.yaml` which the plugin uses to deploy everything onto the cluster you choose.
    * Remember to configure a run/debug target properly with for the plugin
    
* Using `telepresence` - deploy installkit as ususal and swap the controller's deployment so you can run and debug locally
  * Remember that Env file support is required (via the Intellij plugin)
  * Remember to `kubectl scale --replicas=0 deployments/runai-operator` before running `telepresence` or else the `runai-operator` will cause 2 deployments 
  (yours and the one deployed by installkit) to co-exist.

### Build in Local Environment

The following changes are necessary to avoid build issues in local build:

* #### Dockerfile

1) Change: ` FROM registry.access.redhat.com/ubi9/ubi-minimal `
To: ` FROM golang:1.15.1-alpine3.12 `

2) Comment the following by adding '#':
` RUN microdnf install ca-certificates -y && chgrp -R 0 /opt/runai/project-controller && chmod -R g=u /opt/runai/project-controller `

* #### Makefile

1) Change: ` GOBIN=/go/bin `
To: ` GOBIN=$(GOPATH)/bin `

2) Change: ` pre: generate fmt vet golangci `
To: ` pre: generate fmt vet `

