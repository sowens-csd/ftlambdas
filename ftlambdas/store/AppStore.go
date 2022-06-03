package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftlambdas/sharing"
)

type verifyReceiptRequestData struct {
	ReceiptData            string `json:"receipt-data"`
	Password               string `json:"password"`
	ExcludeOldTransactions bool   `json:"exclude-old-transactions"`
}

// PendingRenewal information about the renewal status of an auto renewable product
type PendingRenewal struct {
	AutoRenewProductID        string `json:"auto_renew_product_id"`
	AutoRenewStatus           string `json:"auto_renew_status"`
	ExpirationIntent          string `json:"expiration_intent"`
	OriginalTransactionID     string `json:"original_transaction_id"`
	IsInBillingRetryPeriod    string `json:"is_in_billing_retry_period"`
	ProductID                 string `json:"product_id"`
	GracePeriodExpiresDate    string `json:"grace_period_expires_date"`
	GracePeriodExpiresDateMS  string `json:"grace_period_expires_date_ms"`
	GracePeriodExpiresDatePST string `json:"grace_period_expires_date_pst"`
	PriceConsentStatus        string `json:"price_consent_status"`
}

// InAppPurchase in-app purchases made in a particular receipt
type InAppPurchase struct {
	Quantity                string `json:"quantity"`
	ProductID               string `json:"product_id"`
	TransactionID           string `json:"transaction_id"`
	OriginalTransactionID   string `json:"original_transaction_id"`
	PurchaseDate            string `json:"purchase_date"`
	PurchaseDateMS          string `json:"purchase_date_ms"`
	PurchaseDatePOST        string `json:"purchase_date_pst"`
	OriginalPurchaseDate    string `json:"original_purchase_date"`
	OriginalPurchaseDateMS  string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST string `json:"original_purchase_date_pst"`
	ExpiresDate             string `json:"expires_date"`
	ExpiresDateMS           string `json:"expires_date_ms"`
	ExpiresDatePST          string `json:"expires_date_pst"`
	WebOrderLineItemID      string `json:"web_order_line_item_id"`
	IsTrialPeriod           string `json:"is_trial_period"`
	IsInIntroOfferPeriod    string `json:"is_in_intro_offer_period"`
}

// Receipt information about a single receipt
type Receipt struct {
	ReceiptType                string          `json:"receipt_type"`
	AdamID                     int             `json:"adam_id"`
	AppItemID                  int             `json:"app_item_id"`
	BundleID                   string          `json:"bundle_id"`
	ApplicationVersion         string          `json:"application_version"`
	DownloadID                 int             `json:"download_id"`
	VersionExternalIdentifier  int             `json:"version_external_identifier"`
	ReceiptCreationDate        string          `json:"receipt_creation_date"`
	ReceiptCreationDateMS      string          `json:"receipt_creation_date_ms"`
	ReceiptCreationDatePST     string          `json:"receipt_creation_date_pst"`
	RequestDate                string          `json:"request_date"`
	RequestDateMS              string          `json:"request_date_ms"`
	RequestDatePST             string          `json:"request_date_pst"`
	OriginalPurchaseDate       string          `json:"original_purchase_date"`
	OriginalPurchaseDateMS     string          `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST    string          `json:"original_purchase_date_pst"`
	OriginalApplicationVersion string          `json:"original_application_version"`
	InApp                      []InAppPurchase `json:"in_app"`
}

// LatestReceipt information about in_app purchases
type LatestReceipt struct {
	Quantity                    string `json:"quantity"`
	ProductID                   string `json:"product_id"`
	TransactionID               string `json:"transaction_id"`
	OriginalTransactionID       string `json:"original_transaction_id"`
	PurchaseDate                string `json:"purchase_date"`
	PurchaseDateMS              string `json:"purchase_date_ms"`
	PurchaseDatePST             string `json:"purchase_date_pst"`
	OriginalPurchaseDate        string `json:"original_purchase_date"`
	OriginalPurchaseDateMS      string `json:"original_purchase_date_ms"`
	OriginalPurchaseDatePST     string `json:"original_purchase_date_pst"`
	ExpiresDate                 string `json:"expires_date"`
	ExpiresDateMS               string `json:"expires_date_ms"`
	ExpiresDatePST              string `json:"expires_date_pst"`
	CancellationReason          string `json:"cancellation_reason"`
	CancellationDate            string `json:"cancellation_date"`
	CancellationDateMS          string `json:"cancellation_date_ms"`
	CancellationDatePST         string `json:"cancellation_date_pst"`
	WebOrderLineItemID          string `json:"web_order_line_item_id"`
	IsTrialPeriod               string `json:"is_trial_period"`
	IsInIntroOfferPeriod        string `json:"is_in_intro_offer_period"`
	SubscriptionGroupIdentifier string `json:"subscription_group_identifier"`
	IsUpgraded                  string `json:"is_upgraded"`
	PromotionalOfferID          string `json:"promotional_offer_id"`
}

// VerifyReceiptResponse the response from the /verifyReceipt endpoint that contains
// details about the products and their status.
type VerifyReceiptResponse struct {
	Environment        string           `json:"environment"`
	IsRetryable        bool             `json:"is_retryable"`
	Receipt            Receipt          `json:"receipt"`
	LatestReceiptInfo  []LatestReceipt  `json:"latest_receipt_info"`
	LatestReceipt      string           `json:"latest_receipt"`
	Status             int              `json:"status"`
	PendingRenewalInfo []PendingRenewal `json:"pending_renewal_info"`
}

const statusSandboxRequest = 21007
const statusVerified = 0
const statusIsNotVerified = 21003
const statusSharedSecretMismatch = 21004
const statusTemporaryFailure = 21005
const statusTemporaryInternalError = 21009
const statusUserAccountNotFound = 21010
const statusRequestNotPost = 21000
const productionEndpoint = "https://buy.itunes.apple.com/verifyReceipt"
const sandboxEndpoint = "https://sandbox.itunes.apple.com/verifyReceipt"

// VerifyAppleReceipt verifies a receipt with the app store
func VerifyAppleReceipt(ctx awsproxy.FTContext, verifyRequest verifyPurchaseRequest, accessConfig AccessConfig, client *http.Client) (VerifyReceiptResponse, error) {
	emptyResponse := VerifyReceiptResponse{}
	verifyResp, err := verifyAppleReceiptOnEndpoint(ctx, verifyRequest.VerifyData, accessConfig, productionEndpoint, client)
	if nil != err {
		return emptyResponse, err
	}
	if statusSandboxRequest == verifyResp.Status {
		return verifyAppleReceiptOnEndpoint(ctx, verifyRequest.VerifyData, accessConfig, sandboxEndpoint, client)
	}
	return verifyResp, err
}

// VerifyAppleReceipt verifies a receipt with the app store
func verifyAppleReceiptOnEndpoint(ctx awsproxy.FTContext, verifyData string, accessConfig AccessConfig, endpoint string, client *http.Client) (VerifyReceiptResponse, error) {
	emptyResponse := VerifyReceiptResponse{}
	receiptData := verifyReceiptRequestData{
		ReceiptData:            verifyData,
		Password:               accessConfig.ConnectPassword,
		ExcludeOldTransactions: true,
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(receiptData)
	if nil != err {
		return emptyResponse, err
	}
	ctx.RequestLogger.Info().Str("verifyEndpoint", endpoint).Msg("About to verify")
	resp, err := client.Post(endpoint, "application/json", buf)
	if nil != err {
		return emptyResponse, err
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("HTTP Status on verify call %d", resp.StatusCode))
	if http.StatusOK != resp.StatusCode {
		return emptyResponse, fmt.Errorf("Non 200 status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return emptyResponse, err
	}
	stringBody := string(respBody)
	// fmt.Printf(stringBody)
	inputJSON := []byte(string(stringBody))
	var verifyResponse VerifyReceiptResponse
	err = json.Unmarshal(inputJSON, &verifyResponse)
	if nil != err {
		return VerifyReceiptResponse{}, err
	}
	ctx.RequestLogger.Info().Msg(fmt.Sprintf("Got a response: %d", verifyResponse.Status))
	return verifyResponse, nil
}

// IsVerified returns true if the response successfully verified the receipt, false otherwise.
func (verifyResponse *VerifyReceiptResponse) IsVerified() bool {
	return verifyResponse.Status == statusVerified
}

// IsNotVerifiedReceipt returns true if the response positively confirms that the receipt is not
// a verified receipt, in other words the product should not be provided, false otherwise.
func (verifyResponse *VerifyReceiptResponse) IsNotVerifiedReceipt() bool {
	return verifyResponse.Status == statusIsNotVerified
}

// IsTemporaryFailure returns true if the App store failed but the request should be retried
func (verifyResponse *VerifyReceiptResponse) IsTemporaryFailure() bool {
	return verifyResponse.Status == statusTemporaryFailure ||
		verifyResponse.Status == statusTemporaryInternalError ||
		verifyResponse.Status == statusUserAccountNotFound ||
		verifyResponse.Status == statusSharedSecretMismatch ||
		verifyResponse.Status == statusRequestNotPost
}

// OriginalTransactionIDs returns a list of unique original transaction IDs for all products
// in the receipt.
func (verifyResponse *VerifyReceiptResponse) OriginalTransactionIDs() []string {
	idMap := make(map[string]bool)
	for _, purchase := range verifyResponse.Receipt.InApp {
		idMap[purchase.OriginalTransactionID] = true
	}
	originalIDs := make([]string, 0, len(idMap))
	for key := range idMap {
		originalIDs = append(originalIDs, key)
	}
	return originalIDs
}

// FindMostRecentSubscription returns the productID and expiry time in ms since epoch for the
// newest subscription that corresponds to the given originalTransactionID
func (verifyResponse *VerifyReceiptResponse) FindMostRecentSubscription(originalTransactionID string) (string, int64, bool, int64) {
	var productID string
	var newestExpiry int64
	var gracePeriod int64
	autoRenew := false
	for _, purchase := range verifyResponse.Receipt.InApp {
		purchaseExpire, err := strconv.ParseInt(purchase.ExpiresDateMS, 10, 64)
		if nil != err {
			continue
		}
		if purchase.OriginalTransactionID == originalTransactionID && purchaseExpire > newestExpiry {
			productID = purchase.ProductID
			newestExpiry = purchaseExpire
		}
	}
	foundAuto := false
	for _, renewal := range verifyResponse.PendingRenewalInfo {
		if renewal.OriginalTransactionID == originalTransactionID &&
			productID == renewal.ProductID {
			if "" != renewal.GracePeriodExpiresDateMS {
				graceExpire, err := strconv.ParseInt(renewal.GracePeriodExpiresDateMS, 10, 64)
				if nil != err {
					continue
				}
				if graceExpire > gracePeriod {
					foundAuto = true
					autoRenew = renewal.AutoRenewStatus == "1"
					gracePeriod = graceExpire
				}
			} else if !foundAuto {
				autoRenew = renewal.AutoRenewStatus == "1"
			}
		}
	}
	return productID, newestExpiry, autoRenew, gracePeriod
}

// UpdateSubscriptions updates the DB based on the subscription information in the receipt
//
// There are two separate cases. The first is a new subscription. In that case the subscription
// will be created for the current user and their user record will be updated to reflect the state
// the subscription. The second case is an existing subscription entry for the original transaction
// ID. In that case the user that is already associated with that transaction ID must match the
// the current user. If they do not match it is an error and the method logs and exits. If they do
// match the subscription and user records are updated to ensure the expiry date is correct.
func (verifyResponse *VerifyReceiptResponse) UpdateSubscriptions(ctx awsproxy.FTContext, verifyData string) (*sharing.OnlineUser, error) {
	transactionIDs := verifyResponse.OriginalTransactionIDs()
	ctx.RequestLogger.Info().Msg("Loading user")
	user, err := sharing.LoadOnlineUser(ctx, ctx.UserID)
	if nil != err {
		return nil, err
	}
	ctx.RequestLogger.Debug().Int64("lastExpiry", user.LastExpiryVerify).Msg("In update before")
	for _, transactionID := range transactionIDs {
		ctx.RequestLogger.Info().Msg(fmt.Sprintf("Updating transaction %s", transactionID))
		productID, expiresMS, autoRenew, gracePeriod := verifyResponse.FindMostRecentSubscription(transactionID)
		if len(productID) == 0 {
			continue
		}
		storeType := appleStore
		if verifyResponse.Environment == mockAppleStore {
			storeType = mockAppleStore
		}
		ctx.RequestLogger.Info().Msg(fmt.Sprintf("Most recent product: %s, expiry: %d", productID, expiresMS))
		userSubscription := UserSubscription{
			UserID:                ctx.UserID,
			TransactionID:         transactionID,
			OriginalTransactionID: transactionID,
			ExpiresDateTimeMS:     expiresMS,
			ProductID:             productID,
			AutoRenew:             autoRenew,
			GracePeriodExpiresMS:  gracePeriod,
			VerificationData:      verifyData,
			StoreType:             storeType,
		}
		ctx.RequestLogger.Info().Msg("Saving subscription")
		err := userSubscription.Save(ctx)
		if nil != err {
			return nil, err
		}
		ctx.RequestLogger.Info().Msg("Updating user")
		err = sharing.UpdateSharingSubscription(ctx, productID, userSubscription.ExpiresDateTimeMS,
			userSubscription.AutoRenew, userSubscription.GracePeriodExpiresMS, userSubscription.OriginalTransactionID)
	}
	ctx.RequestLogger.Debug().Int64("lastExpiry", user.LastExpiryVerify).Msg("In update after")
	return user, nil
}
