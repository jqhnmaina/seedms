# Seed Micro-Service

## Introduction

This is a seed app for creating opinionated micro-services that use Cockroach/PostgresSQL
for storage and either:
1. the `micro api` for load balancing via the proxy handler

    or
2. Google App Engine

The micro-services are created from the template defined in this package.

## Pre-requisites

To generate a template:

1. [Golang](https://golang.org/) installation
2. `GOPATH` set. See https://github.com/golang/go/wiki/SettingGOPATH for details.

To build the template, read the README.MD file of the generated project.

## Installation

```bash
go get -u github.com/tomogoma/seedms
```

## Generating a micro-service

Run the following commands
(This assumes your GOPATH/bin is in your environment variable, otherwise replace
`seedms` with `$GOPATH/bin/seedms` in the commands below):

1. See usage of the seedms command
    ```bash
    seedms -help
    ```
2. Example command to generate template for a micro-service
    name: `test_service`
    description: `A demo service`
    package: `github.com/tomogoma/my_test_service`
    ```bash
    seedms  -desc "A demo service" \
       -dest "github.com/tomogoma/my_test_service" \
       -name "test_service"
    ```