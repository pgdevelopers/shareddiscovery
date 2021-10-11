package shareddiscovery

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"github.com/pgdevelopers/shareddiscovery/mocks/mock_dynamodbiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SharedDiscoverySuite struct {
	suite.Suite
	self         SharedDiscoveryIFace
	mockDynamoDB *mock_dynamodbiface.MockDynamoDBAPI
}

func (suite *SharedDiscoverySuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())
	suite.mockDynamoDB = mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
	suite.self = New(suite.mockDynamoDB)
}

func (suite *SharedDiscoverySuite) TestGetConfig_Success() {
	ctx := context.TODO()
	tableName := "apps"
	token := "apiToken"
	value := "value"

	suite.
		mockDynamoDB.
		EXPECT().
		GetItem(&dynamodb.GetItemInput{
			TableName: &tableName,
			Key: map[string]*dynamodb.AttributeValue{
				"apiToken": {
					S: &token,
				},
			}}).
		Return(&dynamodb.GetItemOutput{Item: map[string]*dynamodb.AttributeValue{
			"field": &dynamodb.AttributeValue{S: &value},
		}}, nil)

	res, err := suite.self.GetConfig(ctx, token, tableName)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), res["field"], value)
}

func (suite *SharedDiscoverySuite) TestGetConfig_Err() {
	ctx := context.TODO()
	tableName := "apps"
	token := "apiToken"

	suite.
		mockDynamoDB.
		EXPECT().
		GetItem(&dynamodb.GetItemInput{
			TableName: &tableName,
			Key: map[string]*dynamodb.AttributeValue{
				"apiToken": {
					S: &token,
				},
			}}).
		Return(&dynamodb.GetItemOutput{}, errors.New("err"))

	_, err := suite.self.GetConfig(ctx, token, tableName)
	assert.Error(suite.T(), err, "err")
}

func TestSharedDiscvoerySuite(t *testing.T) {
	suite.Run(t, new(SharedDiscoverySuite))
}

//////////////
// EXAMPLES //
//////////////
func ExampleNew() {
	session := session.New()
	New(dynamodb.New(session))
}

func ExampleSharedDiscovery_GetConfig() {
	session := session.New()
	shared := New(dynamodb.New(session))
	shared.GetConfig(context.Background(), "apitoken", "apps")
}
