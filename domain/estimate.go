package domain

type EstimateRequest struct {
	Pool      string `json:"pool" validate:"required"`
	Src       string `json:"src" validate:"required"`
	Dst       string `json:"dst" validate:"required"`
	SrcAmount string `json:"src_amount" validate:"required"`
}

type EstimateResponse struct {
	DstAmount string `json:"dst_amount"`
}
