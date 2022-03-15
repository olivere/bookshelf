# Bookshelf Demo on ECS with AWS CDK v2

A simple microservices application that consists of a frontend and a backend.
The frontend has a single endpoint `/` that prints statistics about the
hostname and date/time. If you add `?check=true` at the end, the frontend
sends a HTTP request to the backend service, asking for statistics as well.
It will print additional information about the outcome of the HTTP response.

Backend and frontend services can be deployed to an AWS ECS cluster with the
code in the `deploy` folder. It utilizes AWS CDK v2 to do so. You can do:

1. `cd deploy`
2. `npm install`
3. `npm run build`
4. `cdk deploy`
5. ...
6. `cdk destroy`

## License

Copyright (c) 2022 Oliver Eilhard. All rights reserved.
