package notification

import "testing"

func TestExtractsArnFromError(t *testing.T) {
	foundArn := findArnInMessage("Junk")
	if "" != foundArn {
		t.Errorf("Shoud have found nothing, actually found %s", foundArn)
	}
}

func TestShouldReturnKnownArn(t *testing.T) {
	foundArn := findArnInMessage("Could not create Endpoint arn:aws:sns:folktells already exists " +
		"with the same token so no create")
	if "arn:aws:sns:folktells" != foundArn {
		t.Errorf("Shoud have found expected ARN, actually found %s", foundArn)
	}
}
