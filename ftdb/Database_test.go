package ftdb

import (
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

type testRecord struct {
	ResourceID  string `json:"resourceId"`
	ReferenceID string `json:"referenceId"`
	ID          string `json:"id"`
}

func TestGetItemSucceedsWithExistingItem(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	result := testRecord{}
	found, err := GetItem(ftCtx, resID1, refID1, &result)
	if nil != err {
		t.Fatalf("Error returned %s", err.Error())
	}
	if !found {
		t.Fatal("Not found")
	}
	if result.ID != userID1 {
		t.Fatalf("Result had %s instead of %s", result.ID, userID1)
	}
}

func TestGetItemFailsWithMissingItem(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	result := testRecord{}
	found, err := GetItem(ftCtx, missingResID1, refID1, &result)
	if nil != err {
		t.Fatalf("Error returned %s", err.Error())
	}
	if found {
		t.Fatal("Found item not in DB")
	}
}

func TestPutItemSucceeds(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	result := testRecord{ID: userID2}
	err := PutItem(ftCtx, resID2, refID2, result)
	if nil != err {
		t.Fatalf("Error returned %s", err.Error())
	}
}

func TestDeleteItemSucceedsWithExistingItem(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	err := DeleteItem(ftCtx, resID1, refID1)
	if nil != err {
		t.Fatalf("Error returned %s", err.Error())
	}
}
