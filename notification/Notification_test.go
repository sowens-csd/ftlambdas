package notification

import (
	"strings"
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

func TestNoContentIsAnError(t *testing.T) {
	expectBuildSmsDetailsFails(t, "")
}

func expectBuildSmsDetailsFails(t *testing.T, smsData string) {
	testDB := awsproxy.NewTestDBSvc()
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	_, err := buildSmsDetails(ftCtx, smsData)
	if nil == err {
		t.Errorf("Shoud have error with no content")
	}
}

func expectBuildSmsDetailsSucceeds(t *testing.T, smsData string) *smsDetails {
	testDB := awsproxy.NewTestDBSvc()
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	sms, err := buildSmsDetails(ftCtx, smsData)
	if nil != err {
		t.Errorf("Error %s", err.Error())
	}
	return sms
}

func TestInvalidContentReturnsError(t *testing.T) {
	expectBuildSmsDetailsFails(t, "Body=body1")
	expectBuildSmsDetailsFails(t, "NumMedia=0")
	expectBuildSmsDetailsFails(t, "Body=body1&NumMedia=1")
	expectBuildSmsDetailsFails(t, "junk")
}

func TestContentWithoutMediaWorks(t *testing.T) {

	sms := expectBuildSmsDetailsSucceeds(t, "MessageSid=messageSid1&To=to1&From=from1&Body=body1&NumMedia=0")
	msgContent, err := buildMessageContent(*sms)
	if nil != err {
		t.Errorf("Error %s", err.Error())
	}
	if !strings.Contains(msgContent, "body1") {
		t.Errorf("Expected content missing, was %s", msgContent)
	}
}

func TestContentWithMediaWorks(t *testing.T) {
	mediaURL := "https://example.com/img/img1.png"
	sms := expectBuildSmsDetailsSucceeds(t, "MessageSid=messageSid1&To=to1&From=from1&Body=body1&NumMedia=1&MediaUrl0=https%3A%2F%2Fexample.com%2Fimg%2Fimg1.png&MediaContentType0=img/png")
	msgContent, err := buildMessageContent(*sms)
	if nil != err {
		t.Errorf("Error %s", err.Error())
	}
	if !strings.Contains(msgContent, mediaURL) {
		t.Errorf("Expected content missing, was %s", msgContent)
	}
}
