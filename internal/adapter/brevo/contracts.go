package brevo

type sendEmailRequest struct {
	Sender struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"sender"`

	To []struct {
		Email string `json:"email"`
		Name  string `json:"name,omitempty"`
	} `json:"to"`

	Subject     string `json:"subject"`
	HTMLContent string `json:"htmlContent"`
}
