package shareddiscovery

import (
	"context"
	"errors"
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

func TestAdminGetAPIToken_NoAppName_Success(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithoutAppName()
		value        = "value"
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Scan(gomock.Any()).
		Return(&dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{"field1": &dynamodb.AttributeValue{S: &value}},
				{"field2": &dynamodb.AttributeValue{S: &value}},
			},
		}, nil)

	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err != nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == %q, want nil", secretKey, query, err)
	}
}

func TestAdminGetAPIToken_NoAppName_EmptyResult(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithoutAppName()
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Scan(gomock.Any()).
		Return(&dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}, nil)

		// if err doesn't exist fail
	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err == nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == nil, want No Results Found", secretKey, query)
	}
}

func TestAdminGetAPIToken_NoAppName_Error(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithoutAppName()
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Scan(gomock.Any()).
		Return(nil, errors.New("something bad happened"))

		// if err doesn't exist fail
	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err == nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == nil, want something bad happened", secretKey, query)
	}
}

func TestAdminGetAPIToken_WithAppName_Success(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithAppName()
		value        = "value"
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Query(gomock.Any()).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{
				{"field1": &dynamodb.AttributeValue{S: &value}},
				{"field2": &dynamodb.AttributeValue{S: &value}},
			},
		}, nil)

	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err != nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == %q, want nil", secretKey, query, err)
	}
}

func TestAdminGetAPIToken_WithAppName_Empty(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithAppName()
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Query(gomock.Any()).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}, nil)

	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err == nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == %q, want No Results Found", secretKey, query, err)
	}
}

func TestAdminGetAPIToken_WithAppName_Err(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithAppName()
		secretKey    = "secretKey"
	)

	mockDynamoDB.
		EXPECT().
		Query(gomock.Any()).
		Return(nil, errors.New("something bad"))

	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err == nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == nil, want an error", secretKey, query)
	}
}

func TestAdminGetAPIToken_InvalidSignature(t *testing.T) {
	var (
		ctx          = context.TODO()
		mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(gomock.NewController(t))
		self         = New(mockDynamoDB)
		query        = generateQueryWithAppName()
		secretKey    = "badSecret"
	)

	if _, err := self.AdminGetAPIToken(ctx, secretKey, query); err == nil {
		t.Errorf("AdminGetAPIToken(ctx, %q, %q) == nil, want an error", secretKey, query)
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
	shareddiscovery.AdminGetAPIToken(context.Background(), "secretKey", query)
}

// Passing in only the workspace as a query is best used when the
// apiToken is unique for every row. In our case, this is the Firmware
// table
func ExampleSharedDiscovery_GetConfig_withoutCountry() {
	session := session.New()
	shared := New(dynamodb.New(session))
	query := QueryInput{Workspace: "tableName"}
	shared.GetConfig(context.Background(), "apitoken", query)
}

// Passing in a Country will get data from the table filtered by
// the apiToken and Country. This is useful when the apiToken isn't
// unique per row, but an apiToken can have multiple countries.
func ExampleSharedDiscovery_GetConfig_withCountry() {
	session := session.New()
	shared := New(dynamodb.New(session))
	query := QueryInput{Workspace: "tableName", Country: "US"}
	shared.GetConfig(context.Background(), "apitoken", query)
}

func generateQueryWithoutAppName() QueryInput {
	return QueryInput{
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
}

func generateQueryWithAppName() QueryInput {
	return QueryInput{
		Workspace:   "discovery_app",
		AppName:     "sonos",
		Signature:   "545ec3b117f066cf89e18b876ac4534ea900e245c76915003d143488076e1f64",
		Brand:       "oralb",
		Environment: "qa",
		Country:     "US",
		QueryString: map[string]string{
			"brand":       "oralb",
			"environment": "qa",
			"countryCode": "US",
			"appName":     "sonos",
		},
	}
}
