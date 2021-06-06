package notification

import (
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
