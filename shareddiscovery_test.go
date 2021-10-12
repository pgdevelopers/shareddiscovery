package shareddiscovery

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"github.com/pgdevelopers/shareddiscovery/mocks/mock_dynamodbiface"
)

func TestGetConfig_Success(t *testing.T) {
	var (
		ctx          = context.TODO()
		ctrl         = gomock.NewController(t)
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
		self         = New(mockDynamoDB)
		token        = "apiToken"
		value        = "value"
		query        = QueryInput{Workspace: "apps"}
	)

	mockDynamoDB.
		EXPECT().
		GetItem(&dynamodb.GetItemInput{
			TableName: &query.Workspace,
			Key: map[string]*dynamodb.AttributeValue{
				"apiToken": {
					S: &token,
				},
			}}).
		Return(&dynamodb.GetItemOutput{Item: map[string]*dynamodb.AttributeValue{
			"field": &dynamodb.AttributeValue{S: &value},
		}}, nil)

	if _, err := self.GetConfig(ctx, token, query); err != nil {
		t.Errorf("GetConfig(ctc, %q, %q) == %q, want nil", token, query.Workspace, err)
	}

}

func TestGetConfig_Err(t *testing.T) {
	var (
		ctx          = context.TODO()
		ctrl         = gomock.NewController(t)
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
		self         = New(mockDynamoDB)
		query        = QueryInput{Workspace: "apps"}
		token        = "apiToken"
		want         = errors.New("error recieved")
	)

	mockDynamoDB.
		EXPECT().
		GetItem(&dynamodb.GetItemInput{
			TableName: &query.Workspace,
			Key: map[string]*dynamodb.AttributeValue{
				"apiToken": {
					S: &token,
				},
			}}).
		Return(&dynamodb.GetItemOutput{}, want)

	if _, err := self.GetConfig(ctx, token, query); err == nil {
		t.Errorf("GetConfig(ctx, %q, %q) == nil, want %q", token, query.Workspace, want)
	}
}

func TestGetConfig_WithCountry(t *testing.T) {
	var (
		ctx          = context.TODO()
		ctrl         = gomock.NewController(t)
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
		self         = New(mockDynamoDB)
		token        = "apiToken"
		value        = "value"
		query        = QueryInput{Workspace: "apps", Country: "US"}
	)

	mockDynamoDB.
		EXPECT().
		GetItem(&dynamodb.GetItemInput{
			TableName: &query.Workspace,
			Key: map[string]*dynamodb.AttributeValue{
				"apiToken": {
					S: &token,
				},
				"countryCode": {
					S: &query.Country,
				},
			}}).
		Return(&dynamodb.GetItemOutput{Item: map[string]*dynamodb.AttributeValue{
			"field": &dynamodb.AttributeValue{S: &value},
		}}, nil)

	if _, err := self.GetConfig(ctx, token, query); err != nil {
		t.Errorf("GetConfig(ctc, %q, %q) == %q, want nil", token, query.Workspace, err)
	}
}

////////////////////////////////////////////////////////////////
// EXAMPLES
///////////////////////////////////////////////////////////////

// Typical usage is to setup a variable with the interface type
// and initialize that variable in your modules init function using
// the New() function provided
func Example() {
	var shareddiscovery SharedDiscoveryIFace

	// setup AWS Session
	session := session.New()

	// setup DynamoDB
	dynamo := dynamodb.New(session)

	// define a QueryInput
	query := QueryInput{Workspace: "tableName"}

	// setup shareddiscovery now
	shareddiscovery = New(dynamo)

	// call functions
	shareddiscovery.GetConfig(context.Background(), "someApiToken", query)
}

// For making an admin call, a secretKey is required
// and the query string is needed to regenerate the
// signature correctly. The callers signature should also
// be passed via in to verify.
func ExampleSharedDiscovery_AdminGetAPIToken() {
	var shareddiscovery SharedDiscoveryIFace

	// setup AWS Session
	session := session.New()

	// setup DynamoDB
	dynamo := dynamodb.New(session)

	// define a QueryInput
	query := QueryInput{
		Workspace:   "discovery_app",
		Signature:   "e52af791ec66085081f993a42d2a02a4b1dc08ad7b9f030dac5a0c20d5a0c68c",
		Brand:       "oralb",
		Environment: "qa",
		Country:     "US",
		QueryString: map[string]string{
			"brand":       "oralb",
			"environment": "qa",
			"countryCode": "US",
		},
	}

	// setup shareddiscovery now
	shareddiscovery = New(dynamo)

	// call functions
	key, _ := shareddiscovery.AdminGetAPIToken(context.Background(), "secretKey", query)
	fmt.Println(key)
	// OUTPUT:
	// ZWk3QFdBWVVTVjFFbFE0dlBPQDhhSURabmtmN0dEN1k=
}

// Passing in only the workspace as a query is best used when the
// apiToken is unique for every row. In our case, this is the Firmware
// table
func ExampleSharedDiscovery_GetConfig_1() {
	session := session.New()
	shared := New(dynamodb.New(session))
	query := QueryInput{Workspace: "tableName"}
	shared.GetConfig(context.Background(), "apitoken", query)
}

// Passing in a Country will get data from the table filtered by
// the apiToken and Country. This is useful when the apiToken isn't
// unique per row, but an apiToken can have multiple countries.
func ExampleSharedDiscovery_GetConfig_2() {
	session := session.New()
	shared := New(dynamodb.New(session))
	query := QueryInput{Workspace: "tableName", Country: "US"}
	shared.GetConfig(context.Background(), "apitoken", query)
}
