package email

import "context"

type Template struct {
	Id   string                 `yaml:"id" json:"id"`
	Data map[string]interface{} `yaml:"data" json:"data"`
}

type EmailRecepient struct {
	// if a name is provided, then it will be used as the display name for the email
	Name string

	// otherwise email will be used as the display name
	Email string
}

type EmailOpt struct {
	key string
	val interface{}
}

func WithSenderName(name string) EmailOpt {
	return EmailOpt{key: "sender_name", val: name}
}

func WithReplyTo(email string) EmailOpt {
	return EmailOpt{key: "reply_to", val: email}
}

const (
	SenderNameOpt = "sender_name"
	ReplyToOpt    = "reply_to"
)

// type SMSClient interface {
// SendSMS(ctx context.Context, phoneNumber string, message string) error
// SendTemplatedSMS(ctx context.Context, phoneNumber string, template Template) error
// }

type EmailClient interface {
	SendTemplatedEmail(ctx context.Context, recep EmailRecepient, subject string, template Template, o ...EmailOpt) error
	SendHTMLEmail(ctx context.Context, recep EmailRecepient, subject, body string, o ...EmailOpt) error
}

// type PushNotificationClient interface {
// SendPushNotification(ctx context.Context, subject, message string) error
// }
