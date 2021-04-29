package ftauth

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var authTestDBData = awsproxy.TestDBData{
	awsproxy.TestDBDataRecord{
		ResourceID:  authResourceID1,
		ReferenceID: authReferenceID1,
		QueryKey:    authResourceID1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:  authResourceID1,
			ftdb.ReferenceIDField: authReferenceID1,
			ftdb.IDField:          userID1,
			ftdb.EmailField:       email1,
		},
	},
	awsproxy.TestDBDataRecord{
		ResourceID:  ftdb.ResourceIDFromUserID(userID1),
		ReferenceID: ftdb.ReferenceIDFromUserID(userID1),
		QueryKey:    email1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:  ftdb.ResourceIDFromUserID(userID1),
			ftdb.ReferenceIDField: ftdb.ReferenceIDFromUserID(userID1),
			ftdb.IDField:          userID1,
			ftdb.EmailField:       email1,
		},
	},
	awsproxy.TestDBDataRecord{
		ResourceID:  ftdb.ResourceIDForAuthRequest(),
		ReferenceID: ftdb.ReferenceIDFromAuthRequestID(existingRequestID1),
		QueryKey:    email3,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:  ftdb.ResourceIDForAuthRequest(),
			ftdb.ReferenceIDField: ftdb.ReferenceIDFromAuthRequestID(existingRequestID1),
			ftdb.IDField:          userID3,
			ftdb.EmailField:       email3,
		},
	},
	awsproxy.TestDBDataRecord{
		ResourceID:  ftdb.ResourceIDForAuthRequest(),
		ReferenceID: ftdb.ReferenceIDFromAuthRequestID(newUserRequestID1),
		QueryKey:    newUserEmail,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:  ftdb.ResourceIDForAuthRequest(),
			ftdb.ReferenceIDField: ftdb.ReferenceIDFromAuthRequestID(newUserRequestID1),
			ftdb.EmailField:       newUserEmail,
		},
	},
}

func TestSignupFailsWithInvalidEmail(t *testing.T) {
	signupAndExpect(signupRequest{Email: invalidEmail}, true, false, fmt.Sprintf("Email should be invalid %s", invalidEmail), t)
}

func TestSignupSucceedsForNewAccount(t *testing.T) {
	signupAndExpect(signupRequest{Email: email2, AllowSignup: true, AuthToken: token1},
		false, true, "Expected success on signup", t)
}

func TestSignupFailsForNewAccountWhenSignupNotAllowed(t *testing.T) {
	signupAndExpect(signupRequest{Email: email2, AllowSignup: false, AuthToken: token1},
		false, false, "Expected fail on signup", t)
}

func TestSignupFailsForExistingAccountWhenSigninNotAllowed(t *testing.T) {
	signupAndExpect(signupRequest{Email: email1, AllowSignup: true, AllowSignin: false, AuthToken: token1},
		false, false, "Not blocked", t)
}

func TestAddDeviceSucceedsForExistingAccount(t *testing.T) {
	signupAndExpect(signupRequest{Email: email1, AddDevice: true, AllowSignup: false, AllowSignin: false, AuthToken: token1},
		false, true, "Add device should have worked", t)
}

func signupAndExpect(request signupRequest, hasError, succeeds bool, msg string, t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	req, _ := json.Marshal(request)
	resp, err := Signup(ftCtx, string(req))
	if hasError && nil == err {
		t.Fatalf(msg)
	} else if !hasError && nil != err {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	if resp.Success != succeeds {
		t.Fatalf("Success did not match expectations: %s", msg)
	}
}

func TestVerifyFailsWithMissingRequestID(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	_, err := VerifySignup(ftCtx, missingRequestID)
	if nil == err {
		t.Fatal("Should have failed due to not found")
	}
}

func TestVerifySucceedsWithKnownRequestID(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	resp, err := VerifySignup(ftCtx, existingRequestID1)
	if nil != err {
		t.Fatal("Should have succeeded")
	}
	if !resp.Verified {
		t.Fatal("Should have been verified")
	}
}

func TestVerifySucceedsForNewUser(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	resp, err := VerifySignup(ftCtx, newUserRequestID1)
	if nil != err {
		t.Fatal("Should have succeeded")
	}
	if !resp.Verified {
		t.Fatal("Should have been verified")
	}
}

func TestCreatesJWEForFoundToken(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	awsproxy.SetJWEParameters(testPassphrase)
	jweToken, err := AuthenticateUser(ftCtx, email1, token1, false, time.Now)
	require.NoError(t, err)
	require.NotEmpty(t, jweToken)
}

func TestAuthorizesValidNonExpiredJWE(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	awsproxy.SetJWEParameters(testPassphrase)
	auhtResponse, err := AuthenticateUser(ftCtx, email1, token1, false, time.Now)
	if err != nil {
		t.Errorf("Authenticate failed")
	}
	userID, email, err := AuthorizeUser(ftCtx, auhtResponse.AccessToken, time.Now)
	require.NoError(t, err)
	assert.Equal(t, userID1, userID)
	assert.Equal(t, email1, email)
}

func TestAuthorizationFailsForExpiredJWE(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	awsproxy.SetJWEParameters(testPassphrase)
	auhtResponse, err := AuthenticateUser(ftCtx, email1, token1, false, time.Now)
	if err != nil {
		t.Errorf("Authenticate failed")
	}
	timeFn := func() time.Time {
		return time.Now().Add(1 * time.Hour)
	}
	_, _, err = AuthorizeUser(ftCtx, auhtResponse.AccessToken, timeFn)
	if nil == err {
		t.Errorf("Authorization should have failed")
	}
}

func TestHashWorks(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(authTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	hashed, err := hashAuthComponent(ftCtx, token1)
	require.NoError(t, err)
	err = compareHashedAuthComponent(ftCtx, hashed, token1)
	require.NoError(t, err)
}
