package ftdb

import (
	"encoding/json"
	"testing"

	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
	"github.com/stretchr/testify/assert"
)

var emptyQuery QueryRequest = QueryRequest{}
var findRecordQuery QueryRequest = QueryRequest{ResourceID: resID1, ReferenceID: refID1}
var userRecord1 FolktellsRecord = FolktellsRecord{ResourceID: resID1, ReferenceID: refID1, ID: userID1}
var userResponse1 QueryResponse = QueryResponse{Results: []FolktellsRecord{userRecord1}}

func TestEmptyQueryProducesNoResults(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	queryJSON, _ := json.Marshal(emptyQuery)
	_, err := QueryFromRequest(ftCtx, string(queryJSON))
	assert.NoError(t, err)
}

func TestFindRecordQueryFindsExpectedRecord(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	queryJSON, _ := json.Marshal(findRecordQuery)
	result, err := QueryFromRequest(ftCtx, string(queryJSON))
	assert.NoError(t, err)
	assert.Equal(t, userResponse1, result)
}
