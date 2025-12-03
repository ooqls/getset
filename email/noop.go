package email

import "context"

var _ EmailClient = &NoopEmailClient{}

type NoopEmailClient struct {
}

func (cli *NoopEmailClient) SendTemplatedEmail(ctx context.Context, recep EmailRecepient, subject string, template Template, opts ...EmailOpt) error {
	return nil
}

func (cli *NoopEmailClient) SendHTMLEmail(ctx context.Context, recep EmailRecepient, header, htmlContent string, opts ...EmailOpt) error {
	return nil
}
