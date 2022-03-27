package sharing

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sms"
)

type assignRequest struct {
	IsMock      bool   `json:"isMock"`
	PhoneNumber string `json:"phoneNumber"`
}

// UpdateSharingSubscription changes the current subscription information for a user.
func UpdateSharingSubscription(ctx awsproxy.FTContext, productID string, expiry int64, autoRenew bool, gracePeriod int64, originalTransactionID string) error {
	user, err := LoadOnlineUser(ctx, ctx.UserID)
	if nil != err {
		return err
	}
	ctx.RequestLogger.Debug().Int64("lastExpiry", user.LastExpiryVerify).Msg("In update before")

	err = user.UpdateSharingSubscription(ctx, productID, expiry, autoRenew, gracePeriod, originalTransactionID)
	if nil != err {
		return err
	}
	if user.IsSubscriptionExpired() && user.HasPhone() {
		err = sms.DeleteNumber(ctx, user.Phone)
		if nil != err {
			return err
		}
		user.Phone = ""
		user.UpdateAll(ctx)
	}
	return nil
}

// AssignNumber ensures that the current user has a requested phone number or returns an error
func AssignNumber(ftCtx awsproxy.FTContext, body string, client *http.Client) error {
	var request assignRequest
	requestJSON := []byte(body)
	err := json.Unmarshal(requestJSON, &request)
	if nil != err {
		return err
	}
	if len(request.PhoneNumber) == 0 {
		return fmt.Errorf("Phone number is required.")
	}
	phoneNumberSID, phoneNumber, err := sms.AssignNumber(ftCtx, request.PhoneNumber, request.IsMock, client)
	if nil != err {
		return err
	}
	onlineUser, err := LoadOnlineUser(ftCtx, ftCtx.UserID)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("Error loading user.")
		return err
	}
	onlineUser.Phone = phoneNumber
	onlineUser.TwilioPhoneSID = phoneNumberSID
	ftCtx.RequestLogger.Info().Str("phoneNumber", phoneNumber).Str("phoneNumberSID", phoneNumberSID).Msg("Updating user.")
	err = onlineUser.UpdateAll(ftCtx)
	return err
}
