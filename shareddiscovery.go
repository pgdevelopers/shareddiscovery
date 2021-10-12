// SharedDiscovery is a shared library that supports getting data from the
// discovery dynamodb tables based on several criteria, including:
//   - APIToken
//   - Country
//   - Workspace (tablename)
//
// Admin Capabilities
//
// This library also supports the idea of admin calls that can get
// any token needed. This is done using HMAC and the private key is
// stored in Secrets Manager.
package shareddiscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/honeycombio/beeline-go"
)

// QueryInput defines the values used to query dynamo with.
type QueryInput struct {
	Environment string
	AppName     string
	Country     string
	Brand       string
	Workspace   string
	QueryString map[string]string
}

// SharedDiscoveryIFace describes what is required for building a SharedDiscovery implementation.
type SharedDiscoveryIFace interface {
	GetConfig(ctx context.Context, apiToken string, query QueryInput) (map[string]interface{}, error)
}

// SharedDiscovery is a custom service object for interacting with the global config
type SharedDiscovery struct {
	SharedDiscoveryIFace
	DynamodbSvc dynamodbiface.DynamoDBAPI
}

// New is a constructor that takes a preconfigured dynamodbiface and returns an implementation of SharedDiscoveryIFace
// Use this in your init function after creating your aws session and initializing dynamo.
func New(dynamodb dynamodbiface.DynamoDBAPI) SharedDiscovery {
	return SharedDiscovery{DynamodbSvc: dynamodb}
}

// GetConfig uses the provided `APIToken` to get the correct
// configuration from the specified `tableName`.
func (service SharedDiscovery) GetConfig(ctx context.Context, apiToken string, query QueryInput) (map[string]interface{}, error) {
	_, configSpan := beeline.StartSpan(ctx, "GetConfig")
	configSpan.AddField("table_name", query.Workspace)

	var err error
	var appResult *dynamodb.GetItemOutput

	// dynamically build attribute values
	searchAttributes := map[string]*dynamodb.AttributeValue{
		"apiToken": {
			S: &apiToken,
		},
	}
	searchAttributes = addNeededSearchAttributes(searchAttributes, query)

	appResult, err = service.DynamodbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: &query.Workspace,
		Key:       searchAttributes,
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

func addNeededSearchAttributes(searchAttributes map[string]*dynamodb.AttributeValue, query QueryInput) map[string]*dynamodb.AttributeValue {
	if query.Country != "" {
		searchAttributes["countryCode"] = &dynamodb.AttributeValue{S: &query.Country}
	}

	return searchAttributes
}
