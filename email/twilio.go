package email

import (
	"context"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

// func GetTwilioSMS(ident TwilioSMSIdentity) (SMSClient, error) {
// 	twilioSid := os.Getenv("TWILIO_ACCOUNT_SID")
// 	twilioToken := os.Getenv("TWILIO_AUTH_TOKEN")

// 	if twilioSid == "" || twilioToken == "" {
// 		return nil, fmt.Errorf(
// 			"please set TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN environment variables before using twilio SMS",
// 		)
// 	}

// 	return newTwilioSMS(ident), nil
// }

// func newTwilioSMS(identity TwilioSMSIdentity) SMSClient {
// 	return &TwilioSMS{
// 		Origin: identity,
// 		client: twilio.NewRestClient(),
// 		l:      zap.L(),
// 	}
// }

// type TwilioSMS struct {
// 	Origin TwilioSMSIdentity
// 	client *twilio.RestClient
// 	l      *zap.Logger
// }

// func (cli *TwilioSMS) SendSMS(ctx context.Context, phoneNumber string, message string) error {
// 	l := cli.l.With(zap.String("recepient", censorString(phoneNumber)))

// 	params := &twilioApi.CreateMessageParams{}
// 	params.SetFrom(cli.Origin.MessageServiceId)
// 	params.SetTo(phoneNumber)
// 	params.SetBody(message)
// 	resp, err := cli.client.Api.CreateMessage(params)
// 	if err != nil {
// 		return fmt.Errorf("failed to create message (code: %v) (error: %v): %w", resp.ErrorCode, resp.ErrorMessage, err)
// 	} else {
// 		l.Info("sent SMS", zap.Stringp("status", resp.Status))
// 	}
// 	return nil
// }

// func (cli *TwilioSMS) SendTemplatedSMS(ctx context.Context, phoneNumber string, template Template) error {
// 	l := cli.l.With(zap.String("recepient", censorString(phoneNumber)), zap.String("template_id", template.Id))

// 	b, err := json.Marshal(template.Data)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal template data: %w", err)
// 	}

// 	templateVars := string(b)
// 	l.Debug("sending templated sms",
// 		zap.Int("templateVars_length", len(templateVars)))
// 	params := &twilioApi.CreateMessageParams{}
// 	params.SetFrom(cli.Origin.MessageServiceId)
// 	params.SetTo(phoneNumber)
// 	params.SetContentSid(template.Id)
// 	params.SetContentVariables(templateVars)
// 	// params.SetBody("Hello, {{1}}, this is an example of a {{2}} message, as of {{3}}")
// 	resp, err := cli.client.Api.CreateMessage(params)
// 	if err != nil {
// 		return fmt.Errorf("failed to send templated sms message: %w", err)
// 	} else {
// 		l.Info("sent templated SMS", zap.Stringp("status", resp.Status))
// 	}

// 	return nil
// }

func NewTwilioEmail(apiKey string, identity TwilioEmailIdentity) EmailClient {
	return &TwilioEmail{
		Origin:   identity,
		sgClient: sendgrid.NewSendClient(apiKey),
		l: zap.L().With(
			zap.String("sender_email", identity.SenderEmail),
			zap.String("sender_name", identity.SenderName),
		),
	}
}

type TwilioEmail struct {
	Origin   TwilioEmailIdentity
	sgClient *sendgrid.Client
	l        *zap.Logger
}

func applyEmailOption(m *mail.SGMailV3, o EmailOpt) {
	switch o.key {
	case SenderNameOpt:
		m.From.Name = o.val.(string)
	case ReplyToOpt:
		m.ReplyTo = mail.NewEmail("", o.val.(string))
	}
}

func (cli *TwilioEmail) SendTemplatedEmail(
	ctx context.Context,
	recep EmailRecepient,
	subject string,
	template Template,
	opts ...EmailOpt) error {
	l := cli.l.With(
		zap.String("recepient_email", recep.Email),
		zap.String("recepient_name", recep.Name),
		zap.String("subject", subject),
		zap.String("template_id", template.Id),
		zap.Int("data_size", len(template.Data)))
	from := mail.NewEmail(cli.Origin.SenderName, cli.Origin.SenderEmail)
	to := mail.NewEmail(recep.Name, recep.Email)

	m := mail.NewV3Mail()
	m.SetTemplateID(template.Id)
	m.SetFrom(from)

	// personalizations allow to send multiple emails with customer specific data
	data := mail.NewPersonalization()
	data.From = from
	data.To = []*mail.Email{to}
	data.Subject = subject
	data.DynamicTemplateData = template.Data

	m.AddPersonalizations(data)

	// apply all the options after the email has been built
	for _, o := range opts {
		l.Debug("applying email option", zap.String("key", o.key), zap.Any("value", o.val))
		applyEmailOption(m, o)
	}

	l.Info("sending templated email")
	resp, err := cli.sgClient.SendWithContext(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send templated email: %w", err)
	} else {
		l.Info("email sent", zap.String("response", resp.Body), zap.Int("status", resp.StatusCode))
	}

	return nil
}

func (cli *TwilioEmail) SendHTMLEmail(ctx context.Context, recep EmailRecepient, header, htmlContent string, opts ...EmailOpt) error {
	l := cli.l.With(
		zap.Int("html_size", len(htmlContent)),
		zap.String("recepient_email", recep.Email),
		zap.String("recepient_name", recep.Name),
	)
	from := mail.NewEmail(cli.Origin.SenderName, cli.Origin.SenderEmail)
	to := mail.NewEmail(recep.Name, recep.Email)
	m := mail.NewSingleEmail(from, header, to, "", htmlContent)

	// apply all the options after the email has been built
	for _, o := range opts {
		l.Debug("applying email option", zap.String("key", o.key), zap.Any("value", o.val))
		applyEmailOption(m, o)
	}

	l.Debug("sending email")
	resp, err := cli.sgClient.SendWithContext(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	l.Info("email sent", zap.String("response", resp.Body), zap.Int("status", resp.StatusCode))

	return nil
}
