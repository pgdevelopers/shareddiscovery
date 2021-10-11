package shareddiscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// SharedDiscoveryIFace is the interface for building a SharedDiscovery object
type SharedDiscoveryIFace interface {
	GetConfig(ctx context.Context, apiToken string, tableName string) (map[string]interface{}, error)
	GetConfigByCountry(ctx context.Context, apiToken string, countryCode string, tableName string) (map[string]interface{}, error)
}

// SharedDiscovery is a custom service object for interacting with the global config
type SharedDiscovery struct {
	SharedDiscoveryIFace
	DynamodbSvc dynamodbiface.DynamoDBAPI
}

// New is a constructor that takes arguments to build the SharedDiscovery service
func New(dynamodb dynamodbiface.DynamoDBAPI) SharedDiscovery {
	return SharedDiscovery{DynamodbSvc: dynamodb}
}
