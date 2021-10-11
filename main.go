package shareddiscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/honeycombio/beeline-go"
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

// GetConfig uses the provided `apiToken` to get the correct
// configuration from the specified `tableName`.
func (service SharedDiscovery) GetConfig(ctx context.Context, apiToken string, tableName string) (map[string]interface{}, error) {
	_, configSpan := beeline.StartSpan(ctx, "GetConfig")

	configSpan.AddField("table_name", tableName)

	appResult, err := service.DynamodbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: &tableName,
		Key: map[string]*dynamodb.AttributeValue{
			"apiToken": {
				S: &apiToken,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var discovery map[string]interface{}
	err = dynamodbattribute.UnmarshalMap(appResult.Item, &discovery)
	if err != nil {
		return nil, err
	}

	configSpan.Send()
	return discovery, nil
}
