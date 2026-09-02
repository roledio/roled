package payloads

type EmailPayload struct {
	Type            string  `json:"type"`
	To              string  `json:"to"`
	From            string  `json:"from"`
	Subject         string  `json:"subject"`
	AccountName     string  `json:"account_name"`
	ProjectName     string  `json:"project_name"`
	ProjectLogoURL  *string `json:"project_logo_url"`
	ProjectIsSystem bool    `json:"project_is_system"`
	DisplayName     string  `json:"display_name"`
	UserID          string  `json:"user_id"`
	LoginURL        *string `json:"login_url"`
	IsSignup        bool    `json:"is_signup"`
	Token           string  `json:"token"`
}
