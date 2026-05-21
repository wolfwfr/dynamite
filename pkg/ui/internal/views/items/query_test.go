package itemselection

import (
	"context"
	"testing"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gm "go.uber.org/mock/gomock"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/mocks"
)

func TestQuery(t *testing.T) {
	var (
		tableARN             = "table"
		tableName            = "testing-table"
		someIndex            = "some index"
		hkName               = "hk"
		rkName               = "rk"
		hkValue              = "hk-value"
		rkValue1             = "rk-value1"
		rkValue2             = "rk-value2"
		items                = genJSONItems(3)
		AttributeDefinitions = []dynamodbtypes.AttributeDefinition{
			{
				AttributeName: &hkName,
				AttributeType: "S",
			},
			{
				AttributeName: &rkName,
				AttributeType: "B",
			},
		}
		gsi = []dynamodbtypes.GlobalSecondaryIndexDescription{
			{
				IndexName: &someIndex,
				KeySchema: []dynamodbtypes.KeySchemaElement{
					{
						AttributeName: &hkName,
						KeyType:       dynamodbtypes.KeyTypeHash,
					},
					{
						AttributeName: &rkName,
						KeyType:       dynamodbtypes.KeyTypeRange,
					},
				},
			},
		}
	)

	// factory initialising a new system-under-test
	newSUT := func(m *mocks.MockdynamodbClient) *ItemSelectionPane {
		sut := newItemSelectionPane(context.Background(), &appconfig.Config{})
		sut.dynamodbClient = m
		sut.config = &appconfig.Config{}
		sut.selectedTable.TableArn = &tableARN
		sut.selectedTable.TableName = &tableName
		sut.selectedTable.AttributeDefinitions = AttributeDefinitions
		sut.selectedTable.GlobalSecondaryIndexes = gsi
		return sut
	}

	t.Run("item-selection-pane should", func(t *testing.T) {
		t.Run("modify query & query-parameters keys on entering/exiting query-mode", func(t *testing.T) {
			skipIf(t, !queryKeyValid, "skipping due to outdated keymap") // skip if keymap needs updating
			sut := newSUT(nil)                                           // init sut
			require.True(t, sut.KeyMap.Query.Enabled())                  // require key; query-key enabled
			require.False(t, sut.KeyMap.QueryParameters.Enabled())       // require key; query-parameters key disabled
			sut.Update(queryKey)                                         // switch to query-mode
			assert.False(t, sut.KeyMap.Query.Enabled())                  // assert key; query-key disabled
			assert.True(t, sut.KeyMap.QueryParameters.Enabled())         // assert key; query-parameters key enabled
		})
		t.Run("send a complete query request with all parameters included", func(t *testing.T) {
			skipIf(t, !queryKeyValid, "skipping due to outdated keymap") // skip if keymap needs updating
			ctrl := gm.NewController(t)                                  // init mock controller
			db := mocks.NewMockdynamodbClient(ctrl)                      // init mocked DynamoDB client
			sut := newSUT(db)                                            // init sut
			sut.Update(queryKey)                                         // switch to query-mode
			sut.tableIndex.activeIndex = &someIndex                      // set params; active index
			sut.queryParameters.index = &someIndex                       // set params; index
			sut.queryParameters.hashKeyValue = hkValue                   // set params; hash-key value
			sut.queryParameters.rangeKeyValue1 = &rkValue1               // set params; range-key value 1
			sut.queryParameters.rangeKeyValue2 = &rkValue2               // set params; range-key value 2
			sut.queryParameters.rangeOrderDescending = true              // set params; range-order
			sut.queryParameters.rangeKeyOperator = messages.Between      // set params; range operator
			sut.queryLimit = 10                                          // set params; query limit

			// define expected call to DynamoDB client
			db.EXPECT().
				QueryTable(gm.Any(), gm.Any(), tableName, apitypes.QueryParameters{
					KeyDetails:       AttributeDefinitions,
					IndexName:        &someIndex,
					KeySchema:        gsi[0].KeySchema,
					HashKeyValue:     hkValue,
					RangeKeyValue1:   &rkValue1,
					RangeKeyValue2:   &rkValue2,
					RangeKeyOperator: apitypes.RangeBetween,
					Limit:            10,
					LastEvaluatedKey: nil,
					Descending:       true,
				}).
				Return(&apitypes.QueryResponse{Items: items}, nil).
				Times(1)

			cmd := sut.PageNext(true)                                             // force system-under-test to prepare the query call
			msgs := extractMessages[messages.PageReady](cmd)                      // conduct async query call & extract result
			require.Len(t, msgs, 1)                                               // assert response; one result (page)
			assert.EqualValues(t, tableARN, msgs[0].TableARN)                     // assert response; table-arn
			assert.Nil(t, msgs[0].Err)                                            // assert response; error
			assert.EqualValues(t, &someIndex, msgs[0].Index)                      // assert response; index
			assert.EqualValues(t, messages.Page{Items: items}, *msgs[0].Response) // assert response; page
		})
	})
}
