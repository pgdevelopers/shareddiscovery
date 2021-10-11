aws_sdk_version="v1.40.59"


mockgen -source=${GOPATH}/pkg/mod/github.com/aws/aws-sdk-go@${aws_sdk_version}/service/dynamodb/dynamodbiface/interface.go -destination=mocks/aws_mockgen/dynamodb/mock_dynamodbiface.go

