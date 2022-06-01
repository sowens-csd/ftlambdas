package ftlambdas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftauth"
	"github.com/sowens-csd/ftlambdas/ftdb"
	"github.com/sowens-csd/ftlambdas/notification"
	"github.com/sowens-csd/ftlambdas/sharing"
)

type managedUserCreationRequest struct {
	Name      string `json:"name" dynamodbav:"name"`
	CreatedBy string `json:"createdBy" dynamodbav:"createdBy"`
}

func AddManagedUser(ftCtx awsproxy.FTContext, managedUserJSON string) error {
	inputJSON := []byte(managedUserJSON)
	var userRequest managedUserCreationRequest
	err := json.Unmarshal(inputJSON, &userRequest)
	if nil != err {
		ftCtx.RequestLogger.Debug().Err(err).Msg("Error unmarshalling request")
		return err
	}
	managedUser, err := sharing.AddManagedUser(ftCtx, userRequest.Name, userRequest.CreatedBy)
	if nil != err {
		return err
	}
	return ftauth.AuthorizeManagedUser(ftCtx, managedUser.ID, managedUser.CreatedBy)
}

func UpdateOrDeleteUser(ftCtx awsproxy.FTContext, onlineUserJSON string) error {
	inputJSON := []byte(onlineUserJSON)
	var onlineUser sharing.OnlineUser
	err := json.Unmarshal(inputJSON, &onlineUser)
	if nil != err {
		return err
	}
	ou, err := sharing.LoadOnlineUser(ftCtx, ftCtx.UserID)
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg(("can't load current user"))
		return err
	}
	if ou.ID != onlineUser.ID {
		ftCtx.RequestLogger.Info().Str("received", onlineUser.ID).Str("auth", ou.ID).Msg(("user ID mismatch"))
		return fmt.Errorf("Wrong user ID %s", onlineUser.ID)
	}
	if onlineUser.Deleted == ftdb.DeleteRemove {
		sharing.DeleteStoriesForUser(ftCtx)
		sharing.DeleteGroupsForUser(ftCtx)
		ftauth.DeleteAuthentication(ftCtx, ou.Email)
		ftdb.DeleteItem(ftCtx, ftdb.ResourceIDFromTransactionID(ou.OriginalTransactionID), ftdb.ReferenceIDFromUserID(ou.ID))
		ou.Delete(ftCtx)
		return nil
	}
	if ou.Name != onlineUser.Name {
		sharing.UpdateMemberName(ftCtx, onlineUser.Name)
	}
	return sharing.UpdateOnlineUser(ftCtx, onlineUserJSON)
}

// UpdateStoryAndNotify updates the DB with the new state of the shared story and
// nofifies affected users.
func UpdateStoryAndNotify(ftCtx awsproxy.FTContext, storyShareJSON string, client *http.Client) (sharing.StoryUpdateResult, error) {
	inputJSON := []byte(storyShareJSON)
	var sharedStory sharing.SharedStory
	err := json.Unmarshal(inputJSON, &sharedStory)
	if nil != err {
		return sharing.StoryUpdateResult{}, err
	}
	result, err := sharing.UpdateSharedStory(ftCtx, sharedStory)
	if result.Success && nil != sharedStory.Groups {
		notifyAffectedUsers(ftCtx, sharedStory, client)
	}
	return result, err
}

func notifyAffectedUsers(ftCtx awsproxy.FTContext, sharedStory sharing.SharedStory, client *http.Client) {
	for _, group := range sharedStory.Groups {
		ftCtx.RequestLogger.Debug().Str("groupID", group.GroupID).Msg("Getting group users")
		users, err := sharing.FindOnlineUsersForGroup(ftCtx, group.GroupID)
		if nil == err {
			ftCtx.RequestLogger.Debug().Int("users", len(users)).Msg("found users")
			for _, user := range users {
				ftCtx.RequestLogger.Debug().Str("userID", user.ID).Msg("Checking user")
				if user.ID != ftCtx.UserID {
					ftCtx.RequestLogger.Debug().Str("userID", user.ID).Msg("Notifying user")
					notification.SendStoryChangeCommand(ftCtx, sharedStory.StoryID, sharedStory.Title, sharedStory.LastUpdatedBy, user, client)
					ftCtx.RequestLogger.Debug().Str("userID", user.ID).Msg("Notified user")
				}
			}
		} else {
			ftCtx.RequestLogger.Debug().Err(err).Msg("Error loading users")
		}
	}

}
