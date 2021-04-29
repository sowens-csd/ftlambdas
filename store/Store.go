package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sharing"
)

// AccessConfig is the information required to make requests to the play and app store
type AccessConfig struct {
	// This password is from our apple developer account on the Folktells product in-app purchase screen
	// https://appstoreconnect.apple.com/WebObjects/iTunesConnect.woa/ra/ng/app/1489217069/addons
	ConnectPassword string
	// These parameters are from the Google API Console, Folktells Cloud Server credential, GooglePlayAPIServerAccess
	// https://console.developers.google.com/apis/credentials?project=eastern-dream-276521&supportedpurview=project
	ClientSecret string
	RefreshToken string
	AccessToken  string
}

type verifyPurchaseRequest struct {
	VerifyData string `json:"verifyData"`
	StoreType  string `json:"storeType"`
	ProductID  string `json:"productId"`
}

// VerifyStatusType values for the VerifyPurchaseResponse VerifyStatus field
type VerifyStatusType int

// VerifyStatusOk the purchase was verified successfully
const VerifyStatusOk VerifyStatusType = 0

// VerifyStatusInvalidRequest the purchase request was invalid and could not be verified
const VerifyStatusInvalidRequest VerifyStatusType = 1

// VerifyStatusFailed the purchase request failed validation
const VerifyStatusFailed VerifyStatusType = 2

// VerifyStatusError the purchase request could not be validated due to an error
// while we processed the request
const VerifyStatusError VerifyStatusType = 3

// VerifyStatusStoreError the purchase request could not be validated due to an error
// while the play or app store processed the request
const VerifyStatusStoreError VerifyStatusType = 4

// VerifyPurchaseResponse is the result passed back from a verification request.
// If the request was successful then VerifyStatus will be 0, otherwise it will
// be a code that identifies the issue. OnlineUser is only populated if the
// verification succeeded so don't rely on it unless VerifyStatus == 0
type VerifyPurchaseResponse struct {
	VerifyStatus  VerifyStatusType    `json:"verifyStatus"`
	VerifyMessage string              `json:"verifyMessage"`
	OnlineUser    *sharing.OnlineUser `json:"onlineUser"`
}

// VerifyPurchaseFailedError is returned when a purchase request does not pass validation
type VerifyPurchaseFailedError struct {
	error
	Message string
}

const appleStore = "app"
const mockAppleStore = "mock-app"
const androidStore = "play"
const mockAndroidStore = "mock-play"

// UpdateUserSubscription checks the subscription information against the store and
// updates it as required
func UpdateUserSubscription(ftCtx awsproxy.FTContext, onlineUser *sharing.OnlineUser, accessConfig AccessConfig,
	client *http.Client) *sharing.OnlineUser {
	if onlineUser.OriginalTransactionID == "" {
		return onlineUser
	}
	resultingOU := onlineUser
	ftCtx.RequestLogger.Debug().Int64("lastExpiry", onlineUser.LastExpiryVerify).Msg("Before update")
	userSubscription, err := LoadUserSubscriptionByTransaction(ftCtx, onlineUser.OriginalTransactionID)
	if nil != err {
		return resultingOU
	}
	verifyRequest := verifyPurchaseRequest{
		StoreType:  userSubscription.StoreType,
		VerifyData: userSubscription.VerificationData,
		ProductID:  userSubscription.ProductID,
	}
	var verifyResp VerifyPurchaseResponse
	switch verifyRequest.StoreType {
	case appleStore:
		verifyResp = verifyApple(ftCtx, verifyRequest, accessConfig, client)
	case mockAppleStore:
		verifyResp = verifyMockApple(ftCtx, verifyRequest, accessConfig, client)
	case androidStore:
		verifyResp, _ = verifyAndroid(ftCtx, verifyRequest, accessConfig, client)
	}
	if nil != verifyResp.OnlineUser {
		resultingOU = verifyResp.OnlineUser
	}
	ftCtx.RequestLogger.Debug().Int64("lastExpiry", onlineUser.LastExpiryVerify).Msg("After update")
	return resultingOU
}

// ProcessPurchaseRequest takes JSON conforming to verifyPurchaseRequest and verifies it against the app store
// if verified then the user subscription is updated and the user is returned.
func ProcessPurchaseRequest(ftCtx awsproxy.FTContext, verifyRequestJSON string, accessConfig AccessConfig,
	client *http.Client) (VerifyPurchaseResponse, string) {
	updatedToken := ""
	ftCtx.RequestLogger.Debug().Str("purchaseRequest", verifyRequestJSON).Msg("Before processing")
	inputJSON := []byte(verifyRequestJSON)
	var verifyRequest verifyPurchaseRequest
	err := json.Unmarshal(inputJSON, &verifyRequest)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusError, err), updatedToken
	}
	var verifyResp VerifyPurchaseResponse
	switch verifyRequest.StoreType {
	case appleStore:
		verifyResp = verifyApple(ftCtx, verifyRequest, accessConfig, client)
	case mockAppleStore:
		verifyResp = verifyMockApple(ftCtx, verifyRequest, accessConfig, client)
	case androidStore:
		verifyResp, updatedToken = verifyAndroid(ftCtx, verifyRequest, accessConfig, client)
	default:
		return newVerifyPurchaseResponseWithError(VerifyStatusInvalidRequest,
			fmt.Errorf("Unrecognized store %s", verifyRequest.StoreType)), updatedToken

	}
	return verifyResp, updatedToken
}

func verifyApple(ftCtx awsproxy.FTContext, verifyRequest verifyPurchaseRequest, accessConfig AccessConfig,
	client *http.Client) VerifyPurchaseResponse {
	verifyResp, err := VerifyAppleReceipt(ftCtx, verifyRequest, accessConfig, client)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusStoreError, err)
	}
	return handleAppleVerifyResponse(ftCtx, verifyResp, verifyRequest)
}

func verifyMockApple(ftCtx awsproxy.FTContext, verifyRequest verifyPurchaseRequest, accessConfig AccessConfig,
	client *http.Client) VerifyPurchaseResponse {
	ftCtx.RequestLogger.Debug().Msg("Mock Apple request")
	decoded, err := base64.StdEncoding.DecodeString(verifyRequest.VerifyData)
	if err != nil {
		return newVerifyPurchaseResponseWithError(VerifyStatusError, err)
	}
	ftCtx.RequestLogger.Debug().Str("mockReceipt", string(decoded)).Msg("Decoded")
	inputJSON := []byte(decoded)
	var verifyResp VerifyReceiptResponse
	err = json.Unmarshal(inputJSON, &verifyResp)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusError, err)
	}
	ftCtx.RequestLogger.Debug().Int("Mock status", verifyResp.Status).Msg("Unmarshalled")
	return handleAppleVerifyResponse(ftCtx, verifyResp, verifyRequest)
}

func handleAppleVerifyResponse(ftCtx awsproxy.FTContext, verifyResp VerifyReceiptResponse, verifyRequest verifyPurchaseRequest) VerifyPurchaseResponse {
	if !verifyResp.IsVerified() {
		if verifyResp.IsNotVerifiedReceipt() {
			return VerifyPurchaseResponse{VerifyStatus: VerifyStatusFailed}
		} else if verifyResp.IsTemporaryFailure() {
			return VerifyPurchaseResponse{VerifyStatus: VerifyStatusStoreError}
		} else {
			return VerifyPurchaseResponse{VerifyStatus: VerifyStatusFailed}
		}
	}
	ftCtx.RequestLogger.Debug().Msg("Updating subscription")
	ou, err := verifyResp.UpdateSubscriptions(ftCtx, verifyRequest.VerifyData)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusError, err)
	}
	ftCtx.RequestLogger.Debug().Msg("Subscription updated")
	return VerifyPurchaseResponse{VerifyStatus: VerifyStatusOk, OnlineUser: ou}
}

func verifyAndroid(ftCtx awsproxy.FTContext, verifyRequest verifyPurchaseRequest, accessConfig AccessConfig,
	client *http.Client) (VerifyPurchaseResponse, string) {
	subscriptionRequest, updatedToken, err := VerifyAndroidReceipt(ftCtx, verifyRequest, accessConfig, client)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusStoreError, err), updatedToken
	}
	ou, err := subscriptionRequest.UpdateSubscriptions(ftCtx, verifyRequest.ProductID, verifyRequest.VerifyData)
	if nil != err {
		return newVerifyPurchaseResponseWithError(VerifyStatusError, err), updatedToken
	}
	return VerifyPurchaseResponse{VerifyStatus: VerifyStatusOk, OnlineUser: ou}, updatedToken
}

func (e VerifyPurchaseFailedError) Error() string {
	return fmt.Sprintf("Purchase not verified response: %s", e.Message)
}

func newVerifyPurchaseResponseWithError(status VerifyStatusType, err error) VerifyPurchaseResponse {
	return VerifyPurchaseResponse{
		VerifyStatus:  status,
		VerifyMessage: err.Error(),
	}
}
