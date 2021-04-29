package store

import (
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

func TestLoadsMatchingSubscription(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	subscription, err := LoadUserSubscriptionByTransaction(ftCtx, transactionID1)
	if nil != err {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if subscription.OriginalTransactionID != transactionID1 {
		t.Errorf("Wrong transaction %s", subscription.OriginalTransactionID)
	}
}

func TestLoadFailsWithNoMatchingTransaction(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	_, err := LoadUserSubscriptionByTransaction(ftCtx, missingTransactionID)
	if nil == err {
		t.Error("Should have return UserSubscriptionNotFound error")
	}
	switch err.(type) {
	case *UserSubscriptionNotFoundError:
		// this is expected
	default:
		t.Errorf("Unexpected error %s", err.Error())
	}
}

func TestSaveWorksForNewSubscription(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	err := newUserSubscription1.Save(ftCtx)
	ok(t, err)
	testSvc.ExpectPutCount(1, t)
}

func TestSaveFailsForSubscriptionWithNonMatchingUser(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID2, testSvc)
	err := differentUserSubscription1.Save(ftCtx)
	if nil == err {
		t.Error("Should have return UserSubscriptionMismatchError error")
	}
	switch err.(type) {
	case *UserSubscriptionMismatchError:
		// this is expected
	default:
		t.Errorf("Unexpected error %s", err.Error())
	}
}
