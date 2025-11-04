package domain

type CommonResponse struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}
