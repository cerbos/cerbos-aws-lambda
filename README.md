# cerbos-aws-lambda
Gateway service implements AWS Lambda runtime and invokes Cerbos server API hosted in the same AWS Lambda instance.

## Description
This project builds a docker image that can be used to run a Cerbos server in AWS Lambda. The images will contain the gateway executable and the Cerbos binary.
You can use it in a Unix-like system with x86_64 or arm64 architectures. There's also an example of AWS Lambda function based on this image. The function is built using [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html) model.

### Prerequisites

The following tools are required:
- Make - build automation tool
- AWS CLI
- AWS SAM CLI - if you wish to use the provided AWS Lambda function template.
- Docker

### Build the docker image

Run the following command to build the docker image 'cerbos/aws-lambda-gateway' :
```shell
make image
```

### Publish the docker image

In order to publish the image you will need to have a AWS ECR repository. You can create one in the AWS console or using AWS CLI with the following command (replace `<repository-name>` with the name of your repository):

You will see `repositoryUri` in the output of the command. Save it for later use.
```shell
aws ecr create-repository --repository-name <repository-name> --image-scanning-configuration scanOnPush=true
```

Then you will need to get authentication token for the repository. You can do it with the following command:

```shell
export ECR_REPOSITORY_URL=<repositoryUri>
aws ecr get-login-password  | docker login --username AWS --password-stdin $ECR_REPOSITORY_URL
```

Now you can publish the image with the following command:
```shell
make publish-image
```