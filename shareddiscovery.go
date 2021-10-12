// SharedDiscovery is a shared library that supports getting data from the
// discovery dynamodb tables based on several criteria, including:
//   - APIToken
//   - Country
//   - Workspace (tablename)
//
// Admin Capabilities
//
// This library also supports the idea of admin calls that can get
// any token needed. This is done using HMAC. For out purposes, the
// private key is stored in Secrets Manager under the name of the
// brand/environment.
package shareddiscovery

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/honeycombio/beeline-go"
)

// QueryInput defines the values used to query dynamo with.
type QueryInput struct {
	AppName     string
	Brand       string
	Country     string
	Environment string
	Signature   string
	QueryString map[string]string
	Workspace   string
}

// SharedDiscoveryIFace describes what is required for building a SharedDiscovery implementation.
type SharedDiscoveryIFace interface {
	GetConfig(ctx context.Context, apiToken string, query QueryInput) (map[string]interface{}, error)
	AdminGetAPIToken(ctx context.Context, secretKey string, query QueryInput) (string, error)
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

// AdminGetAPIToken queries the dynamo table using the provided query
// and returns the correct apiToken for the caller to use to then make
// a request using the GetConfig call.
// It first validates the HMAC signature against the provided secretKey/query params
// to verify the caller is who they say they are.
func (service SharedDiscovery) AdminGetAPIToken(ctx context.Context, secretKey string, query QueryInput) (string, error) {
	_, getAPIKeySpan := beeline.StartSpan(ctx, "adminGetAPIToken")
	defer getAPIKeySpan.Send()

	// validate signature
	if !validateSignature(ctx, query, secretKey) {
		getAPIKeySpan.AddField("error.message", "invalid signature detected")
		return "", errors.New("invalid signature")
	}

	// run query
	items, err := getAPITokenQuery(ctx, service, query)
	if err != nil {
		getAPIKeySpan.AddField("error.message", err.Error())
		return "", err
	}

	// parse token
	return parseAPIToken(ctx, items)

}

func validateSignature(ctx context.Context, query QueryInput, secretKey string) bool {
	_, validateSignatureSpan := beeline.StartSpan(ctx, "validateSignature")
	validateSignatureSpan.AddField("query.object", query)
	// order the query string keys alphabetically
	message := messageFromQuery(query)
	validateSignatureSpan.AddField("query.string", message)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	decoded, err := hex.DecodeString(query.Signature)
	if err != nil {
		validateSignatureSpan.AddField("error.message", fmt.Sprintf("unable to decode signature: %s", err.Error()))
		return false
	}

	return hmac.Equal([]byte(decoded), expectedMAC)
}

func messageFromQuery(query QueryInput) string {
	// order the query string keys alphabetically
	keys := make([]string, len(query.QueryString))

	i := 0
	for k := range query.QueryString {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var message = ""
	for _, val := range keys {
		message = fmt.Sprintf("%s%s", message, query.QueryString[val])
	}
	return message
}

func getAPITokenQuery(ctx context.Context, service SharedDiscovery, query QueryInput) ([]map[string]*dynamodb.AttributeValue, error) {
	_, getAPIKeySpan := beeline.StartSpan(ctx, "getAPITokenQuery")
	defer getAPIKeySpan.Send()
	if query.AppName == "" {
		appResult, err := service.DynamodbSvc.Scan(&dynamodb.ScanInput{
			TableName:        &query.Workspace,
			FilterExpression: aws.String("environment = :e and countryCode = :c and brandName = :b"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":e": {
					S: aws.String(query.Environment),
				},
				":c": {
					S: aws.String(query.Country),
				},
				":b": {
					S: aws.String(query.Brand),
				},
			},
		})
		if err != nil {
			getAPIKeySpan.AddField("error.message", fmt.Sprintf("Unable to get apiToken from discovery v3 admin: %s", err.Error()))
			getAPIKeySpan.AddField("query.values", fmt.Sprintf("%s,%s,%s", query.AppName, query.Country, query.Environment))
			return nil, err
		}
		return appResult.Items, nil
	}

	appResult, err := service.DynamodbSvc.Query(&dynamodb.QueryInput{
		TableName: &query.Workspace,
		IndexName: aws.String("appNameCountryIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"appName": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(query.AppName)},
				},
			},
			"countryCode": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(query.Country)},
				},
			},
		},
		FilterExpression: aws.String("environment = :e"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":e": {
				S: aws.String(query.Environment),
			},
		},
	})
	if err != nil {
		getAPIKeySpan.AddField("error.message", fmt.Sprintf("Unable to getApiToken from discovery v3 admin: %s", err.Error()))
		getAPIKeySpan.AddField("query.values", fmt.Sprintf("%s,%s,%s", query.AppName, query.Country, query.Environment))
		return nil, err
	}
	return appResult.Items, nil
}

func parseAPIToken(ctx context.Context, result []map[string]*dynamodb.AttributeValue) (string, error) {
	_, getQueryAPIKeySpan := beeline.StartSpan(ctx, "parseAPIToken")
	defer getQueryAPIKeySpan.Send()
	var discovery map[string]interface{}

	if len(result) > 0 {
		err := dynamodbattribute.UnmarshalMap(result[0], &discovery)
		if err != nil {
			getQueryAPIKeySpan.AddField("error.message", fmt.Sprintf("Unable to unmarshal results: %s", err.Error()))
			return "", err
		}
		getQueryAPIKeySpan.AddField("success.message", "successfully retrieved ApiToken")
		getQueryAPIKeySpan.Send()
		return fmt.Sprintf("%v", discovery["apiToken"]), nil
	}
	return "", errors.New("No results found")
}
