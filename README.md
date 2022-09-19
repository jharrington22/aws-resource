# aws-resource
![Lint](https://github.com/jharrington22/aws-resource/actions/workflows/golangci-lint.yml/badge.svg)
[![codecov](https://codecov.io/gh/jharrington22/aws-resource/branch/main/graph/badge.svg?token=G8UO8GII3A)](https://codecov.io/gh/jharrington22/aws-resource)

Tool to easily list and delete AWS resources

Warning: All delete commands will delete resources with no warning, use the --dry-run command for any delete to see what resources would be deleted

## General usage

First use the aws-resource tool to list all resources within a given account

```
$ aws-resource list all
```
