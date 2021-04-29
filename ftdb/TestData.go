package ftdb

import "github.com/sowens-csd/ftlambdas/awsproxy"

const (
	userID1       = "userID1"
	userID2       = "userID2"
	resID1        = "resID1"
	resID2        = "resID2"
	missingResID1 = "missingRes1"
	refID1        = "refID1"
	refID2        = "refID2"
)

var testDBData = awsproxy.TestDBData{
	awsproxy.TestDBDataRecord{
		ResourceID:  resID1,
		ReferenceID: refID1,
		Record: map[string]interface{}{
			ResourceIDField:  resID1,
			ReferenceIDField: refID1,
			IDField:          userID1,
		},
	},
}
