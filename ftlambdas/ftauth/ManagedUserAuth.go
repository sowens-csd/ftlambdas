package ftauth

import (
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

type managedUserAuthorization struct {
	UserID    string `json:"userID" dynamodbav:"userID"`
	CreatedBy string `json:"createdBy" dynamodbav:"createdBy"`
	CreatedAt int    `json:"createdAt" dynamodbav:"createdAt"`
}

// AuthorizeManagedUser creates a temporary authentication record that can be used once by a
// managed user.
//
// A ManagedUser is one that is not created by the individual using the account, for example
// a retirement home using the system for its residents. These users need not have an
// email address so they can't receive a one time authentication token and use that to
// authenticate the first time. Instead when their account is created a single use
// authentication record is created with this call. The first time the users signs in with
// their username the existence of this record automatically authenticates them and is then
// deleted.
// One difference with these requests is that the auth token is not known when the single use
// authentication is created. The auth token will be provided when the user logs in for
// the first time using the single use authentication grant. After that happens these users
// can authenticate in the same way as a non-managed user by providing their authentication
// token.
func AuthorizeManagedUser(ftCtx awsproxy.FTContext, userID, createdBy string) error {
	resourceID := ftdb.ResourceIDForManagedUserAuth(userID)
	auth := managedUserAuthorization{
		UserID:    userID,
		CreatedBy: createdBy,
		CreatedAt: ftdb.NowMillisecondsSinceEpoch(),
	}
	err := ftdb.PutItem(ftCtx, resourceID, resourceID, auth)
	if nil != err {
		ftCtx.RequestLogger.Debug().Err(err).Msg("failed to create managed user authorization")
	}
	return err
}
