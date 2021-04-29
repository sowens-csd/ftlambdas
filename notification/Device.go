package notification

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sharing"
)

type deviceNotificationTokenReq struct {
	AppInstallID string `json:"appInstallId" dynamodbav:"appInstallId"`
	DeviceToken  string `json:"deviceToken" dynamodbav:"deviceToken"`
}

// RegisterFromRequest registers a new endpoint from a JSON version of an DeviceNotificationTokenReq
func RegisterFromRequest(ftCtx awsproxy.FTContext, tokenReq string) error {
	var regReq deviceNotificationTokenReq
	reqJSON := []byte(tokenReq)
	err := json.Unmarshal(reqJSON, &regReq)
	if err != nil {
		ftCtx.RequestLogger.Info().Str("req", tokenReq).Msg("Failed to parse request")
		return err
	}
	if len(regReq.AppInstallID) == 0 || len(regReq.DeviceToken) == 0 {
		ftCtx.RequestLogger.Info().Str("req", tokenReq).Msg("Invalid request")
		return fmt.Errorf("Invalid request")
	}
	return Register(ftCtx, regReq.AppInstallID, regReq.DeviceToken)
}

// Register creates or updates an endpoint ARN for the given device token and then stores it
// in the OnlineUser for future use.
func Register(ftCtx awsproxy.FTContext, appInstallID string, deviceToken string) error {
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("noteToken", deviceToken).Msg("Registering token")
	user, err := sharing.LoadOnlineUser(ftCtx, ftCtx.UserID)
	if nil != err {
		return err
	}
	snsClient := getSNSClient(ftCtx)
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("noteToken", deviceToken).Msg("Loaded user")
	endpointArn := retrieveEndpointArn(ftCtx, appInstallID, user)

	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("endpointArn", endpointArn).Str("noteToken", deviceToken).Msg("Registering token")
	updateNeeded := false
	createNeeded := endpointArn == ""

	if createNeeded {
		// No platform endpoint ARN is stored; need to call createEndpoint.
		endpointArn, _ = createEndpoint(ftCtx, appInstallID, deviceToken, user)
		createNeeded = false
	}

	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("endpointArn", endpointArn).Str("noteToken", deviceToken).Msg("Checking attributes")
	// Look up the platform endpoint and make sure the data in it is current, even if
	// it was just created.
	geaReq :=
		sns.GetEndpointAttributesInput{
			EndpointArn: &endpointArn,
		}
	geaRes, err :=
		snsClient.GetEndpointAttributes(ftCtx.Context, &geaReq)
	if nil == err {
		updateNeeded = geaRes.Attributes["Token"] != deviceToken || strings.ToLower(geaRes.Attributes["Enabled"]) != "true"
	} else {
		// if err == ErrCodeNotFoundException {
		// 	createNeeded = true
		// } else {
		return err
		// }
	}

	if createNeeded {
		endpointArn, err = createEndpoint(ftCtx, appInstallID, deviceToken, user)
	}

	if updateNeeded {
		ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("endpointArn", endpointArn).Str("noteToken", deviceToken).Msg("Updating attributes")
		// The platform endpoint is out of sync with the current data;
		// update the token and enable it.
		attribs := make(map[string]string)
		attribs["Token"] = deviceToken
		enabled := "true"
		attribs["Enabled"] = enabled
		saeReq :=
			sns.SetEndpointAttributesInput{
				Attributes:  attribs,
				EndpointArn: &endpointArn,
			}
		_, err = snsClient.SetEndpointAttributes(ftCtx.Context, &saeReq)
		if nil != err {
			ftCtx.RequestLogger.Info().Str("appID", appInstallID).Str("endpointArn", endpointArn).Str("noteToken", deviceToken).Msg("Error updating attributes " + err.Error())
		}
	}
	return nil
}

func createEndpoint(ftCtx awsproxy.FTContext, appInstallID string, deviceToken string, user *sharing.OnlineUser) (string, error) {
	endpointArn := ""
	appArn := os.Getenv("snsAppArn")
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("noteToken", deviceToken).Str("appArn", appArn).Msg("Creating endpoint")
	cpeReq :=
		sns.CreatePlatformEndpointInput{
			PlatformApplicationArn: &appArn,
			Token:                  &deviceToken,
		}
	cpeRes, err := snsClient.CreatePlatformEndpoint(ftCtx.Context, &cpeReq)
	if nil == err {
		endpointArn = *cpeRes.EndpointArn
	} else {
		ftCtx.RequestLogger.Info().Str("appID", appInstallID).Str("noteToken", deviceToken).Str("appArn", appArn).Msg("Error creating endpoint " + err.Error())
		message := err.Error()
		endpointArn := findArnInMessage(message)
		if "" == endpointArn {
			return "", err
		}
	}
	storeEndpointArn(ftCtx, endpointArn, appInstallID, deviceToken, user)
	return endpointArn, nil
}

func findArnInMessage(message string) string {
	endpointArn := ""
	p := regexp.MustCompile(".*Endpoint (arn:aws:sns[^ ]+) already exists " +
		"with the same [Tt]oken.*")
	m := p.FindStringSubmatch(message)
	if nil != m {
		// The platform endpoint already exists for this token, but with
		// additional custom data that
		// createEndpoint doesn't want to overwrite. Just use the
		// existing platform endpoint.
		endpointArn = m[1]
	}
	return endpointArn
}

func storeEndpointArn(ftCtx awsproxy.FTContext, endpointArn string, appInstallID string, notificationToken string, user *sharing.OnlineUser) {
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("endpointArn", endpointArn).Str("noteToken", notificationToken).Msg("Saving endpoint")
	user.SetDeviceNotificationToken(appInstallID, endpointArn, notificationToken)
	err := user.UpdateAll(ftCtx)
	if nil != err {
		ftCtx.RequestLogger.Info().Str("endpoint", endpointArn).Str("appID", appInstallID).Str("noteToken", notificationToken).Msg("Error saving user " + err.Error())
	}
}

func retrieveEndpointArn(ftCtx awsproxy.FTContext, appInstallID string, user *sharing.OnlineUser) string {
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Msg("Retrieving endpoint")

	deviceToken := user.FindDeviceNotificationToken(appInstallID)
	if nil == deviceToken {
		return ""
	}
	ftCtx.RequestLogger.Debug().Str("appID", appInstallID).Str("endpointArn", deviceToken.SNSEndpoint).Msg("Found endpoint")
	return deviceToken.SNSEndpoint
}
