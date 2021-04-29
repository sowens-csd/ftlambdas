package store

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestIsVerifiedTrueForVerifiedResponse(t *testing.T) {
	verifyResponse := VerifyReceiptResponse{Status: 0}
	expectVerifyResponse(t, verifyResponse, true)
}

func TestIsVerifiedFalseForNotVerifiedResponse(t *testing.T) {
	verifyResponse := VerifyReceiptResponse{Status: 21004}
	expectVerifyResponse(t, verifyResponse, false)
}

func TestSingleOriginalID(t *testing.T) {
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, folktellsResponse1)
	ok(t, err)
	originalIDs := verifyResponse.OriginalTransactionIDs()
	if len(originalIDs) != 1 {
		t.Errorf("Expected 1, was %d", len(originalIDs))
	}
}

func TestJustUniqueValuesReturnedForOriginalIDs(t *testing.T) {
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, varyingOriginalIDs)
	ok(t, err)
	originalIDs := verifyResponse.OriginalTransactionIDs()
	if len(originalIDs) != 2 {
		t.Errorf("Expected 2, was %d", len(originalIDs))
	}
}

func TestFindsMostRecent(t *testing.T) {
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, folktellsResponse1)
	ok(t, err)
	productID, expiryMS, autoRenew, graceExpiryMS := verifyResponse.FindMostRecentSubscription(transactionID1)
	if productID != mostRecentProduct1 || expiryMS != mostRecentEpxiry1 || !autoRenew || graceExpiryMS != 0 {
		t.Errorf("Subscription did not match: %s, %d, %d", productID, expiryMS, graceExpiryMS)
	}
}

func TestFindsMostRecentHandlesGracePeriod(t *testing.T) {
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, gracePeriodReceipt)
	ok(t, err)
	productID, expiryMS, autoRenew, graceExpiryMS := verifyResponse.FindMostRecentSubscription(transactionID1)
	if productID != mostRecentProduct1 || expiryMS != mostRecentEpxiry1 || !autoRenew || graceExpiryMS != gracePeriodExpiry1 {
		t.Errorf("Subscription did not match: %s, %d, %d", productID, expiryMS, graceExpiryMS)
	}
}

func TestRealSandboxVerifyCallWithSuccess(t *testing.T) {
	// _, err := VerifyAppleReceipt(receiptBody2, "https://sandbox.itunes.apple.com/verifyReceipt", &http.Client{})
	// if nil != err {
	// 	t.Errorf("Response fail %s", err.Error().Msg())
	// 	return
	// }
}

func TestSuccessfulVerification(t *testing.T) {
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, folktellsResponse1)
	ok(t, err)
	expectVerifyResponse(t, verifyResponse, true)
}

func TestUnVerified(t *testing.T) {
	expectedResponse := sampleResponse + failStatus
	verifyResponse, err := testWithExpectedResponse(http.StatusOK, expectedResponse)
	ok(t, err)
	expectVerifyResponse(t, verifyResponse, false)
}

func TestFailedVerification(t *testing.T) {
	expectedResponse := sampleResponse + failStatus
	_, err := testWithExpectedResponse(http.StatusInternalServerError, expectedResponse)
	if nil == err {
		t.Error("Should have been an error")
	}
}

func TestUpdateNewSubscriptionWorks(t *testing.T) {
	verifyResponse, _ := testWithExpectedResponse(http.StatusOK, newOriginalID)
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	user, err := verifyResponse.UpdateSubscriptions(ftCtx, receiptBody1)
	if nil != err {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if user.ID != userID1 {
		t.Errorf("User not loaded properly %s", user.ID)
	}
}

func testWithExpectedResponse(statusCode int, expectedResponse string) (VerifyReceiptResponse, error) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(expectedResponse)),
			Header:     make(http.Header),
		}
	})
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	verifyRequest := verifyPurchaseRequest{StoreType: appleStore, ProductID: monthlyProductID, VerifyData: receiptBody1}
	accessConfig := AccessConfig{}
	return VerifyAppleReceipt(ftCtx, verifyRequest, accessConfig, client)
}

func expectVerifyResponse(t *testing.T, verifyResponse VerifyReceiptResponse, expected bool) {
	if expected != verifyResponse.IsVerified() {
		if expected {
			t.Error("Expected verified")
		} else {
			t.Error("Expected unverified ")
		}
	}
}

func ok(t *testing.T, err error) {
	if nil != err {
		t.Errorf("Unexpected error %s", err.Error())
	}
}
