package email

import (
	"log/slog"

	"go-template/internal/domain"
)

type Sender struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) domain.EmailSender {
	return &Sender{logger: logger}
}

func (s *Sender) Send(to, subject, body string) error {
	// TODO: заменить на реальный SMTP/SendGrid/etc
	s.logger.Info("email stub", "to", to, "subject", subject)
	return nil
}
