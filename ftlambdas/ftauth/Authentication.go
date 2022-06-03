package ftauth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	uuid "github.com/satori/go.uuid"
	"github.com/sowens-csd/ftlambdas/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftlambdas/ftdb"
	"github.com/sowens-csd/ftlambdas/ftlambdas/notification"
	"github.com/sowens-csd/ftlambdas/ftlambdas/sharing"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// AuthenticateResponse a token that contains the token to be used to authorize
// against the endpoints.
type AuthenticateResponse struct {
	AccessToken string `json:"accessToken"`
	Expiry      int64  `json:"expiry"`
}

// VerifyResponse is the response to a verify request, it
// contains the authToken for an addDevice request.
type VerifyResponse struct {
	AuthToken              string `json:"authToken"`
	Verified               bool   `json:"verified"`
	WaitingForVerification bool   `json:"waitingForVerification"`
}

// SignupResponseStatus the type of possible values for the status field
// of SignupResponse
type SignupResponseStatus int32

const (
	// SuccessSignupResponse signup completed successfully
	SuccessSignupResponse SignupResponseStatus = 0
	// BlockedExistingSignupResponse signup failed because a user with
	// that email already exists
	BlockedExistingSignupResponse SignupResponseStatus = 1
	// BlockedNewSignupResponse signup failed because no user with that
	// email exists
	BlockedNewSignupResponse SignupResponseStatus = 2
)

// SignupResponse is the response to a signup request, it
// contains the authToken for an addDevice request.
type SignupResponse struct {
	Success   bool                 `json:"success"`
	RequestID string               `json:"requestId,omitempty"`
	Status    SignupResponseStatus `json:"status"` // 0 - success, 1- Blocked existing, 2- Blocked new
}

// signupRequest holds both the request and a verified authentication record.
//
// Email - the email of the user that authenticated
//
// AuthToken - a hash of the token used to authenticate except in the special
//   case when there is an outstanding addDevice request being handled in which
//   case it is the actualy auth token until the request is granted or denied.
//
// RequestID - the verification code sent to the user that requested the auth
//  via email. They enter this code into the verification form to approve the request.
//
// ID - the userID of the user that made the request.
//
// AddDevice - true if a request was made to add a new device to this request.
//
// AllowSignup - true if the request should allow a new user to be created.
// This is to block new users from signing up when they should only be signing
// in as an existing user.
//
// AllowSignin - true if the request should allow an existing user to signin.
// This is used to block existing users from signing in when they should be
// creating a new prrofile.
//
// CreatedAt - the millis since epoch when the request was created.
//
// IsAdmin - true if the record is usable to log into the admin utility.
type signupRequest struct {
	ResourceID  string `json:"resourceId" dynamodbav:"resourceId"`
	ReferenceID string `json:"referenceId" dynamodbav:"referenceId"`
	Email       string `json:"email" dynamodbav:"email"`
	AuthToken   string `json:"authToken,omitempty" dynamodbav:"authToken,omitempty"`
	RequestID   string `json:"requestId,omitempty" dynamodbav:"requestId,omitempty"`
	ExternalID  string `json:"externalId,omitempty" dynamodbav:"externalId,omitempty"`
	ID          string `json:"id,omitempty" dynamodbav:"id,omitempty"`
	AddDevice   bool   `json:"addDevice,omitempty" dynamodbav:"addDevice,omitempty"`
	AllowSignup bool   `json:"allowSignup,omitempty" dynamodbav:"allowSignup,omitempty"`
	AllowSignin bool   `json:"allowSignin,omitempty" dynamodbav:"allowSignin,omitempty"`
	CreatedAt   int    `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
	IsAdmin     bool   `json:"isAdmin,omitempty" dynamodbav:"isAdmin,omitempty"`
}

type folktellsClaims struct {
	*jwt.Claims
	Email string `json:"email,omitempty"`
}

type mailIndexQueryResult struct {
	ID          string `json:"id" dynamodbav:"id"`
	Email       string `json:"email" dynamodbav:"email"`
	ResourceID  string `json:"resourceId" dynamodbav:"resourceId"`
	ReferenceID string `json:"referenceId" dynamodbav:"referenceId"`
}

type standardQueryResult struct {
	ResourceID  string `json:"resourceId" dynamodbav:"resourceId"`
	ReferenceID string `json:"referenceId" dynamodbav:"referenceId"`
}

// BlockedUserExistsError returned by Signup when the user is already in the DB
// and allowsSignup == false
type BlockedUserExistsError struct {
	error
	Email string
}

// BlockedNoUserError returned by Signup when the user is not already in the DB
// and allowsSignin == false
type BlockedNoUserError struct {
	error
	Email string
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// AuthorizeUser given an encrypted JWE token decrypt it and return
// the ID and email of the user.
// If the token can't be successfully decrypted and verified then
// return an error.
func AuthorizeUser(ftCtx awsproxy.FTContext, jweToken string, timeNow func() time.Time) (string, string, error) {
	jwe, err := jose.ParseEncrypted(jweToken)
	if err != nil {
		return "", "", err
	}
	passphrase := awsproxy.JWEParameters()
	rawJWT, err := jwe.Decrypt(passphrase)
	if err != nil {
		return "", "", err
	}
	parsedJWT, err := jwt.ParseSigned(string(rawJWT))
	if err != nil {
		return "", "", err
	}
	key := []byte(passphrase)
	jwk := jose.JSONWebKey{
		Key:       key,
		Algorithm: string(jose.HS256),
	}
	resultCl := folktellsClaims{}
	err = parsedJWT.Claims(jwk, &resultCl)
	if err != nil {
		return "", "", err
	}
	if nil == resultCl.Expiry {
		return "", "", fmt.Errorf("No expiry time found")
	}
	expiryTime := resultCl.Expiry.Time()
	if timeNow().After(expiryTime) {
		return "", "", fmt.Errorf("Expired token")
	}
	return resultCl.Claims.Subject, resultCl.Email, nil
}

// AuthenticateUserWithToken loads the identity from the DB using a token that encodes the
// email and bearer token. The token usually is taken from the Authentication header in
// the request. The format looks like: `Bearer {email} {token}`
func AuthenticateUserWithToken(ftCtx awsproxy.FTContext, token string, isAdmin bool, timeNow func() time.Time) (AuthenticateResponse, error) {
	emptyResponse := AuthenticateResponse{}
	tokenSlice := strings.Split(token, " ")
	var bearerToken string
	var email string
	if len(tokenSlice) > 2 {
		email = tokenSlice[len(tokenSlice)-2]
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}
	if len(bearerToken) == 0 {
		return emptyResponse, fmt.Errorf("No token")
	}
	return AuthenticateUser(ftCtx, email, bearerToken, isAdmin, timeNow)
}

// AuthenticateUser loads the identity from the DB using an email and a token.
// If there is a matching record then the user is authenticated and a JWT token
// is returned.
func AuthenticateUser(ftCtx awsproxy.FTContext, email, token string, isAdmin bool, timeNow func() time.Time) (AuthenticateResponse, error) {
	emptyResponse := AuthenticateResponse{}
	req, err := findSignupRequest(ftCtx, email, token)
	if nil != err || nil == req {
		return emptyResponse, fmt.Errorf("Unrecognized token")
	}
	if isAdmin && !req.IsAdmin {
		return emptyResponse, fmt.Errorf("Failed")
	}
	passphrase := awsproxy.JWEParameters()

	now := timeNow()
	claims := folktellsClaims{Claims: &jwt.Claims{
		Issuer:   "https://folktells.com/",
		Subject:  req.ID,
		Audience: []string{"https://api.folktells.com/"},
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(1 * time.Hour)),
	},
		Email: req.Email,
	}

	key := jose.SigningKey{Algorithm: jose.HS256, Key: []byte(passphrase)}
	var signerOpts = jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signer, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		return emptyResponse, err
	}
	builder := jwt.Signed(signer)

	builder = builder.Claims(claims)
	signedJWT, err := builder.CompactSerialize()
	if err != nil {
		return emptyResponse, err
	}
	rcpt := jose.Recipient{
		Algorithm: jose.PBES2_HS256_A128KW,
		Key:       passphrase,
	}
	enc, err := jose.NewEncrypter(jose.A128CBC_HS256, rcpt, nil)
	if err != nil {
		return emptyResponse, err
	}
	jweToken, err := enc.Encrypt([]byte(signedJWT))
	if err != nil {
		return emptyResponse, err
	}
	plaintextJWE, err := jweToken.CompactSerialize()
	if err != nil {
		return emptyResponse, err
	}
	offset := time.Duration(mrand.Intn(1000)) * time.Millisecond
	expiryNanos := now.Add(1 * time.Hour).Add(-1 * time.Minute).Add(-offset).UnixNano()
	expiryMillis := expiryNanos / 1000000

	return AuthenticateResponse{AccessToken: plaintextJWE, Expiry: expiryMillis}, nil
}

// VerifySignup takes a request that contains a previously issued token and
// verifies the associated token.
//
// This method is called when the submits the request code to verify their
// signup request. They must submit the code within one hour of receiving
// it. The token is verified if there is a signup request in the DB that has
// the request code and is less than an hour old. Being verified means that
// the token in the signup request should be associated with the email address
// in the request. If that email address is for an existing user the token
// will be associated with that user. If there is not an existing user with
// that email then a new user will be created.
func VerifySignup(ftCtx awsproxy.FTContext, requestID string) (VerifyResponse, error) {
	verifyResponse := VerifyResponse{Verified: false}
	req := loadVerifyRequest(ftCtx, requestID)
	if nil == req {
		return verifyResponse, fmt.Errorf("No matching request %s", requestID)
	}
	if req.AddDevice && len(req.AuthToken) == 0 {
		return VerifyResponse{Verified: false, WaitingForVerification: true}, nil
	}
	// This is a new user
	if len(req.ID) == 0 {
		ftCtx.RequestLogger.Debug().Str("req", requestID).Str("email", req.Email).Msg("Auth request no user")
		// If this is a new user send them an invite
		ou, err := sharing.LoadOnlineUserByEmail(ftCtx, req.Email)
		if err != nil {
			switch err.(type) {
			case *sharing.UserNotFoundError:
				ftCtx.RequestLogger.Debug().Str("req", requestID).Str("email", req.Email).Msg("Existing user not found")
				uuid := uuid.NewV4()
				newUser := sharing.NewOnlineUser(uuid.String(), req.Email, req.Email)
				ou = &newUser
				ftCtx.RequestLogger.Debug().Str("req", requestID).Str("email", req.Email).Msg("New user created")
				if err = ou.Save(ftCtx); nil != err {
					return verifyResponse, err
				}
				ftCtx.RequestLogger.Debug().Str("req", requestID).Str("email", req.Email).Msg("Saved user")
				sendWelcomeEmail(ftCtx, ou)
				break
			default:
				ftCtx.RequestLogger.Info().Str("req", requestID).Str("email", req.Email).Msg("Error loading user")
				return verifyResponse, err
			}
		} else {
			// Existing user who's been invited but not yet accepted so we need to accept the invitation
			if ou.IsPending() {
				ou.AcceptInvitation(ftCtx)
				sendWelcomeEmail(ftCtx, ou)
			}
		}
		req.ID = ou.ID
	}
	if !req.AddDevice {
		ftCtx.RequestLogger.Debug().Str("req", requestID).Msg("Adding verified signup")
		resID, err := authResourceIDFromEmail(ftCtx, req.Email)
		if nil != err {
			return verifyResponse, err
		}
		refID := ftdb.ReferenceIDFromAuthTokenHash(req.AuthToken)
		deleteExistingAuth(ftCtx, resID, req.ID)
		err = addSignupRequest(ftCtx, resID, refID, requestID, *req)
		if nil != err {
			return verifyResponse, err
		}
	}
	ftCtx.RequestLogger.Debug().Str("req", requestID).Msg("Deleting request")
	err := ftdb.DeleteItem(ftCtx, req.ResourceID, req.ReferenceID)

	verifyResponse = VerifyResponse{Verified: true}
	if req.AddDevice {
		verifyResponse.AuthToken = req.AuthToken
	}
	return verifyResponse, err
}

func sendWelcomeEmail(ftCtx awsproxy.FTContext, ou *sharing.OnlineUser) {
	if ou.AllowsEmail() {
		ftCtx.EmailSvc.SendEmail(ftCtx.Context, ou.Email, sharing.EnglishNewUserWelcome(), ftCtx.RequestLogger)
	}

}

// VerifyAddDevice is responsible for taking an add device verify request from a client that
// verifies an add device request from another client.
//
// The flow is that the new device initiates an add device request, that is sent
// to existing devices via a push notification. An existing device can choose to
// allow or reject the request. If they allow it then the token for the login is
// added to verification request. The new device polls the system to see if the
// token is there, if it is it completes the add device cycle. If instead the
// signup request is gone then the request was denied.
//
// This is the only time that the DB will contain an actual unhashed version of
// an auth token. From the time that the user approves the authentication until
// the new device requests the token the DB has the real value of the token. At
// all other times it has only a hashed version of it.
func VerifyAddDevice(ftCtx awsproxy.FTContext, signupJSON string) (SignupResponse, error) {
	signupResponse := SignupResponse{Success: false}
	var req signupRequest
	err := json.Unmarshal([]byte(signupJSON), &req)
	if err != nil {
		return signupResponse, err
	}
	ftCtx.RequestLogger.Debug().Str("email", ftCtx.Email).Str("requestID", req.RequestID).Msg("Verify add device")
	verifyReq := loadVerifyRequest(ftCtx, req.RequestID)
	if nil == verifyReq {
		return signupResponse, fmt.Errorf("No matching request %s", req.RequestID)
	}
	if verifyReq.ID != ftCtx.UserID {
		ftCtx.RequestLogger.Debug().Str("userID", ftCtx.UserID).Str("req.userID", verifyReq.ID).Msg("Verify add device user mismatch")
		return signupResponse, nil
	}
	if !req.AllowSignin {
		err = ftdb.DeleteItem(ftCtx, verifyReq.ResourceID, verifyReq.ReferenceID)
		return signupResponse, nil
	}
	type signupRequestUpdate struct {
		AuthToken string `json:":at" dynamodbav:":at"`
	}
	err = ftdb.UpdateItem(ftCtx, verifyReq.ResourceID, verifyReq.ReferenceID, "set authToken = :at", signupRequestUpdate{
		AuthToken: req.AuthToken,
	})
	if err != nil {
		return signupResponse, err
	}
	return SignupResponse{Success: true, Status: 0}, nil
}

// DeleteAuthentication delete the authentication record for the given user
func DeleteAuthentication(ftCtx awsproxy.FTContext, email string) error {
	auths, err := ftdb.QueryByResource(ftCtx, ftdb.ResourceIDFromEmail(email))
	if nil != err {
		return err
	}
	for _, auth := range auths {
		ftCtx.RequestLogger.Debug().Str("resID", auth.ResourceID).Str("refID", auth.ReferenceID).Msg("deleting")
		err = ftdb.DeleteItem(ftCtx, auth.ResourceID, auth.ReferenceID)
		if nil != err {
			ftCtx.RequestLogger.Err(err).Msg("Could not delete ")
		}
	}
	return nil
}

// Given a verification code (requestID) look for a matching outstanding authentication request.
// Since authentication requests are only valid for an hour only two searches need
// to be done. First for the current hour primary key with the provided verification
// code, then for the previous hour primary key with the same code. If no record exists
// that matches either combination then the load fails.
func loadVerifyRequest(ftCtx awsproxy.FTContext, requestID string) *signupRequest {
	resourceID := ftdb.ResourceIDForAuthRequest()
	referenceID := ftdb.ReferenceIDFromAuthRequestID(requestID)
	ftCtx.RequestLogger.Debug().Str("resID", resourceID).Str("refID", referenceID).Msg("Looking for verify request 1")
	req, err := loadSignupRequest(ftCtx, resourceID, referenceID)
	if nil != err {
		return nil
	}
	if nil == req {
		resourceID = ftdb.ResourceIDForPrevAuthRequest(1)
		ftCtx.RequestLogger.Debug().Str("resID", resourceID).Str("refID", referenceID).Msg("Looking for verify request 2")
		req, err = loadSignupRequest(ftCtx, resourceID, referenceID)
	}
	if nil != err {
		return nil
	}
	if nil != req {
		ftCtx.RequestLogger.Debug().Str("resID", resourceID).Str("refID", referenceID).Msg("Found verify request")
	}
	return req
}

// Load a specific signup request given its primary and secondary key.
func loadSignupRequest(ftCtx awsproxy.FTContext, resourceID string, referenceID string) (*signupRequest, error) {
	req := signupRequest{}
	found, err := ftdb.GetItem(ftCtx, resourceID, referenceID, &req)
	if nil != err {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &req, nil
}

// Find an existing authentication that matches the given email and token.
// The email is used to create the primary key directly while the token has
// to be hashed and then compared against the hashed value of the secondary
// key of the records found using the email address.
// If a record is found then this is a successful authentication since only
// someone in possession of both the email and the token could have made the
// request.
func findSignupRequest(ftCtx awsproxy.FTContext, email, token string) (*signupRequest, error) {
	resID, err := authResourceIDFromEmail(ftCtx, email)
	if nil != err {
		return nil, err
	}
	ftCtx.RequestLogger.Debug().Str("resID", resID).Msg("findSignupRequest 1")
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		KeyConditionExpression: aws.String("resourceId = :resID "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":resID": &types.AttributeValueMemberS{Value: resID},
		},
	})
	if err != nil {
		ftCtx.RequestLogger.Err(err).Msg("findSignupRequest 2")
		return nil, err
	} else if result.Count == 0 {
		ftCtx.RequestLogger.Debug().Msg("findSignupRequest no matching requests")
		return nil, fmt.Errorf("Not found")
	}

	resIDResults := []signupRequest{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &resIDResults)
	if nil != err {
		return nil, err
	}
	var foundRequest *signupRequest
	for _, resIDResult := range resIDResults {
		ftCtx.RequestLogger.Debug().Str("resID", resIDResult.ResourceID).Str("refID", resIDResult.ReferenceID).Msg("findSignupRequest 3")
		if ftdb.IsAuthResource(resIDResult.ResourceID) && ftdb.IsAuthReference(resIDResult.ReferenceID) {
			tokenHash := ftdb.AuthTokenHashFromReferenceID(resIDResult.ReferenceID)
			ftCtx.RequestLogger.Debug().Str("tokenHash", tokenHash).Msg("findSignupRequest 3")
			compareErr := compareHashedAuthComponent(ftCtx, tokenHash, token)
			if nil == compareErr {
				foundRequest = &resIDResult
			} else {
				ftCtx.RequestLogger.Debug().Msg("password hash compare failed")
			}
		}
	}
	if nil == foundRequest {
		ftCtx.RequestLogger.Debug().Msg("findSignupRequest not found")
		return nil, fmt.Errorf("Not found")
	}
	return foundRequest, nil
}

// Signup stores a signup request associated with the provided email and token.
//
// Signup takes a JSON formatted signup request that includes a token value and an
// email address. This is a request to associate a new token with an email address. At
// this point the system doesn't care whether the email is for an existing user or a
// new user, or if they already have a token associated with their account. All it
// does care about is whether the client that provided the token and email is able to
// receive emails at the provided email address.
// An email is sent to the provided address, the mail has a request code in it. Sending that code
// back to VerifySignup confirms ownership of the email address and converts the provisional
// association of the token into a verified association. At that point it also checks to
// see if there is a user account already, and if it had a previous token.
func Signup(ftCtx awsproxy.FTContext, signupJSON string, client *http.Client) (SignupResponse, error) {
	signupResponse := SignupResponse{Success: false}
	var req signupRequest
	err := json.Unmarshal([]byte(signupJSON), &req)
	if err != nil {
		return signupResponse, err
	}
	ftCtx.RequestLogger.Debug().Str("authToken", req.AuthToken).Msg("Signup")
	if isEmailValid(req.Email) {
		var requestID string
		if strings.HasSuffix(req.Email, "@auto.folktells.com") && strings.HasPrefix(ftdb.GetTableName(), "dev-") {
			requestID = req.RequestID
		} else {
			requestID, err = generateOTP(6)
			if nil != err {
				return signupResponse, err
			}

		}
		if req.AddDevice {
			return addDeviceSignup(ftCtx, req, requestID, client)
		}
		return newSignup(ftCtx, req, requestID)

	}
	ftCtx.RequestLogger.Debug().Str("email", req.Email).Msg("Invalid email")
	return signupResponse, fmt.Errorf("No valid email address found")

}

// Once an authentication request has been successfully verified that
// request is deleted, after creating a new authentication.
func deleteExistingAuth(ftCtx awsproxy.FTContext, resID, userID string) {
	ftCtx.RequestLogger.Debug().Str("resID", resID).Str("userID", userID).Msg("Querying for signups to delete")
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		KeyConditionExpression: aws.String("resourceId = :resID "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":resID": &types.AttributeValueMemberS{Value: resID},
		},
	})
	if err != nil {
		ftCtx.RequestLogger.Err(err).Msg("Error querying for signups to delete")
		return
	} else if result.Count == 0 {
		ftCtx.RequestLogger.Debug().Msg("No signups to delete")
		return
	}
	resIDResults := []signupRequest{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &resIDResults)
	if nil != err {
		ftCtx.RequestLogger.Err(err).Msg("Error unmarshalling signups to delete")
		return
	}
	for _, resIDResult := range resIDResults {
		ftCtx.RequestLogger.Debug().Str("resID", resIDResult.ResourceID).Str("refID", resIDResult.ReferenceID).Msg("findSignupRequest 3")
		if (ftdb.IsAuthResource(resIDResult.ResourceID) || ftdb.IsP2PAuthResource(resIDResult.ResourceID)) && ftdb.IsAuthReference(resIDResult.ReferenceID) && resIDResult.ID == userID {
			ftCtx.RequestLogger.Debug().Str("resID", resIDResult.ResourceID).Str("refID", resIDResult.ReferenceID).Msg("Deleting")
			ftdb.DeleteItem(ftCtx, resIDResult.ResourceID, resIDResult.ReferenceID)
		}
	}
}

// Add a new signup record for an add device request.
// This request is used to hold the authToken if the user
// approves the request. This method also sends a push notification
// to an existing device of the user so that they can approve the
// request in Folktells.
func addDeviceSignup(ftCtx awsproxy.FTContext, req signupRequest, requestID string, client *http.Client) (SignupResponse, error) {
	ftCtx.RequestLogger.Debug().Msg("addDeviceSignup")
	signupResponse := SignupResponse{Success: false}
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		IndexName:              aws.String(ftdb.EmailIndex),
		KeyConditionExpression: aws.String("email = :email "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: strings.ToLower(req.Email)},
		},
	})
	if err != nil {
		ftCtx.RequestLogger.Err(err).Msg("error finding email")
		return signupResponse, err
	} else if result.Count == 0 {
		ftCtx.RequestLogger.Debug().Msg("email not found")
		return signupResponse, fmt.Errorf("Invalid")
	}

	byMailResults := []mailIndexQueryResult{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &byMailResults)
	if nil != err {
		return signupResponse, err
	}
	var userID string
	var foundRequest *signupRequest
	for _, mailResult := range byMailResults {
		if ftdb.IsAuthResource(mailResult.ResourceID) && ftdb.IsAuthReference(mailResult.ReferenceID) {
			userID = mailResult.ID
			foundRequest, err = loadSignupRequest(ftCtx, mailResult.ResourceID, mailResult.ReferenceID)
			if nil != err {
				return signupResponse, err
			}
		}
	}
	if len(userID) > 0 {
		ftCtx.RequestLogger.Debug().Str("userID", userID).Msg("found user")
		ou, err := sharing.LoadOnlineUser(ftCtx, userID)
		if nil != err {
			return signupResponse, err
		}
		addDevReq := signupRequest{
			Email:     req.Email,
			ID:        foundRequest.ID,
			RequestID: requestID,
			AddDevice: true,
		}
		err = addSignupRequestWithRetry(ftCtx, requestID, addDevReq)
		if nil != err {
			return signupResponse, err
		}
		notification.SendAuthVerifyCommand(ftCtx, requestID, ou, client)
	}
	return SignupResponse{Status: 0, Success: true, RequestID: requestID}, nil
}

// Add a new signup request to the DB when a user requests a new
// signup. This is different from an add device request because
// it replaces any previous authentications for the user, invalidating
// the token used by those authentications.
func newSignup(ftCtx awsproxy.FTContext, req signupRequest, requestID string) (SignupResponse, error) {
	ftCtx.RequestLogger.Debug().Str("requestID", requestID).Msg("newSignup")
	hashedToken, err := hashAuthComponent(ftCtx, req.AuthToken)
	if nil != err {
		return SignupResponse{}, err
	}
	newReq := signupRequest{
		Email:     req.Email,
		AuthToken: hashedToken,
	}
	if !req.AllowSignin || !req.AllowSignup {
		ftCtx.RequestLogger.Debug().Str("email", req.Email).Msg("user search")
		ou, err := sharing.LoadOnlineUserByEmail(ftCtx, req.Email)
		hasUser := nil != ou
		if !req.AllowSignup && !hasUser {
			if nil != err {
				ftCtx.RequestLogger.Info().Err(err).Msg("no matching user, signup not allowed")
			} else {
				ftCtx.RequestLogger.Debug().Msg("no matching user, signup not allowed")
			}
			return SignupResponse{
					Success: false,
					Status:  BlockedNewSignupResponse},
				nil
		} else if !req.AllowSignin && hasUser && ou.IsAccepted() {
			ftCtx.RequestLogger.Debug().Msg("found user, signin not allowed")
			return SignupResponse{
					Success: false,
					Status:  BlockedExistingSignupResponse},
				nil
		}
	}
	ftCtx.RequestLogger.Debug().Str("authToken", hashedToken).Msg("newSignup")
	err = addSignupRequestWithRetry(ftCtx, requestID, newReq)
	if nil != err {
		return SignupResponse{}, err
	}
	ftCtx.EmailSvc.SendEmail(ftCtx.Context, req.Email, EnglishAuthCodeContent(requestID),
		ftCtx.RequestLogger)
	return SignupResponse{Success: true, Status: SuccessSignupResponse}, nil
}

func addSignupRequestWithRetry(ftCtx awsproxy.FTContext, requestID string, req signupRequest) error {
	var err error
	for retry := 0; retry < 5; retry++ {
		err = addSignupRequest(ftCtx, ftdb.ResourceIDForAuthRequest(), ftdb.ReferenceIDFromAuthRequestID(requestID), requestID, req)
		if nil == err {
			break
		}
	}
	return err
}

// Put the signup request in the DB
func addSignupRequest(ftCtx awsproxy.FTContext, resourceID string, referenceID string, requestID string, signup signupRequest) error {
	now := time.Now()
	sec := now.UTC().Unix()
	createdAt := int(sec * 1000)

	ftCtx.RequestLogger.Debug().Str("authToken", signup.AuthToken).Msg("addSignupRequest 1")
	req := signupRequest{
		Email:       signup.Email,
		AuthToken:   signup.AuthToken,
		ID:          signup.ID,
		RequestID:   requestID,
		AddDevice:   signup.AddDevice,
		ExternalID:  signup.ExternalID,
		ResourceID:  resourceID,
		ReferenceID: referenceID,
		CreatedAt:   createdAt,
	}
	ftCtx.RequestLogger.Debug().Str("authToken", req.AuthToken).Msg("addSignupRequest 2")
	reqMap, err := attributevalue.MarshalMap(req)
	if nil != err {
		return err
	}
	// ftCtx.RequestLogger.Debug().Str("reqMap", reqMap).Msg("addSignupRequest 3")
	_, err = ftCtx.DBSvc.PutItem(ftCtx.Context, &dynamodb.PutItemInput{
		TableName:           aws.String(ftdb.GetTableName()),
		ConditionExpression: aws.String("attribute_not_exists(referenceId) and attribute_not_exists(authToken)"),
		Item:                reqMap,
	})
	if nil != err {
		ftCtx.RequestLogger.Error().Str("req", requestID).Err(err).Msg(fmt.Sprintf("save request failed %s", err.Error()))
	} else {
		ftCtx.RequestLogger.Info().Str("req", requestID).Msg("save signup request succeeded")
	}
	return err
}

// CleanupOutstanding remove any waiting auth requests that haven't been
// verified. These are requests before successful verification.
func CleanupOutstanding(ftCtx awsproxy.FTContext) error {
	for hours := 2; hours < 6; hours++ {
		resID := ftdb.ResourceIDForPrevAuthRequest(hours)
		outstanding, err := ftdb.QueryByResource(ftCtx, resID)
		if nil != err {
			return err
		}
		for _, singleRecord := range outstanding {
			ftdb.DeleteItem(ftCtx, singleRecord.ResourceID, singleRecord.ReferenceID)
		}
	}
	return nil
}

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

func authResourceIDFromEmail(ftCtx awsproxy.FTContext, email string) (string, error) {
	// encrypted, err := encryptUsingPassphrase(ftCtx, email)
	// if err != nil {
	// 	return "", err
	// }
	return ftdb.ResourceIDFromEmail(email), nil
}

// This encrypts the email using a passphrase from the
// system parameters.
func encryptUsingPassphrase(ftCtx awsproxy.FTContext, email string) (string, error) {
	passphrase := awsproxy.JWEParameters()
	rcpt := jose.Recipient{
		Algorithm: jose.PBES2_HS256_A128KW,
		Key:       passphrase,
	}
	enc, err := jose.NewEncrypter(jose.A128CBC_HS256, rcpt, nil)
	if err != nil {
		return "", err
	}
	encrypted, err := enc.Encrypt([]byte(email))
	if err != nil {
		return "", err
	}
	plaintext, err := encrypted.CompactSerialize()
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

// Hash the token so that it can be safely stored in the DB
// and used for future comparison and then add the prefix that
// identifies it as auth reference ID.
func authReferenceIDFromToken(ftCtx awsproxy.FTContext, token string) (string, error) {
	hashed, err := hashAuthComponent(ftCtx, token)
	if nil != err {
		return "", err
	}
	return ftdb.ReferenceIDFromAuthTokenHash(hashed), nil
}

// Hash the token so that it can be safely stored in the DB
// and used for future comparison.
func hashAuthComponent(ftCtx awsproxy.FTContext, authComponent string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(authComponent), bcrypt.DefaultCost)
	if nil != err {
		return "", err
	}
	return string(hashed), nil
}

// Compare a previously hashed value with a provided value to
// ensure that the hash could have come from the provided value.
// The user provides their token, that token is compared against
// the hashed value of the token in the DB to see if they match.
func compareHashedAuthComponent(ftCtx awsproxy.FTContext, hash, authComponent string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(authComponent))
}

const otpChars = "1234567890"

// Generate a one time code to use as a verification code for
// an authentication request.
func generateOTP(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

func (e BlockedUserExistsError) Error() string {
	return fmt.Sprintf("Already a user for this email %s", e.Email)
}

func (e BlockedNoUserError) Error() string {
	return fmt.Sprintf("No existing user for signin %s", e.Email)
}
