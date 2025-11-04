package domain

type ErrorResponse struct {
	Error       string `json:"error"`
	Code        int    `json:"code"`
	Description string `json:"description"`
}
