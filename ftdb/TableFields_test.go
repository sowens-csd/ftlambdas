package ftdb

import (
	"strings"
	"testing"
)

func TestResourceIDForAuthRequest(t *testing.T) {
	resID := ResourceIDForAuthRequest()
	if !strings.HasPrefix(resID, "AR#20") {
		t.Errorf("Was not as expected %s", resID)
	}
}
