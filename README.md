# aws-resource
![Lint](https://github.com/jharrington22/aws-resource/actions/workflows/golangci-lint.yml/badge.svg)
[![codecov](https://codecov.io/gh/jharrington22/aws-resource/branch/main/graph/badge.svg?token=G8UO8GII3A)](https://codecov.io/gh/jharrington22/aws-resource)

Tool to easily list and delete AWS resources

Warning: All delete commands will delete resources with no warning, use the --dry-run flag for any delete command to preview what resources will be deleted

The command supports iterating all regions and only supports the following AWS resources

```
aws-resource list --help
List AWS resources
aws-resource list ec2
aws-resource list elb
aws-resource list elbv2
aws-resource list images
aws-resource list route53
aws-resource list snapshots
aws-resource list volumes

Usage:
  aws-resource list [flags]
  aws-resource list [command]

Available Commands:
  all         List all AWS resources
  ec2         List EC2 instances
  elb         List ELB instances
  elbv2       List ELBv2 instances
  images      List AMIs
  route53     List route53 resources
  snapshots   List EBS snapshots
  volumes     List EBS volumes

Flags:
  -h, --help   help for list

Global Flags:
  -p, --profile string    AWS Profile
  -r, --region string     AWS Region (default "us-east-1")
  -a, --role-arn string   AWS IAM Role ARN
```

## Global flags

Each command supports the following flags;

`--region` list/delete resources in a specific AWS region

`--profile` use a specific AWS profile configure in your local AWS credential configuration

`--role-arn` assume the AWS IAM role before running any operations

## Assuming roles

The `aws-resource` tool supports assuming IAM roles. You can pass the `--role-arn` flag to any command to first assume the role and then run the operation 

```
$ aws sts get-caller-identity
{
    "UserId": "AIDAJHGNVE44PNAXR2KQ2",
    "Account": "123456789101",
    "Arn": "arn:aws:iam::123456789101:user/jaharrin"
}
$ aws-resource whoami --role-arn "arn:aws:iam::234567891011:role/OrganizationAccountAccessRole"
I: AWS Account: 234567891011

```

## General usage

Use the aws-resource tool to list all supported resources in all regions;

```
$ aws-resource list all
I: Listing all resources
I: Listing ec2 instances
I: Found 1 running instances in ap-northeast-3
I: Found 7 running instances in us-east-1
I: Listing elb instances
I: Found 1 running load balancers in us-east-1
I: Found 1 running load balancers in us-east-2
I: Listing elbv2 instances
I: Found 2 running v2 load balancers in us-east-1
I: Found 2 running v2 load balancers in us-east-2
I: Listing route53 hosted zones
I: Found 6 hosted zones
I: Listing ebs snapshots
I: Found 2 snapshots in us-east-1
I: Listing ebs volumes
I: Found 1 volumes in ap-northeast-3
I: 1 attached volumes in ap-northeast-3
I: 0 unattached volumes in ap-northeast-3
I: Found 12 volumes in us-east-1
I: 12 attached volumes in us-east-1
I: 0 unattached volumes in us-east-1
I: Found 27 volumes in us-east-2
I: 11 attached volumes in us-east-2
I: 16 unattached volumes in us-east-2
```

You can list specific resources, some commands have additional flags to output extra information like tags and creation date;

```
aws-resource list ec2 --instance-names --launch-time --instance-type --image-id
I: Listing ec2 instances
I: Found 1 running instances in ap-northeast-3
I: Instance has no tag "Name", 2021-12-07 17:44:09 +0000 UTC, t2.micro, ami-070dd2ec8c4a6df38
I: Found 7 running instances in us-east-1
I: jh-kp6v5-master-2, 2021-09-29 13:01:04 +0000 UTC, m5.2xlarge, ami-093573e55a618974b
I: jh-kp6v5-master-1, 2021-09-29 13:01:01 +0000 UTC, m5.2xlarge, ami-093573e55a618974b
I: jh-kp6v5-master-0, 2021-09-29 13:01:01 +0000 UTC, m5.2xlarge, ami-093573e55a618974b
I: jh-kp6v5-worker-us-east-1a-6jxsz, 2021-09-29 13:10:36 +0000 UTC, m5.xlarge, ami-01ea0772949cb189a
I: jh-kp6v5-worker-us-east-1a-s5xmx, 2021-09-29 13:11:30 +0000 UTC, m5.xlarge, ami-01ea0772949cb189a
I: jh-kp6v5-infra-us-east-1a-xxpdl, 2021-09-29 13:30:32 +0000 UTC, r5.xlarge, ami-093573e55a618974b
I: jh-kp6v5-infra-us-east-1a-7qmqb, 2021-09-29 13:30:29 +0000 UTC, r5.xlarge, ami-093573e55a618974b
```