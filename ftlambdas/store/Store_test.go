package store

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
)

const requestTemplate = `
{
	"storeType": "%s",
	"verifyData": "%s",
	"productId": "%s"
}
`

func TestAppleRequestSuccessful(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	request := fmt.Sprintf(requestTemplate, appleStore, receiptBody1, monthlyProductID)
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(folktellsResponse1)),
			Header:     make(http.Header),
		}
	})
	resp, _ := ProcessPurchaseRequest(ftCtx, request, testAccessConfig(), client)
	if nil == resp.OnlineUser {
		t.Error("Should have returned a valid user")
	}
}

func TestSMSReceiptConvert(t *testing.T) {
	fmt.Println("SMS Response1: ")
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(smsResponse1)))
	fmt.Println("\nSMS Response2: ")
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(smsResponse2)))
	fmt.Println("\nInvited User Response: ")
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(sharingInvitedResponse)))
	fmt.Println("\nAuto1 User Response: ")
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(auto1Response)))
	fmt.Println("\nAuto2 User Response: ")
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(auto2Response)))
}

func TestMockAppleRequestSuccessful(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	request := fmt.Sprintf(requestTemplate, mockAppleStore, base64.StdEncoding.EncodeToString([]byte(folktellsResponse1)), monthlyProductID)
	resp, _ := ProcessPurchaseRequest(ftCtx, request, testAccessConfig(), nil)
	if nil == resp.OnlineUser {
		t.Error("Should have returned a valid user")
	}
}

func TestAndroidRequestSuccessful(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	request := fmt.Sprintf(requestTemplate, androidStore, androidServerData1, monthlyProductID)
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(playstoreResponse1)),
			Header:     make(http.Header),
		}
	})
	resp, token := ProcessPurchaseRequest(ftCtx, request, testAccessConfig(), client)
	if len(token) > 0 {
		t.Errorf("Unexpected token %s", token)
	}
	if nil == resp.OnlineUser {
		t.Error("Should have returned a valid user")
	}
}

func testAccessConfig() AccessConfig {
	return AccessConfig{}
}
