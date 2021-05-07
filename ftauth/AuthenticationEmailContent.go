package ftauth

import (
	"fmt"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

// EnglishAuthCodeContent the english content for the new user welcome email
func EnglishAuthCodeContent(authCode string) awsproxy.EmailContent {
	return awsproxy.EmailContent{
		Subject: "Folktells Account Verification",
		TextBody: fmt.Sprintf(`Your Folktells verification code is: 
%s

To protect the security of your account, DO NOT give this code to anyone else. If you did not make this request, delete this email. This code will expire in 2 hours.

Have questions? Email us at hello@folktells.com. We're here to help!

The Folktells Support team
`, authCode),
	}
}
