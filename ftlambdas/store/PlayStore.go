package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sharing"
)

// see https://developers.google.com/android-publisher/api-ref/purchases/subscriptions/get for information

// IntroductoryPrice information about a single introductory price
type IntroductoryPrice struct {
	IntroductoryPriceCurrencyCode string `json:"introductoryPriceCurrencyCode"`
	IntroductoryPriceAmountMicros int    `json:"introductoryPriceAmountMicros,string"`
	IntroductoryPricePeriod       string `json:"introductoryPricePeriod"`
	IntroductoryPriceCycles       int    `json:"introductoryPriceCycles"`
}

// PurchaseSubscriptionsResponse response to a request for information about a single subscription
type PurchaseSubscriptionsResponse struct {
	Kind                       string            `json:"kind"`
	StartTimeMillis            int               `json:"startTimeMillis,string"`
	ExpiryTimeMillis           int               `json:"expiryTimeMillis,string"`
	AutoResumeTimeMillis       int               `json:"autoResumeTimeMillis,string"`
	AutoRenewing               bool              `json:"autoRenewing"`
	PriceCurrencyCode          string            `json:"priceCurrencyCode"`
	PriceAmountMicros          int               `json:"priceAmountMicros,string"`
	IntroductoryPriceInfo      IntroductoryPrice `json:"introductoryPriceInfo"`
	CountryCode                string            `json:"countryCode"`
	DeveloperPayload           string            `json:"developerPayload"`
	PaymentState               int               `json:"paymentState"`
	CancelReason               int               `json:"cancelReason"`
	UserCancellationTimeMillis int               `json:"userCancellationTimeMillis,string"`
	OrderID                    string            `json:"orderId"`
	LinkedPurchaseToken        string            `json:"linkedPurchaseToken"`
	PurchaseType               int               `json:"purchaseType"`
	ProfileName                string            `json:"profileName"`
	EmailAddress               string            `json:"emailAddress"`
	GivenName                  string            `json:"givenName"`
	FamilyName                 string            `json:"familyName"`
	ProfileID                  string            `json:"profileId"`
	AcknowledgementState       int               `json:"acknowledgementState"`
	PromotionType              int               `json:"promotionType"`
	PromotionCode              string            `json:"promotionCode"`
}

// TokenResponse the JSON returned from the Google auth token refresh endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

/*
{
	"kind": "androidpublisher#subscriptionPurchase",
	"startTimeMillis": long,
	"expiryTimeMillis": long,
	"autoResumeTimeMillis": long,
	"autoRenewing": boolean,
	"priceCurrencyCode": string,
	"priceAmountMicros": long,
	"introductoryPriceInfo": {
	  "introductoryPriceCurrencyCode": string,
	  "introductoryPriceAmountMicros": long,
	  "introductoryPricePeriod": string,
	  "introductoryPriceCycles": integer
	},
	"countryCode": string,
	"developerPayload": string,
	"paymentState": integer,
	"cancelReason": integer,
	"userCancellationTimeMillis": long,
	"cancelSurveyResult": {
	  "cancelSurveyReason": integer,
	  "userInputCancelReason": string
	},
	"orderId": string,
	"linkedPurchaseToken": string,
	"purchaseType": integer,
	"priceChange": {
	  "newPrice": {
		"priceMicros": string,
		"currency": string
	  },
	  "state": integer
	},
	"profileName": string,
	"emailAddress": string,
	"givenName": string,
	"familyName": string,
	"profileId": string,
	"acknowledgementState": integer,
	"promotionType": integer,
	"promotionCode": string
  }*/

const folktellsPackage = "com.csdcorp.app.folktells.android"
const playstoreEndpoint = "https://www.googleapis.com/androidpublisher/v3/applications/"
const tokenRefreshEndpoint = "https://accounts.google.com/o/oauth2/token"
const subscriptionsPath = "/purchases/subscriptions/"
const tokensPathParameter = "/tokens/"
const authorizationHeader = "Authorization"

const clientID = "987669758110-02ivcsm9t2e5d7e74ic290b6pg51i8en.apps.googleusercontent.com"

// VerifyAndroidReceipt verifies a receipt with the play store
func VerifyAndroidReceipt(ctx awsproxy.FTContext, verifyRequest verifyPurchaseRequest, accessConfig AccessConfig, client *http.Client) (PurchaseSubscriptionsResponse, string, error) {
	updatedToken := ""
	authToken := accessConfig.AccessToken
	emptyResponse := PurchaseSubscriptionsResponse{}
	tryWithUpdate := true
	var resp *http.Response
	if len(authToken) != 0 {
		tryWithUpdate = false
		resp, err := makePlaystoreRequest(verifyRequest, authToken, client)
		if nil != err {
			return emptyResponse, updatedToken, err
		}
		tryWithUpdate = http.StatusUnauthorized == resp.StatusCode
	}
	if tryWithUpdate {
		authToken, err := updateAccessToken(ctx, accessConfig, client)
		if nil != err {
			return emptyResponse, updatedToken, err
		}
		updatedToken = authToken
		resp, err = makePlaystoreRequest(verifyRequest, authToken, client)
		if nil != err {
			return emptyResponse, updatedToken, err
		}
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("HTTP Status on verify call %d", resp.StatusCode))
	if http.StatusOK != resp.StatusCode {
		if http.StatusUnauthorized == resp.StatusCode {
			ctx.RequestLogger.Error().Msg("Unauthorized even after a refresh token, should not happen")
		}
		// printBody(ctx, resp)
		return emptyResponse, updatedToken, fmt.Errorf("Non 200 status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return emptyResponse, updatedToken, err
	}
	stringBody := string(respBody)
	// fmt.Printf(stringBody)
	inputJSON := []byte(string(stringBody))
	var subscriptionResponse PurchaseSubscriptionsResponse
	err = json.Unmarshal(inputJSON, &subscriptionResponse)
	if nil != err {
		return emptyResponse, updatedToken, err
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("Got a response: %s", subscriptionResponse.OrderID))
	return subscriptionResponse, updatedToken, nil
}

func makePlaystoreRequest(verifyRequest verifyPurchaseRequest, accessToken string, client *http.Client) (*http.Response, error) {
	subscriptionURL := fmt.Sprintf("%s%s%s%s%s%s", playstoreEndpoint, folktellsPackage,
		subscriptionsPath, verifyRequest.ProductID, tokensPathParameter, url.PathEscape(verifyRequest.VerifyData))
	req, err := http.NewRequest(http.MethodGet, subscriptionURL, nil)
	if nil != err {
		return nil, err
	}
	req.Header.Add(authorizationHeader, fmt.Sprintf("Bearer %s", accessToken))
	return client.Do(req)
}

func printBody(ctx awsproxy.FTContext, resp *http.Response) {
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return
	}
	stringBody := string(respBody)
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("Error details %s", stringBody))
}

func updateAccessToken(ctx awsproxy.FTContext, accessConfig AccessConfig, client *http.Client) (string, error) {
	refreshForm := url.Values{}
	refreshForm.Set("grant_type", "refresh_token")
	refreshForm.Set("refresh_token", accessConfig.RefreshToken)
	refreshForm.Set("client_id", clientID)
	refreshForm.Set("client_secret", accessConfig.ClientSecret)
	ctx.RequestLogger.Info().Msg("About to request a new token")
	resp, err := client.PostForm(tokenRefreshEndpoint, refreshForm)
	if nil != err {
		return "", err
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("HTTP Status on token call %d", resp.StatusCode))
	if http.StatusOK != resp.StatusCode {
		return "", fmt.Errorf("Non 200 status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return "", err
	}
	stringBody := string(respBody)
	// fmt.Printf(stringBody)
	inputJSON := []byte(string(stringBody))
	var tokenResponse TokenResponse
	err = json.Unmarshal(inputJSON, &tokenResponse)
	if nil != err {
		return "", err
	}
	return tokenResponse.AccessToken, nil

}

// UpdateSubscriptions updates the DB based on the subscription information in the receipt
//
// There are two separate cases. The first is a new subscription. In that case the subscription
// will be created for the current user and their user record will be updated to reflect the state
// the subscription. The second case is an existing subscription entry for the original transaction
// ID. In that case the user that is already associated with that transaction ID must match the
// the current user. If they do not match it is an error and the method logs and exits. If they do
// match the subscription and user records are updated to ensure the expiry date is correct.
func (subscriptionResponse *PurchaseSubscriptionsResponse) UpdateSubscriptions(ctx awsproxy.FTContext,
	productID string, verifyData string) (*sharing.OnlineUser, error) {
	ctx.RequestLogger.Info().Msg("Loading user")
	user, err := sharing.LoadOnlineUser(ctx, ctx.UserID)
	if nil != err {
		return nil, err
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("Updating subscription: %s, expiry: %d", productID, subscriptionResponse.ExpiryTimeMillis))
	userSubscription := UserSubscription{
		UserID:                ctx.UserID,
		TransactionID:         subscriptionResponse.OrderID,
		OriginalTransactionID: subscriptionResponse.OrderID,
		ExpiresDateTimeMS:     int64(subscriptionResponse.ExpiryTimeMillis),
		ProductID:             productID,
		AutoRenew:             subscriptionResponse.AutoRenewing,
		GracePeriodExpiresMS:  int64(subscriptionResponse.ExpiryTimeMillis),
		VerificationData:      verifyData,
		StoreType:             androidStore,
	}
	ctx.RequestLogger.Info().Msg("Saving subscription")
	err = userSubscription.Save(ctx)
	if nil != err {
		return nil, err
	}
	ctx.RequestLogger.Info().Msg("Updating user")
	err = sharing.UpdateSharingSubscription(ctx, productID, userSubscription.ExpiresDateTimeMS,
		userSubscription.AutoRenew, userSubscription.GracePeriodExpiresMS, userSubscription.OriginalTransactionID)
	return user, nil
}
