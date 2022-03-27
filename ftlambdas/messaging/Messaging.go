package messaging

import (
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

type socketConnection struct {
	userID       string
	connectionID string
	openedAt     int64
}

func RecordConnection(ftCtx awsproxy.FTContext, connectionID string) error {
	return ftdb.PutItem(ftCtx, ftdb.ResourceIDFromUserID(ftCtx.UserID), referenceIDFromConnectionID(connectionID),
		&socketConnection{connectionID: connectionID, userID: ftCtx.UserID, openedAt: ftdb.NowMillisecondsSinceEpoch()})
}

func RemoveConnection(ftCtx awsproxy.FTContext, connectionID string) error {
	return ftdb.DeleteItem(ftCtx, ftdb.ResourceIDFromUserID(ftCtx.UserID), referenceIDFromConnectionID(connectionID))
}

func referenceIDFromConnectionID(connectionID string) string {
	return "C#" + connectionID
}
