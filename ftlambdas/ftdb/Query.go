package ftdb

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
)

// QueryResponse the results of performing a query
type QueryResponse struct {
	Results []FolktellsRecord `json:"results,omitempty"`
}

// QueryRequest a set of parameters that drive a particular query
type QueryRequest struct {
	ResourceID  string `json:"resourceId,omitempty"`
	ReferenceID string `json:"referenceId,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
	UserID      string `json:"userId,omitempty"`
	Email       string `json:"email,omitempty"`
}

// QueryFromRequest execute a query against the DB based on the parameters in a QueryRequest
func QueryFromRequest(ftCtx awsproxy.FTContext, queryReqJSON string) (QueryResponse, error) {
	emptyResp := QueryResponse{}
	var req QueryRequest
	err := json.Unmarshal([]byte(queryReqJSON), &req)
	if err != nil {
		return emptyResp, err
	}
	if len(req.ResourceID) > 0 && len(req.ReferenceID) > 0 {
		return findSingleRecord(ftCtx, req.ResourceID, req.ReferenceID)
	} else if len(req.Email) > 0 {
		return findByEmail(ftCtx, req.Email)
	} else if len(req.UserID) > 0 {
		return findByUser(ftCtx, req.UserID)
	} else if len(req.GroupID) > 0 {
		return findByGroup(ftCtx, req.GroupID)
	}
	return emptyResp, nil
}

func findSingleRecord(ftCtx awsproxy.FTContext, resID, refID string) (QueryResponse, error) {
	emptyResp := QueryResponse{}
	ftCtx.RequestLogger.Debug().Str("resID", resID).Str("refID", refID).Msg("Query single record")
	var records []FolktellsRecord
	found, records := appendSingleRecord(ftCtx, records, resID, refID)
	if found {
		ftCtx.RequestLogger.Debug().Str("resID", resID).Str("refID", refID).Msg("Found it")
		result := QueryResponse{Results: records}
		return result, nil
	}
	return emptyResp, nil
}

func findByEmail(ftCtx awsproxy.FTContext, email string) (QueryResponse, error) {
	ftCtx.RequestLogger.Debug().Str("email", email).Msg("Query by email")
	results, err := Query(ftCtx, EmailIndex, "email = :em1 ", map[string]types.AttributeValue{
		":em1": &types.AttributeValueMemberS{Value: email},
	})
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("findByEmail failed")
	} else {
		ftCtx.RequestLogger.Debug().Int("results", len(results)).Msg("findByEmail found")
	}
	return QueryResponse{Results: results}, err
}

func findByUser(ftCtx awsproxy.FTContext, userID string) (QueryResponse, error) {
	ftCtx.RequestLogger.Debug().Str("userID", userID).Msg("Query by user")
	emptyResp := QueryResponse{}
	var records []FolktellsRecord
	found, records := appendSingleRecord(ftCtx, records, ResourceIDFromUserID(userID), ReferenceIDFromUserID(userID))
	if found {
		ftCtx.RequestLogger.Debug().Str("userID", userID).Msg("Found it")
		results, err := Query(ftCtx, UserToGroupIndex, "memberId = :uid1 ", map[string]types.AttributeValue{
			":uid1": &types.AttributeValueMemberS{Value: userID},
		})
		if nil == err {
			ftCtx.RequestLogger.Debug().Int("results", len(results)).Msg("member query successful")
			uniqueGroups := make(map[string]bool)
			for _, singleRecord := range results {
				found, records = appendSingleRecord(ftCtx, records, singleRecord.ResourceID, singleRecord.ReferenceID)
				uniqueGroups[singleRecord.ResourceID] = true
			}
			for groupRes := range uniqueGroups {
				found, records = appendSingleRecord(ftCtx, records, groupRes, groupRes)
			}
		} else {
			ftCtx.RequestLogger.Debug().Err(err).Msg("Error querying for members")
		}
		ftCtx.RequestLogger.Debug().Int("results", len(results)).Msg("user total results")
		return QueryResponse{Results: records}, err
	}
	return emptyResp, nil
}

func findByGroup(ftCtx awsproxy.FTContext, groupID string) (QueryResponse, error) {
	ftCtx.RequestLogger.Debug().Str("groupID", groupID).Msg("Query by group")
	emptyResp := QueryResponse{}
	var records []FolktellsRecord
	groupResourceID := ResourceIDFromGroupID(groupID)
	found, records := appendSingleRecord(ftCtx, records, groupResourceID, groupResourceID)
	if found {
		ftCtx.RequestLogger.Debug().Str("groupID", groupID).Msg("Found it")
	}
	groupRefs, err := QueryByResource(ftCtx, groupResourceID)
	if nil == err {
		for _, singleRecord := range groupRefs {
			if singleRecord.ResourceID != groupResourceID || singleRecord.ReferenceID != groupResourceID {
				_, records = appendSingleRecord(ftCtx, records, singleRecord.ResourceID, singleRecord.ReferenceID)
			}
		}
	}
	return QueryResponse{Results: records}, nil
	return emptyResp, nil
}

func appendSingleRecord(ftCtx awsproxy.FTContext, records []FolktellsRecord, resourceID, referenceID string) (bool, []FolktellsRecord) {
	var ftRec FolktellsRecord
	found, err := GetItem(ftCtx, resourceID, referenceID, &ftRec)
	if nil == err && found {
		return true, append(records, ftRec)
	}
	return false, records
}
