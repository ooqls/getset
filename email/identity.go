package email

type TwilioSMSIdentity struct {
	MessageServiceId string `yaml:"messageServiceId" json:"messageServiceId"`
}

type TwilioEmailIdentity struct {
	// SenderName is the name that will be used as display name for the email
	SenderName string `yaml:"emailName" json:"emailName"`

	// if SenderName is not provided, then email will be used as the display name
	SenderEmail string `yaml:"email" json:"email"`
}
