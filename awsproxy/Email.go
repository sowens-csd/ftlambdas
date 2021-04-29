package awsproxy

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sest "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type emailSender interface {
	SendEmail(ctx context.Context, sendTo string, emailContent EmailContent, requestLogger zerolog.Logger)
}

type sesEmailSender struct {
	emailSender
}

// TestEmailSender is useful for test cases that want to mock emails
type TestEmailSender struct {
	emailSender
	lastSendTo       string
	lastEmailContent EmailContent
}

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	sender = "Folktells <no-reply@folktells.com>"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	subject = "Folktells Invitation"

	// The HTML body for the email.
	htmlBody = "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"

	//The email body for recipients with non-HTML email clients.
	textBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	charSet = "UTF-8"
)

// EmailContent describes the various pieces of content required for an
// email. All of the components can be templates which will be populated
// with the values given in the Send command.
type EmailContent struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// SendSESEmail send a new email to a known recipient using Amazon SES
func SendSESEmail(ctx context.Context, sendTo string, emailContent EmailContent, requestLogger zerolog.Logger) {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}

	// Create an SES session.
	svc := ses.NewFromConfig(cfg)

	var body sest.Body
	if len(emailContent.HTMLBody) > 0 {
		body = sest.Body{
			Html: &sest.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(emailContent.HTMLBody),
			},
			Text: &sest.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(emailContent.TextBody),
			},
		}
	} else {
		body = sest.Body{
			Text: &sest.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(emailContent.TextBody),
			},
		}

	}
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &sest.Destination{
			CcAddresses: []string{},
			ToAddresses: []string{
				sendTo,
			},
		},
		Message: &sest.Message{
			Body: &body,
			Subject: &sest.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(emailContent.Subject),
			},
		},
		Source: aws.String(sender),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(ctx, input)

	// Display error messages if they occur.
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			requestLogger.Error().Msg(fmt.Sprintf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap()))
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			requestLogger.Error().Err(err).Msg("could not sent email")
		}
		return
	}

	fmt.Println("Email Sent to address: " + sendTo)
	fmt.Println(result)
}

// SendEmail sends the given content out to the given address
func (sender *sesEmailSender) SendEmail(ctx context.Context, sendTo string, emailContent EmailContent, requestLogger zerolog.Logger) {
	SendSESEmail(ctx, sendTo, emailContent, requestLogger)
}

// SendEmail sends the given content out to the given address
func (sender *TestEmailSender) SendEmail(ctx context.Context, sendTo string, emailContent EmailContent, requestLogger zerolog.Logger) {
	sender.lastSendTo = sendTo
	sender.lastEmailContent = EmailContent{
		Subject:  emailContent.Subject,
		HTMLBody: emailContent.HTMLBody,
		TextBody: emailContent.TextBody,
	}
}

// ExpectLastSendTo SendEmail sends the given content out to the given address
func (sender *TestEmailSender) ExpectLastSendTo(expectedSendTo string, t *testing.T) {
	if sender.lastSendTo != expectedSendTo {
		t.Errorf("Email send to expected %s was %s", expectedSendTo, sender.lastSendTo)
	}
}

// ExpectLastEmailContent SendEmail sends the given content out to the given address
func (sender *TestEmailSender) ExpectLastEmailContent(expectedContent EmailContent, t *testing.T) {
	if sender.lastEmailContent.Subject != expectedContent.Subject {
		t.Errorf("Email subject expected %s was %s", expectedContent.Subject, sender.lastEmailContent.Subject)
	}
}
