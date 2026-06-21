package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Service struct {
	apiKey string
	from   string
	name   string
	host   string
	client *http.Client
}

func New(apiKey, from, name, host string) *Service {
	return &Service{
		apiKey: apiKey,
		from:   from,
		name:   name,
		host:   host,
		client: &http.Client{},
	}
}

func (s *Service) Send(to string, subject string, html string) error {
	reqBody := sendEmailRequest{
		Subject:     subject,
		HTMLContent: html,
	}

	reqBody.Sender.Name = s.name
	reqBody.Sender.Email = s.from

	reqBody.To = append(reqBody.To, struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	}{
		Email: to,
	})

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		s.host,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("brevo returned status %d", resp.StatusCode)
	}

	return nil
}
