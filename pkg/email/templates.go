package email

import (
	"bytes"
	"html/template"
)

// Templates holds pre-compiled email templates
type Templates struct {
	EmailVerification *template.Template
	PasswordReset     *template.Template
	WelcomeEmail      *template.Template
}

// NewTemplates creates and compiles all email templates
func NewTemplates() *Templates {
	return &Templates{
		EmailVerification: template.Must(template.New("email_verification").Parse(emailVerificationTemplate)),
		PasswordReset:     template.Must(template.New("password_reset").Parse(passwordResetTemplate)),
		WelcomeEmail:      template.Must(template.New("welcome").Parse(welcomeEmailTemplate)),
	}
}

// EmailVerificationData holds data for email verification template
type EmailVerificationData struct {
	UserName        string
	VerificationURL string
	ExpiresIn       string
	AppName         string
}

// PasswordResetData holds data for password reset template
type PasswordResetData struct {
	UserName    string
	ResetURL    string
	ExpiresIn   string
	AppName     string
	SupportEmail string
}

// WelcomeEmailData holds data for welcome email template
type WelcomeEmailData struct {
	UserName     string
	AppName      string
	DashboardURL string
}

// RenderEmailVerification renders the email verification template
func (t *Templates) RenderEmailVerification(data EmailVerificationData) (string, error) {
	var buf bytes.Buffer
	if err := t.EmailVerification.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderPasswordReset renders the password reset template
func (t *Templates) RenderPasswordReset(data PasswordResetData) (string, error) {
	var buf bytes.Buffer
	if err := t.PasswordReset.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderWelcomeEmail renders the welcome email template
func (t *Templates) RenderWelcomeEmail(data WelcomeEmailData) (string, error) {
	var buf bytes.Buffer
	if err := t.WelcomeEmail.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const emailVerificationTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #4F46E5; }
        .button { display: inline-block; background: #4F46E5; color: #fff !important; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .button:hover { background: #4338CA; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; text-align: center; }
        .expires { background: #FEF3C7; padding: 10px; border-radius: 4px; font-size: 14px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">{{.AppName}}</div>
        <h2>Verify Your Email Address</h2>
        <p>Hi {{.UserName}},</p>
        <p>Thanks for signing up! Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.VerificationURL}}" class="button">Verify Email</a>
        </p>
        <div class="expires">
            This link will expire in <strong>{{.ExpiresIn}}</strong>.
        </div>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all; font-size: 12px; color: #666;">{{.VerificationURL}}</p>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const passwordResetTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #4F46E5; }
        .button { display: inline-block; background: #DC2626; color: #fff !important; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .button:hover { background: #B91C1C; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; text-align: center; }
        .expires { background: #FEE2E2; padding: 10px; border-radius: 4px; font-size: 14px; margin: 20px 0; }
        .warning { background: #FEF3C7; padding: 15px; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">{{.AppName}}</div>
        <h2>Reset Your Password</h2>
        <p>Hi {{.UserName}},</p>
        <p>We received a request to reset your password. Click the button below to choose a new password:</p>
        <p style="text-align: center;">
            <a href="{{.ResetURL}}" class="button">Reset Password</a>
        </p>
        <div class="expires">
            This link will expire in <strong>{{.ExpiresIn}}</strong>.
        </div>
        <div class="warning">
            <strong>Didn't request this?</strong><br>
            If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
        </div>
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all; font-size: 12px; color: #666;">{{.ResetURL}}</p>
        <div class="footer">
            <p>Need help? Contact us at {{.SupportEmail}}</p>
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

const welcomeEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to {{.AppName}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #4F46E5; }
        .button { display: inline-block; background: #4F46E5; color: #fff !important; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; text-align: center; }
        .features { background: #F3F4F6; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .features ul { margin: 10px 0; padding-left: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">{{.AppName}}</div>
        <h2>Welcome to {{.AppName}}!</h2>
        <p>Hi {{.UserName}},</p>
        <p>Thank you for joining {{.AppName}}! We're excited to have you on board.</p>
        <div class="features">
            <strong>Here's what you can do:</strong>
            <ul>
                <li>Connect your ad accounts (Meta, TikTok, Shopee)</li>
                <li>View unified analytics across all platforms</li>
                <li>Track campaign performance in real-time</li>
                <li>Generate reports and insights</li>
            </ul>
        </div>
        <p style="text-align: center;">
            <a href="{{.DashboardURL}}" class="button">Go to Dashboard</a>
        </p>
        <p>If you have any questions, feel free to reach out to our support team.</p>
        <div class="footer">
            <p>&copy; {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`
