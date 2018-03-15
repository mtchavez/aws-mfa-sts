# AWS MFA STS

CLI tool to generate temporary AWS credentials using STS for accounts that
require MFA.

## Installation

`go get github.com/mtchavez/aws-mfa-sts`

## Usage

### Command Flags

```
Usage of aws-mfa-sts:
  -device-arn string
        MFA Serial Number
  -profile string
        Profile name (default "default")
  -region string
        AWS Region (default "us-east-1")
  -token string
        MFA token
```

### Generating a Token

`aws-mfa-sts -token 123456 -device-arn arn:aws:iam::012345678912:mfa/myuser`

Will show the following output if successful:

```
Generating STS Token for default profile
Adding credentials for default-sts to /path/to/.aws/credentials
Use with --profile=default-sts or export AWS_DEFAULT_PROFILE=default-sts
```

### Usage with AWS CLI

Once a new STS profile is generated you can use the `awscli` by passing
the profile in to commands:

`aws --profile=default-sts s3 ls`

Or you can configure your environment `AWS_DEFAULT_PROFILE=default-sts` and
call things as normal `aws s3 ls`

### Using a different profile

If you use the `-profile` command the generated STS profile will be prepended
with the profile name e.g `api-user` will generate `api-user-sts`.
