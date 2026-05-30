package humaapi

import "net/http"

type envelopeBody struct {
	IsSuccess bool   `json:"isSuccess"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
}

func OKBody(data any) envelopeBody {
	return envelopeBody{
		IsSuccess: true,
		Message:   "OK",
		Data:      data,
	}
}

type OKOutput struct {
	Body envelopeBody
}

type CreatedOutput struct {
	Status int `status:"201"`
	Body   envelopeBody
}

func OK(data any) *OKOutput {
	return &OKOutput{Body: OKBody(data)}
}

func Created(data any) *CreatedOutput {
	return &CreatedOutput{
		Status: http.StatusCreated,
		Body:   OKBody(data),
	}
}
