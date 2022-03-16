# Bookshelf Demo on ECS with AWS CDK v2

A simple microservices application that consists of a frontend and a backend.
The frontend has a single endpoint `/` that prints statistics about the
hostname and date/time. If you add `?services=backend` at the end, the frontend
sends a HTTP request to the backend service, asking for statistics as well.
It will print additional information about the outcome of the HTTP response.

## Development

In a nutshell:

1. `make build` to build. You need Go 1.18+.
2. `make up` to run on Docker.
3. `make down` to stop on Docker.
4. `make build down up` or `make build restart` for a typical dev cycle.

To push the images into ECR repositories (after configuring your AWS account),
run `make push` in the directory of each service.

## Deployment on AWS

First, configure your environment by doing:

1. Install `direnv` from [here](https://direnv.net/)
2. `cp .envrc.template .envrc`
3. Edit the `.envrc` file to your environment
4. Run `direnv allow` to set the environment variables every time your
   enter the directory
5. Double check with `env | sort` that the environments are actually set

Now, Backend and frontend services can be deployed to an AWS ECS cluster with
the code in the `deploy` folder. It utilizes AWS CDK v2 to do so. You can do:

1. `cd deploy`
2. `npm install`
3. `make bootstrap` to bootstrap the CDK. This is only required once per
   AWS account and region.
4. `make build` to synthesize the AWS resources into a CloudFormation template
5. `make deploy` to deploy the CloudFormation template to your AWS account
6. ...
7. `make destroy` to remove the resources from your AWS account

## License

Copyright (c) 2022 Oliver Eilhard. All rights reserved.
