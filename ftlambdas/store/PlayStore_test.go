package store

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
)

const verificationData1 = "kflajlkfjdsoiaruewoqu84093"
const monthlySubscriptionProduct = "052020monthlysharingplan"
const yearlySubscriptionProduct = "052020yearlysharingplan"

func TestPlaystoreVerifyCallWithSuccess(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	accessConfig := AccessConfig{}
	verifyRequest := verifyPurchaseRequest{StoreType: androidStore, VerifyData: verificationData1, ProductID: monthlySubscriptionProduct}
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(playstoreResponse1)),
			Header:     make(http.Header),
		}
	})
	_, _, err := VerifyAndroidReceipt(ftCtx, verifyRequest, accessConfig, client)
	if nil != err {
		t.Errorf("Response fail %s", err.Error())
		return
	}
}

func TestUpdateAccessTokenSuccess(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	accessConfig := AccessConfig{}
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(tokenResponse1)),
			Header:     make(http.Header),
		}
	})
	token, err := updateAccessToken(ftCtx, accessConfig, client)
	if nil != err {
		t.Errorf("Response fail %s", err.Error())
		return
	}
	if token != accessToken1 {
		t.Errorf("Wrong token, expected %s, was %s", accessToken1, token)
	}
}
