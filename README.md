<p align="center"><img src="lightyear-logo.png"></img></p>

# Shared Discovery

![Test](https://github.com/pgdevelopers/shareddiscovery/actions/workflows/gotest.yml/badge.svg)

A library for shared code in the Discovery Service.

[API Documentation](https://pgdevelopers.github.io/shareddiscovery)

## Contributing
To contribute, you will need:
  * `go` >= 1.17
  * [gopages](https://johnstarich.com/go/gopages/pkg/github.com/johnstarich/go/gopages/) for generating documentation

Create a feature branch from `main` and make your changes updates. Be sure to follow [documentation standards](https://go.dev/blog/godoc) when writing public functions and types and provide [examples](https://pkg.go.dev/testing#hdr-Examples) where applicable.

Make sure to run `make docgen` to update the documentation files.

### Make tasks
* `make test`    runs unit tests
* `make mockgen` builds mocks needed for testing

Once you feel good about your change and get it merged to main, you'll want to create a new release. Please follow [semantic versioning](https://semver.org/) as you release this library.

## Installation
Since this is a private repository, you'll need a new environment variable called [GOPRIVATE](https://www.goproxy.io/docs/GOPRIVATE-env.html) which you'll need to set to this repo. Put this in your .bashrc or .zshrc file:
```bash
export GOPRIVATE="github.com/pgdevelopers"
```
Now you should be able to run `go get github.com/pgdevelopers/shareddiscovery` to install this package in your project.

## Usage

See [the docs](https://pgdevelopers.github.io/shareddiscovery) for complete API documentation.

### Basic usage
```go

  import (
    "github.com/pgdevelopers/shareddiscovery"
    "github.com/aws/aws-sdk-go/aws/session"
	  "github.com/aws/aws-sdk-go/service/dynamodb"
  )

  var discovery SharedDiscoveryIFace

  func main() {
    query := QueryInput{Workspace: "tableName"}
    discovery.GetConfig(context.Background(), "someApiToken", query)
  }

  func init() {
    // setup AWS Session
    session := session.New()

    // setup DynamoDB
    dynamo := dynamodb.New(session)

    // setup shareddiscovery now
    discovery = shareddiscovery.New(dynamo)
  }
```

### Testing 

Provided is an interface that can be used with [gomock](https://github.com/golang/mock) to generate a mock for testing with. See [the discovery service](https://github.com/pgdevelopers/discovery/blob/qa/src/functions/discoveryConfig/main_test.go#L43) for more examples of testing with this library.
