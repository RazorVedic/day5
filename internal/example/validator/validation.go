package validator

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// ValidateCreatePaymentRequest validates the CreatePaymentRequest
func ValidateCreatePaymentRequest(
	ctx context.Context, req *rpc.CreatePaymentRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.ReferenceId, validation.Required),
		validation.Field(&req.Amount, validation.Required, validation.Min(1)),
		validation.Field(&req.Currency, validation.Required),
		validation.Field(&req.Description, validation.Required),
		validation.Field(&req.Payer, validation.Required),
		validation.Field(&req.Payee, validation.Required),
	)
}

// ValidateGetPaymentRequest validates the GetPaymentRequest
func ValidateGetPaymentRequest(
	ctx context.Context, req *rpc.GetPaymentRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required),
	)
}

// ValidateUpdatePaymentRequest validates the UpdatePaymentRequest
func ValidateUpdatePaymentRequest(
	ctx context.Context, req *rpc.UpdatePaymentRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required),
		validation.Field(&req.ReferenceId, validation.Required),
		validation.Field(&req.Amount, validation.Required, validation.Min(1)),
		validation.Field(&req.Currency, validation.Required),
		validation.Field(&req.Description, validation.Required),
		validation.Field(&req.Status, validation.Required),
		validation.Field(&req.Payer, validation.Required),
		validation.Field(&req.Payee, validation.Required),
	)
}

// ValidateDeletePaymentRequest validates the DeletePaymentRequest
func ValidateDeletePaymentRequest(
	ctx context.Context, req *rpc.DeletePaymentRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required),
	)
}

// ValidateListPaymentsRequest validates the ListPaymentsRequest
func ValidateListPaymentsRequest(
	_ context.Context, req *rpc.ListPaymentsRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.ReferenceId, validation.Required),
		validation.Field(&req.Count, validation.Required, validation.Min(1)),
	)
}
