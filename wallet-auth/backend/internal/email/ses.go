//nolint:staticcheck // aws-sdk-go v1 until SES migrates to aws-sdk-go-v2.
package email

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// EmailService handles sending emails via AWS SES
type EmailService struct {
	sesClient *ses.SES
	fromEmail string
	baseURL   string
}

// NewEmailService creates a new email service instance
func NewEmailService(region, accessKeyID, secretAccessKey, fromEmail, baseURL string) (*EmailService, error) {
	config := &aws.Config{
		Region: aws.String(region),
	}

	if accessKeyID != "" && secretAccessKey != "" {
		config.Credentials = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &EmailService{
		sesClient: ses.New(sess),
		fromEmail: fromEmail,
		baseURL:   baseURL,
	}, nil
}

// SendVerificationEmail sends a magic link email for email signup/login
func (e *EmailService) SendVerificationEmail(toEmail, token string) error {
	verificationURL := fmt.Sprintf("%s/email-verify?token=%s", e.baseURL, token)

	subject := "Verify your email address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Verify your email address</h2>
			<p>Click the link below to verify your email address and sign in:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 24 hours.</p>
		</body>
		</html>
	`, verificationURL, verificationURL)

	return e.sendEmail(toEmail, subject, body)
}

// SendWalletVerificationEmail sends an email with SIWE message for wallet verification
func (e *EmailService) SendWalletVerificationEmail(toEmail, address, message string) error {
	subject := "Verify your wallet address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Verify your wallet address</h2>
			<p>You are adding wallet address %s to your account.</p>
			<p>Please sign the following message with your wallet:</p>
			<pre style="background-color: #f5f5f5; padding: 10px; border-radius: 5px;">%s</pre>
			<p>This verification is required to link your wallet to your account.</p>
		</body>
		</html>
	`, address, message)

	return e.sendEmail(toEmail, subject, body)
}

// sendEmail is a helper method to send emails via SES
func (e *EmailService) sendEmail(toEmail, subject, bodyHTML string) error {
	// Create email template
	emailTemplate := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { margin-top: 30px; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				{{.Body}}
				<div class="footer">
					<p>If you did not request this email, please ignore it.</p>
				</div>
			</div>
		</body>
		</html>
	`

	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]interface{}{
		"Body": template.HTML(bodyHTML),
	}); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Send email via SES
	input := &ses.SendEmailInput{
		Source: aws.String(e.fromEmail),
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(toEmail)},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Data:    aws.String(buf.String()),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	_, err = e.sesClient.SendEmail(input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	return nil
}

// GenerateVerificationToken generates a secure token for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetVerificationExpiry returns the expiration time for verification tokens (24 hours)
func GetVerificationExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}
