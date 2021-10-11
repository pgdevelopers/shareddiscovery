package shareddiscovery

import (
	"testing"

	"github.com/pgdevelopers/shareddiscovery/aws_mockgen/dynamodb/mock_dynamodbiface/mock_dynamodbiface"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	assert.Equal(t, "hello", "hello")
}

func ExampleNew_F(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDynamoDB := mock_dynamodbiface.NewMockDynamoDBAPI(ctrl)
	New(mockDynamoDB)
}
