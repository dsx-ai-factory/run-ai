# run:ai Backend 
[![CircleCI](https://circleci.com/gh/run-ai/backend/tree/master.svg?style=shield&circle-token=9aa8e4f592826e9fbf9b0c5eee8fcf181652295e)](https://circleci.com/gh/run-ai/backend/tree/master)
![Uptime Robot ratio (30 days)](https://img.shields.io/uptimerobot/ratio/m784112917-c7cf2308c28826e2d67077b3)
![Security Headers](https://img.shields.io/security-headers?url=https%3A%2F%2Fapp.run.ai)


## Dev environment
### Prerequisites
The following tools are required (all available with `brew / cask`) : `node`, `kind`, `sops`, `google-cloud-sdk`, `helm` (the charts here are compatible with Helm 3.x)

 `PM2` is used locally  to  have both `ts-node-dev` and `webpack-dev-server` run concurrently for convenience.
 Configuration for `PM2` resides in `dev/pm2.dev.config.js`


 ===
 ### Run on a local k8s cluster.
 * Prerequisite: Kind (`brew install kind`)
 * An npm target `dev:init-local-cluster` exists to create a local cluster fast, use `npm run dev:init-local-cluster -- --help` to see options.
 * For each docker image you want to use in the local cluster (i.e. what  `.Values.deployment.docker` section is pointing to), given that image is available on your docker daemon locally,
 you have to run: `kind load docker-image <image_name>:<image_tag> --name <kind_cluster_name>`. Alternatively the npm target `npm run manual-deploy` takes care of this automatically.
 * `env.example.values.yaml` template file is under `dev/`
 * Use the`namespace.yaml` file (under `dev/`) to create a namespace (since Helm 3 does not create it automatically anymore).

 #### An important note about Kind
 Kind is meant to serve as an ephemeral cluster for testing purposes - currently it does not service container stops (or restarts) of its node containers.

 If you stop the containers running the kind nodes you will have to create a new cluster (`kind delete cluster --name <name>` and use the `dev:init-local-cluster` npm target again).

 ===
### Database:
Prerequisites:
* An up and running database instance (postgres) is perquisite - fastest way to do that is with docker:
`docker run -d -p 5432:5432 -e POSTGRES_DB=runai -e POSTGRES_USER=backend_db -e POSTGRES_PASSWORD=password --name=postgres_11_6  postgres:11.6-alpine`

       Note: When running the docker image locally (not `npm run dev`) with another container for the database you should
       reference the db url with docker host loopback i.e. `postgres://backend_db:password@host.docker.internal:5432/runai`

* A database needs to be manually created based on what you have in your `.env` file (i.e. `CREATE DATABASE runai`) - if you're using the command above this is governed by
the `-e POSTGRES_DB=` part.

#### Schema Migration
Database migrations are performed by the [DB Mechanic](https://github.com/run-ai/db-mechanic) service, see the readme there on how to write and test them.

Once you have created a new migration you MUST update the db migration job's tag to the new one you released in all `values.yaml` files. (see below [Environment](#Environment)) section.

#### Database init for local development
Since migrations depend on an external service, docker must be running for this phase. The database container (see above) must be running as well.
 Connection parameters are taken from the `.env` file.

 Using either one of the `dev:initDb` or `db:init` targets will execute the latest production migration container, if you wish to change that
 set `RUN_AI_BACKEND_LOCAL_DEV_DB_MECHANIC_IMAGE=` in your `.env` file


 ===
### npm targets:

`-- npm run build` - Installs dependencies and build frontend and backend

`-- npm run dev` - starts a `webpack-dev-server` for the frontend and `ts-node-dev` for the backend so both get reloaded automatically on file changes.

`-- npm run dev:initDb` - inits the db (given in the `.env` file) and inserts dummy data prior to running the `dev` target.

#### Having trouble with node-gyp?
try this : https://github.com/nodejs/node-gyp/blob/master/macOS_Catalina.md

(doing the last 'solution step' seems to work best - remove the xcode command line tools, restart and install it again).

 ===
### Environment
All env variables the service expects are passed using an `.env` file - an example of it is committed as `dev/.env.example` - when working locally
copy the `.env.example` file to `.env` and change the values for your environment.  `.env` is ignored by git so you won't commit secrets by mistake.

Production and Staging `values.yaml` files used by Helm to deploy reside under `build/helm/env` and are separated per gcp project.

They are encrypted using `sops`, which enables encryption/decryption of files based on GCP KMS which our google users have permissions for.

In order to modify them you need to:

        gcloud auth application-default login (do this only once)
        sops -d -i <path_to_file>
        ... modify file
        sops -e -i <path_to_file>
        commit changes

All modification must be made via `sops` since it performs checksum validation on the files.

Configuration for sops (how it knows what to encrypt with what key) is stored in `.sops.yaml`

To have sops behave with the GCP KMS you should run (once) `gcloud auth application-default login`

 ===
### Debugging in WebsStorm

Note: `ts-node-dev` isn't working although I kinda tried every possible config - if you manage to get it working update here.
Debugging with `ts-node` means that you don't have an active watch on files and must restart the debug process to see changes.

Add a new `node.js` configuration and enter the following:
```
Node parameters: -r ts-node/register --inspect
Javascript file: js/backend/backendmain.ts
Application parameters: --project tsconfig.json
```
Lastly you have to set the flag `js.debugger.use.node.options` to false in the Webstorm registry (double shift -> actions -> type registry),
if you don't do it the debugger won't stop on breakpoints.

When choosing debug for this configuration WebStorm will attach a debugger on its own when it sees the `--inspect` flag.

Alternatively you can always use the `dev` target and remote-attach a debugger.

For first time use you should use the `dev:initDb` npm target to have a functioning db.

 =
## Build

### Building the Docker image
There are two variants of the Dockerfile:
* `backend.Dockerfile` used in the CI build.
* `backend-dev.Dockerfile` a quicker variant which uses intermediate containers which cache `node_modules` that you can keep and get faster builds

-- See comments in the Dockerfiles as to why and how.

npm targets are available to build both variants: `build:docker-dev` and `build:docker` - where you have to pass a tag and env, i.e. `npm run build:docker staging 0.9.1-test1`

Dockerfiles are located under `build/docker/`.

A third docker file `runai-backend-ci.Dockerfile` exists under the `ci` folder - this docker image (and accompanying scripts) is what we use in our pipeline to build
the backend docker image and and deploy to gke.

===
### Building the Helm chart
We deploy the chart from source code, without building a chart and deploying into a repo.  For now, it seems like this is enough.

===
### CI/CD
Separated to 2 parts:
* Every commit pushed into master triggers the pipeline in circleci and is automatically deployed to `staging.run.ai`
* In order to push changes to production you need to create a tag which conforms to the regex: `prod/<previous_docker_tag>`.
* The npm target `npm run prod:release` does just that. So all you need to do is take the docker tag created by the staging build and use it with this command.
* For instance if the staging build created a docker tag `123-62b64460` you simply need to run `npm run prod:release 123-62b64460`

  There's a job in circleci that picks up on tags conforming to this regex which pulls the docker image tagged like the git tag,
  promotes it to the production environment and pushes it to the production cluster.

  Each such git tag matches exactly:

            * One ci pipeline (that created the original 'staging' docker image and tagged it).
            * One git commit that is the last one appearing in the docker image that was built by the aforementione job.
            * One docker image tagged with the same string, built by the aforementioned job.
            * One docker image, with the same exact tag, that was promoted to the production environment (thus its tag varies in the 'project' part of the image name)
---


### Architecture
![run-ai-architecture](dev/images/Architecture.jpg)

## Running the client against a backend running in production/staging
Sometimes it is more convenient to debug the client against a running backend with its data, to do so:

* Add the localhost to the cross origin in the auth server.
* Go to the client directory (`js/frontend`) and run either one of:
    * > npm run staging:serve
    * > npm run prod:serve

## Creating a New Service

Create a new service from a service temaple using cookiecutter tool, the tool is located under `control-plane/infra/service-cookiecutter`.
Step to run it
1. Install cookiecutter
1. Go to the root path of the repository
1. run `cookiecutter infra/service-cookiecutter` 

for more information read the readme under `control-plane/infra/service-cookiecutter/README.md`
