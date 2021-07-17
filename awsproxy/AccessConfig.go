package awsproxy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/rs/zerolog/log"
)

var connectPassword = ""
var clientSecret = ""
var refreshToken = ""
var accessToken = ""
var twilioAccount = ""
var twilioSID = ""
var twilioAppSID = ""
var twilioSecret = ""
var plivoAccount = ""
var plivoSID = ""
var plivoAppSID = ""
var plivoSecret = ""
var jwePassphrase = ""
var webrtcAccessToken = ""
var webrtcSecret = ""
var fcmKey = ""
var initialized = false
var storeInitialized = false
var parameterStore ParameterStore

// When adding a value here the resource permissions in serverless.yml may also have to be updated
// Secrets for App Store access
const connectPasswordPath = "/appstore/connect"
const clientSecretPath = "/appstore/client"
const refreshTokenPath = "/appstore/refresh"
const accessTokenPath = "/appstore/access"

// Secrets for Twillio
const twilioAccountPath = "/twilio/account"
const twilioSIDPath = "/twilio/sid"
const twilioAppSIDPath = "/twilio/appsid"
const twilioSecretPath = "/twilio/secret"

// Secrets for Plivo
const plivoAccountPath = "/plivo/account"
const plivoSIDPath = "/plivo/sid"
const plivoAppSIDPath = "/plivo/appsid"
const plivoSecretPath = "/plivo/secret"

const webrtcAccessTokenPath = "/webrtc/access"
const webrtcSecretPath = "/webrtc/secret"

const fcmKeyPath = "/fcm/secret"

// JWE
const jwePassphrasePath = "/jwe/passphrase"

// AccessParameters returns the current values of the access parameters, always call
// SetupAccessParameters first or they will be empty
func AccessParameters() (string, string, string, string) {
	return connectPassword, clientSecret, refreshToken, accessToken
}

// TwilioParameters returns the current values of the Twilio parameters, always call
// SetupAccessParameters first or they will be empty
func TwilioParameters() (string, string, string, string) {
	return twilioAccount, twilioSID, twilioSecret, twilioAppSID
}

// PlivoParameters returns the current values of the Plivo parameters, always call
// SetupAccessParameters first or they will be empty
func PlivoParameters(ctx context.Context) (string, string, string) {
	if len(plivoAccount) == 0 {
		SetupParameterStore(ctx)

		plivoAccount = getParameter(ctx, plivoAccountPath)
		plivoSID = getParameter(ctx, plivoSIDPath)
		plivoSecret = getParameter(ctx, plivoSecretPath)
	}
	return plivoAccount, plivoSID, plivoSecret
}

// PlivoAppSID returns the current values of the application SID, always call
// SetupAccessParameters first or it will be empty
func PlivoAppSID(ctx context.Context) string {
	if len(plivoAppSID) == 0 {
		SetupParameterStore(ctx)
		plivoAppSID = getParameter(ctx, plivoAppSIDPath)
	}

	return plivoAppSID
}

// WebRTCParameters returns the current values of the WebRTC parameters, always call
// SetupAccessParameters first or they will be empty
func WebRTCParameters() (string, string) {
	return webrtcAccessToken, webrtcSecret
}

// FCMParameters returns the current values of the FCM parameters, always call
// SetupAccessParameters first or they will be empty
func FCMParameters(ctx context.Context) string {
	SetupParameterStore(ctx)

	if len(fcmKey) == 0 {
		fcmKey = getParameter(ctx, fcmKeyPath)
	}
	return fcmKey
}

// SetJWEParameters useful for testing to control the values
func SetJWEParameters(passphrase string) {
	jwePassphrase = passphrase
}

// JWEParameters returns the current values of the JWE parameters, always call
// SetupAccessParameters first or they will be empty
func JWEParameters() string {
	return jwePassphrase
}

// UpdateAccessToken if the token has a value stuff it back into the parameter store
func UpdateAccessToken(ctx context.Context, updatedToken string) {
	if "" != updatedToken {
		accessToken = updatedToken
		parameterStore.PutParameter(ctx, accessTokenPath, updatedToken)
	}
}

func SetupParameterStore(ctx context.Context) {
	if storeInitialized {
		return
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}
	parameterStore = NewParameterStore(cfg)
	storeInitialized = true
}

// SetupAccessParameters sets the internal access params based on values from the AWS
// ParameterStore
func SetupAccessParameters(ctx context.Context) {
	if initialized {
		return
	}
	SetupParameterStore(ctx)
	initialized = true
	connectPassword = getParameter(ctx, connectPasswordPath)
	clientSecret = getParameter(ctx, clientSecretPath)
	refreshToken = getParameter(ctx, refreshTokenPath)
	accessToken = getParameter(ctx, accessTokenPath)
	if "-" == accessToken {
		accessToken = ""
	}

	twilioAccount = getParameter(ctx, twilioAccountPath)
	twilioSID = getParameter(ctx, twilioSIDPath)
	twilioAppSID = getParameter(ctx, twilioAppSIDPath)
	twilioSecret = getParameter(ctx, twilioSecretPath)

	webrtcAccessToken = getParameter(ctx, webrtcAccessTokenPath)
	webrtcSecret = getParameter(ctx, webrtcSecretPath)

	jwePassphrase = getParameter(ctx, jwePassphrasePath)
}

func getParameter(ctx context.Context, name string) string {
	val, err := parameterStore.GetParameter(ctx, name)
	if nil != err {
		log.Error().Str("param", name).Msg("Required parameter missing")
	}
	log.Debug().Str(name, "redacted").Msg("Found param")
	return val
}
