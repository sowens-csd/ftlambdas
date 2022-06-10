package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/notification"
	"github.com/sowens-csd/ftlambdas/sharing"
)

// Handler is responsible for taking a signup request from a client that contains
// a new or changed user token and confirming the request.
//
// Client devices generate a token then use it to create an authentication record for
// the external P2P service to use againt the p2p_authorizer endpoint. There is always
// only one token active for a given client at a time.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	emailEnc, found := request.PathParameters["email"]
	if found {
		email, err := url.QueryUnescape(emailEnc)
		if err != nil {
			return awsproxy.HandleError(fmt.Errorf("Bad email encoding"), ftCtx.RequestLogger), nil
		}
		ou, err := sharing.LoadOnlineUserByEmail(ftCtx, email)
		if err != nil {
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		ftCtx.RequestLogger.Debug().Msg("Found the user")
		callItParam, found := request.QueryStringParameters["call"]
		if found {
			ftCtx.RequestLogger.Debug().Msg("Call parameter found")
			callIt, err := strconv.ParseBool(callItParam)
			if callIt && nil == err {
				ftCtx.RequestLogger.Debug().Msg("Call parameter bool")
				title := "Folktels Call"
				callType, found := request.QueryStringParameters["calltype"]
				if found {
					title = fmt.Sprintf("Folktells %s Call", callType)
				}
				callerName := email
				nameParam, found := request.QueryStringParameters["caller"]
				if found {
					callerName = nameParam
				}
				ftCtx.RequestLogger.Debug().Msg("sending alert")
				notification.SendAlert(ftCtx, title, fmt.Sprintf("Call from %s", callerName), ou, &http.Client{Timeout: 30 * time.Second})
			} else {
				ftCtx.RequestLogger.Debug().Str("email", email).Bool("callIt", callIt).Msg("not valid request")
			}
		} else {
			ftCtx.RequestLogger.Debug().Msg("Call parameter not found")
		}
		return awsproxy.NewTextResponse(ftCtx, ou.CallPeerId), nil
	}
	return awsproxy.HandleError(fmt.Errorf("Email path parameter missing"), ftCtx.RequestLogger), nil

}

func main() {
	lambda.Start(Handler)
}
