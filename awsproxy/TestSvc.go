package awsproxy

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TestDBDataRecord a single record in the test database
type TestDBDataRecord struct {
	ReferenceID string
	ResourceID  string
	QueryKey    string
	Record      map[string]interface{}
}

// TestDBData the structure of the data in the test database
type TestDBData []TestDBDataRecord

// TestDynamoDB pretends to be dynamo while tracking calls and exposing
// assertions for testing
type TestDynamoDB struct {
	putItemCount    int
	updateItemCount int
	deleteItemCount int
	queryCount      int
	lastPut         map[string]interface{}
	lastUpdate      map[string]interface{}
	lastQuery       map[string]interface{}
	lastDelete      map[string]interface{}
	dbData          TestDBData
}

const resourceIDField = "resourceId"
const referenceIDField = "referenceId"

// NewTestDBSvc creates a new test service with no data
func NewTestDBSvc() *TestDynamoDB {
	return &TestDynamoDB{
		putItemCount:    0,
		queryCount:      0,
		deleteItemCount: 0,
	}
}

// NewTestDBSvcWithData creates a new test service with the given test data
func NewTestDBSvcWithData(data TestDBData) *TestDynamoDB {
	dbData := make([]TestDBDataRecord, 0, len(data))
	for _, record := range data {
		dbData = append(dbData, record)
	}
	return &TestDynamoDB{
		putItemCount: 0,
		queryCount:   0,
		dbData:       dbData,
	}
}

// PutItem mimics the DynamoDB PutItem functionality
func (m *TestDynamoDB) PutItem(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.putItemCount++
	m.lastPut = make(map[string]interface{})
	for key, element := range input.Item {
		val, ok := element.(*types.AttributeValueMemberS)
		if ok {
			m.lastPut[key] = val.Value

		}
	}
	return nil, nil
}

// ExpectPutCount fails the test if the put count doesn't match the expected value
func (m *TestDynamoDB) ExpectPutCount(expectedCount int, t *testing.T) {
	if m.putItemCount != expectedCount {
		t.Errorf("Expected %d, actual %d", expectedCount, m.putItemCount)
	}
}

// ExpectPutItem fails the test if the last put doesn't match the expected values
func (m *TestDynamoDB) ExpectPutItem(expectedValues map[string]interface{}, t *testing.T) {
	if m.putItemCount == 0 {
		t.Errorf("PutItem not called")
	}
	for key, expectedValue := range expectedValues {
		if m.lastPut[key] != expectedValue {
			t.Errorf("Expected put of %s, actual %s, for field %s", m.lastPut[key], expectedValue, key)
		}
	}
}

// Query mimics the DynamoDB Query
func (m *TestDynamoDB) Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	m.queryCount++
	m.lastQuery = make(map[string]interface{})
	var count int32 = 0
	queryKey := ""
	for k, v := range input.ExpressionAttributeValues {
		val, ok := v.(*types.AttributeValueMemberS)
		if ok {
			attrVal := val.Value
			m.lastQuery[k] = attrVal
			if len(queryKey) > 0 {
				queryKey = queryKey + "__" + attrVal
			} else {
				queryKey = queryKey + attrVal
			}
		}
	}
	output := dynamodb.QueryOutput{
		Count: count,
	}
	records := m.findMatchingQueryRecords(queryKey)
	if len(records) > 0 {
		var items []map[string]types.AttributeValue
		for _, dbEntry := range records {
			count++
			item, err := attributevalue.MarshalMap(dbEntry.Record)
			if nil != err {
				return nil, err
			}
			items = append(items, item)
		}
		output = dynamodb.QueryOutput{
			Count: count,
			Items: items,
		}
	}
	return &output, nil
}

// ExpectQuery fails the test if the last query doesn't match the expected values
func (m *TestDynamoDB) ExpectQuery(expectedValues map[string]string) bool {
	if m.queryCount == 0 {
		return false
	}
	for key, expectedValue := range expectedValues {
		if m.lastQuery[key] != expectedValue {
			return false
		}
	}
	return true
}

// GetItem mimics the DynamoDB GetItem
func (m *TestDynamoDB) GetItem(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if input.Key[resourceIDField] == nil || input.Key[referenceIDField] == nil {
		return nil, errors.New("Nil key component")
	}
	resID := input.Key[resourceIDField].(*types.AttributeValueMemberS).Value
	refID := input.Key[referenceIDField].(*types.AttributeValueMemberS).Value
	dbEntries := m.findMatchingRecords(resID, refID)

	output := &dynamodb.GetItemOutput{}

	if len(dbEntries) == 1 {
		dbEntry := dbEntries[0]
		resp, err := attributevalue.MarshalMap(dbEntry.Record)
		if nil != err {
			return nil, err
		}
		output = &dynamodb.GetItemOutput{
			Item: resp,
		}
	} else if len(dbEntries) > 1 {
		return nil, fmt.Errorf("Test setup error, GetItem should never return multiple values")
	}
	return output, nil
}

// DeleteItem mimics the DynamoDB DeleteItem
func (m *TestDynamoDB) DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	m.deleteItemCount++
	m.lastDelete = make(map[string]interface{})
	for key, element := range input.Key {
		m.lastDelete[key] = element.(*types.AttributeValueMemberS).Value
	}

	if input.Key[resourceIDField] == nil || input.Key[referenceIDField] == nil {
		return nil, errors.New("Nil key component")
	}
	// resID := *input.Key[resourceIDField].S
	// refID := *input.Key[referenceIDField].S
	// dbEntries := m.findMatchingRecords(resID, refID)

	output := &dynamodb.DeleteItemOutput{}
	return output, nil
}

// ExpectDeleteItem fails the test if the last DeleteItem call was not for the expected values
func (m *TestDynamoDB) ExpectDeleteItem(expectedValues map[string]interface{}, t *testing.T) {
	if m.deleteItemCount == 0 {
		t.Errorf("DeleteItem not called")
		return
	}
	for key, expectedValue := range expectedValues {
		if m.lastDelete[key] != expectedValue {
			t.Errorf("Expected delete of %s, actual %s, for field %s", expectedValue, m.lastDelete[key], key)
		}
	}
}

// UpdateItem mimics the DynamoDB UpdateItem functionality
func (m *TestDynamoDB) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.updateItemCount++
	m.lastUpdate = make(map[string]interface{})
	for key, element := range input.ExpressionAttributeValues {
		val, ok := element.(*types.AttributeValueMemberS)
		if ok {
			m.lastUpdate[key] = val.Value
		}
	}
	return nil, nil
}

// ExpectUpdateCount fails the test if the update count doesn't match the expected value
func (m *TestDynamoDB) ExpectUpdateCount(expectedCount int, t *testing.T) {
	if m.updateItemCount != expectedCount {
		t.Errorf("Expected %d, actual %d", expectedCount, m.updateItemCount)
	}
}

func (m *TestDynamoDB) findMatchingQueryRecords(queryKey string) TestDBData {
	matching := []TestDBDataRecord{}
	for _, record := range m.dbData {
		if record.QueryKey == queryKey || record.ResourceID == queryKey {
			matching = append(matching, record)
		}
	}
	return matching
}

func (m *TestDynamoDB) findMatchingRecords(resourceID string, referenceID string) TestDBData {
	matching := []TestDBDataRecord{}
	for _, record := range m.dbData {
		if record.ReferenceID == referenceID && record.ResourceID == resourceID {
			matching = append(matching, record)
		}
	}
	return matching
}
