Cerbos AWS Lambda Docker Image
==============================
Gateway service implements AWS Lambda runtime and invokes Cerbos server API hosted in the same AWS Lambda instance.

Cerbos is the open core, language-agnostic, scalable authorization solution that makes user permissions and authorization simple to implement and manage by writing context-aware access control policies for your application resources.

* [Cerbos website](https://cerbos.dev)
* [Cerbos documentation](https://docs.cerbos.dev)
* [Cerbos GitHub repository](https://github.com/cerbos/cerbos)
* [Cerbos Slack community](http://go.cerbos.io/slack)

## Description
This project builds a docker image that can be used to run a Cerbos server in AWS Lambda. The images will contain the gateway executable and the Cerbos binary.

The following commands assume you run a Unix-like system with x86_64 or arm64 architectures.

There's also an example of AWS Lambda function based on this image. The function is built using [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html) model.

### Prerequisites

The following tools are required:
- Make - build automation tool
- AWS CLI
- AWS SAM CLI - if you wish to use the provided AWS Lambda function template
- Docker

### Build the Docker image

Check out `conf.default.yml` for Cerbos configuration. The default configuration uses blob storage, e.g. AWS S3 bucket. Cerbos config can read from environment variables. If you choose to do so, your AWS Lambda has to expose them. 

By default, the latest release of Cerbos is used. If you want to use a particular Cerbos version, you can specify it in `CERBOS_RELEASE` environment variable.

Run the following command to build the docker image 'cerbos/aws-lambda-gateway':
```shell
make image
```

Note that the image will be built in whatever architecture you are running on (x86 or arm64) - the AWS region you use must support the architecture you are deploying too also else you will get an exec format error when it tries to start up the lambda.

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
You can create an AWS Lambda function referencing the published image with any tool. Alternatively, you can use the provided template `sam.yml`. For the latter option, please visit the template and replace `<repositoryUri>` with the value you saved in the previous step. The template exposes these environment variables:
- BUCKET_URL - the URL of the S3 bucket where Cerbos policies are stored.
- BUCKET_PREFIX - optional prefix for the S3 bucket.
- CERBOS_LOGGING_LEVEL - Cerbos logging level. It defaults to INFO.

You will need to grant the role access to the S3 bucket you are storing policies in.

To publish the function, run the following command:
```shell
make publish-lambda
```

The command will create an AWS Lambda function as part of the stack called as per `CERBOS_STACK_NAME` environment variable (if unset, defaults to `Cerbos`). The stack will also create API Gateway resources and an IAM role for the function. **Ensure the role has the necessary permissions to access the S3 bucket (or other policy storage you might use)**.

Should you change the configuration and rebuild the image, you can update the Lambda via:

```shell
make clean
make image
make publish-image
make update-lambda
```