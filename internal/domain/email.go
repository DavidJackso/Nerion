package domain

type EmailSender interface {
	Send(to, subject, body string) error
}

type NoopEmailSender struct{}

func (NoopEmailSender) Send(to, subject, body string) error { return nil }
