# cerbos-aws-lambda
Gateway service implements AWS Lambda runtime and invokes Cerbos server API hosted in the same AWS Lambda instance.

## Description
This project builds a docker image that can be used to run a Cerbos server in AWS Lambda. The images will contain the gateway executable and the Cerbos binary.
You can use it in a Unix-like system with x86_64 or arm64 architectures. There's also an example of AWS Lambda function based on this image. The function is built using [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html) model.

### Prerequisites

The following tools are required:
- Make - build automation tool
- AWS CLI
- AWS SAM CLI - if you wish to use the provided AWS Lambda function template
- Docker

### Build the Docker image

Check out `conf.default.yml` for Cerbos configuration. The default configuration uses blob storage, e.g. AWS S3 bucket. Cerbos config can read from environment variables. If you choose to do so, your AWS Lambda has to expose them.

Run the following command to build the docker image 'cerbos/aws-lambda-gateway':
```shell
make image
```

### Publish the Docker image

To publish the image, you will need to have an AWS ECR repository. You can create one in the AWS console or using AWS CLI with the following command (replace `<repository-name>` with the name of your repository):

You will see `repositoryUri` in the output of the command. Save it for later use.
```shell
aws ecr create-repository --repository-name <repository-name> --image-scanning-configuration scanOnPush=true
```

Then you will need to get an authentication token for the repository. You can do it with the following command:

```shell
export ECR_REPOSITORY_URL=<repositoryUri>
aws ecr get-login-password  | docker login --username AWS --password-stdin $ECR_REPOSITORY_URL
```

Now you can publish the image with the following command:
```shell
make publish-image
```

### Create AWS Lambda function
You can create an AWS Lambda function referencing the published image with any tool. Alternatively, you can use the provided template `sam.yml`. For the latter option, please replace `<repositoryUri>` with the value you saved in the previous step. The template exposes these environment variables:
- BUCKET_URL - the URL of the S3 bucket where Cerbos policies are stored.
- BUCKET_PREFIX - optional prefix for the S3 bucket.
- CERBOS_LOGGING_LEVEL - Cerbos logging level. It defaults to INFO.

To publish the function, run the following command:
```shell
make publish-lambda
```

The command will create an AWS Lambda function as part of the stack called `$CERBOS_STACK_NAME` (if unset, defaults to `Cerbos`). The stack will also create API Gateway resources and an IAM role for the function. Ensure the role has the necessary permissions to access the S3 bucket (or other policy storage you might use).