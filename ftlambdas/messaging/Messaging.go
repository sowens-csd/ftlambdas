package messaging

import (
	"strings"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

type socketConnection struct {
	ID           string `json:"id,omitempty" dynamodbav:"id,omitempty"`
	ConnectionID string `json:"connectionId,omitempty" dynamodbav:"connectionId,omitempty"`
	CreatedAt    int64  `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
}

func RecordConnection(ftCtx awsproxy.FTContext, connectionID string) error {
	return ftdb.PutItem(ftCtx, ftdb.ResourceIDFromUserID(ftCtx.UserID), referenceIDFromConnectionID(connectionID),
		socketConnection{ConnectionID: connectionID, ID: ftCtx.UserID, CreatedAt: ftdb.NowMillisecondsSinceEpoch()})
}

func RemoveConnection(ftCtx awsproxy.FTContext, userID, connectionID string) error {
	return ftdb.DeleteItem(ftCtx, ftdb.ResourceIDFromUserID(userID), referenceIDFromConnectionID(connectionID))
}

func LoadUserConnections(ftCtx awsproxy.FTContext, userID string) ([]string, error) {
	ftCtx.RequestLogger.Debug().Str("userId", userID).Msg("Loading user connections")
	connectionIds := make([]string, 0)
	records, err := ftdb.QueryByResource(ftCtx, ftdb.ResourceIDFromUserID(userID))
	if nil != err {
		return connectionIds, err
	}
	for _, record := range records {
		if isConnectionRecord(record) {
			connectionId := connectionIDFromReferenceID(record.ReferenceID)
			connectionIds = append(connectionIds, connectionId)
			ftCtx.RequestLogger.Debug().Str("connectionId", connectionId).Msg("found connection")
		}
	}
	return connectionIds, nil
}

func isConnectionRecord(record ftdb.FolktellsRecord) bool {
	return strings.HasPrefix(record.ReferenceID, "C#")
}

func referenceIDFromConnectionID(connectionID string) string {
	return "C#" + connectionID
}

func connectionIDFromReferenceID(referenceID string) string {
	connectionRunes := []rune(referenceID)
	return string(connectionRunes[2:])
}
