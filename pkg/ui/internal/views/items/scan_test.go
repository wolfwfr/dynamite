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

func TestScan(t *testing.T) {
	var (
		tableARN             = "table"
		tableName            = "testing-table"
		someIndex            = "some index"
		hkName               = "hk"
		rkName               = "rk"
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
		t.Run("modify scan & scan-parameters keys on entering/exiting scan-mode", func(t *testing.T) {
			skipIf(t, !queryKeyValid, "skipping due to outdated keymap") // skip if keymap needs updating
			sut := newSUT(nil)                                           // init sut
			require.False(t, sut.KeyMap.Scan.Enabled())                  // require key; scan-key disabled
			require.True(t, sut.KeyMap.ScanParameters.Enabled())         // require key; scan-parameters key enabled
			sut.Update(queryKey)                                         // exit scan mode
			assert.True(t, sut.KeyMap.Scan.Enabled())                    // assert key; scan-key enabled
			assert.False(t, sut.KeyMap.ScanParameters.Enabled())         // assert key; scan-parameters key disabled
		})
		t.Run("send a complete scan request with all parameters included", func(t *testing.T) {
			skipIf(t, !queryKeyValid, "skipping due to outdated keymap") // skip if keymap needs updating
			skipIf(t, !scanKeyValid, "skipping due to outdated keymap")  // skip if keymap needs updating
			ctrl := gm.NewController(t)                                  // init mock controller
			db := mocks.NewMockdynamodbClient(ctrl)                      // init mocked DynamoDB client
			sut := newSUT(db)                                            // init sut
			sut.Update(scanKey)                                          // ensure scan-mode is active
			sut.tableIndex.activeIndex = &someIndex                      // set params; active index
			sut.scanParameters.index = &someIndex                        // set params; index
			sut.scanLimit = 10                                           // set params; query limit

			// define expected call to DynamoDB client
			db.EXPECT().
				ScanTable(gm.Any(), gm.Any(), tableName, apitypes.ScanParameters{
					KeyDetails:       AttributeDefinitions,
					IndexName:        &someIndex,
					KeySchema:        gsi[0].KeySchema,
					Limit:            10,
					LastEvaluatedKey: nil,
				}).
				Return(&apitypes.ScanResponse{Items: items}, nil).
				Times(1)

			cmd := sut.PageNext(true)                                             // force system-under-test to prepare the query call
			msgs := extractMessages[messages.PageReady](cmd)                      // conduct async scan call & extract result
			require.Len(t, msgs, 1)                                               // assert response; one result (page)
			assert.EqualValues(t, tableARN, msgs[0].TableARN)                     // assert response; table-arn
			assert.Nil(t, msgs[0].Err)                                            // assert response; error
			assert.EqualValues(t, &someIndex, msgs[0].Index)                      // assert response; index
			assert.EqualValues(t, messages.Page{Items: items}, *msgs[0].Response) // assert response; page
		})
	})
}
