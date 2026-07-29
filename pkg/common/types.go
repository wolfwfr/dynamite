package common

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// TODO: move more types that are shared between dynamo-connector &
// dynamo-adapter to common types

// /more information at github.com/aws/aws-sdk-go-v2/service/dynamodb@v1.57.2/types/types.go
type DynamoDBAttributeType int

const (
	DynamoDBAttributeTypeB DynamoDBAttributeType = iota
	DynamoDBAttributeTypeBOOL
	DynamoDBAttributeTypeBS
	DynamoDBAttributeTypeL
	DynamoDBAttributeTypeM
	DynamoDBAttributeTypeN
	DynamoDBAttributeTypeNS
	DynamoDBAttributeTypeNULL
	DynamoDBAttributeTypeS
	DynamoDBAttributeTypeSS
)

func ParseDynamoAttributeType(v types.AttributeValue) DynamoDBAttributeType {
	switch v.(type) {
	case *types.AttributeValueMemberB:
		return DynamoDBAttributeTypeB
	case *types.AttributeValueMemberBOOL:
		return DynamoDBAttributeTypeBOOL
	case *types.AttributeValueMemberBS:
		return DynamoDBAttributeTypeBS
	case *types.AttributeValueMemberL:
		return DynamoDBAttributeTypeL
	case *types.AttributeValueMemberM:
		return DynamoDBAttributeTypeM
	case *types.AttributeValueMemberN:
		return DynamoDBAttributeTypeN
	case *types.AttributeValueMemberNS:
		return DynamoDBAttributeTypeNS
	case *types.AttributeValueMemberNULL: // TODO: ignore?
		return DynamoDBAttributeTypeNULL
	case *types.AttributeValueMemberS:
		return DynamoDBAttributeTypeS
	case *types.AttributeValueMemberSS:
		return DynamoDBAttributeTypeSS
	default:
		return -1
	}
}
