package email

import (
	"time"
)

type NotificationType string

const (
	Templated NotificationType = "templated"
	Plain     NotificationType = "plain"
)

type EmailMetadata struct {
	ReplyTo    *string `json:"reply_to,omitempty"`
	SenderName *string `json:"sender_name,omitempty"`
}

// func (e *EmailMetadata) GetEmailOptions() []EmailOpt {
// 	var opts []EmailOpt
// 	if e.ReplyTo != nil {
// 		opts = append(opts, WithReplyTo(*e.ReplyTo))
// 	}
// 	if e.SenderName != nil {
// 		opts = append(opts, WithSenderName(*e.SenderName))
// 	}

// 	return opts
// }

type ContactInfo struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

type NotificationMetadata struct {
	Source string `json:"source"`

	// Sender name will show up as the sender of the notification
	EmailMetadata     *EmailMetadata   `json:"email_metadata,omitempty"`
	CreationTimestamp time.Time        `json:"creation_timestamp"`
	DeliveryMethods   []string         `json:"delivery_methods"`
	Type              NotificationType `json:"type"`
}

type TemplatedNotificationDto struct {
	Metadata NotificationMetadata `json:"metadata"`

	Subject    string `json:"subject"`
	TemplateId string `json:"template_id"`

	// data will be passed into the template engine
	Data map[string]interface{} `json:"data"`

	ContactInfo ContactInfo `json:"contact_info"`
}

type PlainNotificationDto struct {
	Metadata NotificationMetadata `json:"metadata"`

	Subject     string               `json:"subject"`
	Body        string               `json:"body"`
	ContactInfo ContactInfo `json:"contact_info"`
}
