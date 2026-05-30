package dto

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok"`
	}
}
