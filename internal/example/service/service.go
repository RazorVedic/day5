package service

import (
	"context"

	"github.com/razorpay/goutils/sqlstorage/query"

	"github.com/razorpay/go-foundation-v2/internal/example/model"
	"github.com/razorpay/go-foundation-v2/internal/example/service/builder"
	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// Service implements the Payment Service
type Service struct {
	core    *Core
	builder *builder.Builder
}

// New ...
func New(opt ...Opt) (*Service, error) {
	p := &Service{}
	for _, o := range opt {
		o(p)
	}

	p.builder = builder.New()

	return p, nil
}

// CreatePayment creates a new payment
func (s *Service) CreatePayment(
	ctx context.Context,
	req *rpc.CreatePaymentRequest) (*rpc.PaymentResponse, error) {

	payment, err := s.builder.BuildFromCreatePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create the payment
	payment, err = s.core.CreatePayment(ctx, payment)
	if err != nil {
		return nil, err
	}

	return toPaymentResponse(payment), nil
}

// GetPayment retrieves a payment by ID
func (s *Service) GetPayment(
	ctx context.Context,
	req *rpc.GetPaymentRequest) (*rpc.PaymentResponse, error) {

	// Get the payment
	payment, err := s.core.FindPaymentByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toPaymentResponse(payment), nil
}

// UpdatePayment updates an existing payment
func (s *Service) UpdatePayment(
	ctx context.Context,
	req *rpc.UpdatePaymentRequest) (*rpc.PaymentResponse, error) {

	// Fetch payment from the database
	payment, err := s.core.FindPaymentByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	updateables := []string{model.Description}

	// Update the payment
	updatedPayment, err := s.core.UpdatePayment(ctx, updateables, payment)
	if err != nil {
		return nil, err
	}

	return toPaymentResponse(updatedPayment), nil
}

// DeletePayment deletes a payment by ID
func (s *Service) DeletePayment(
	ctx context.Context,
	req *rpc.DeletePaymentRequest) (*rpc.DeletePaymentResponse, error) {

	// Delete the payment
	if err := s.core.DeletePayment(ctx, req.Id); err != nil {
		return nil, err
	}

	return &rpc.DeletePaymentResponse{Id: req.Id}, nil
}

// ListPayments lists all payments
func (s *Service) ListPayments(
	ctx context.Context,
	req *rpc.ListPaymentsRequest) (*rpc.ListPaymentsResponse, error) {

	qb := query.NewBuilder()
	if req.GetReferenceId() != "" {
		qb = qb.Where(model.ReferenceID, query.Equal, req.GetReferenceId())
	}
	if req.GetStatus() != "" {
		qb = qb.Where(model.Status, query.Equal, req.GetStatus())
	}
	if req.GetCount() != 0 {
		qb = qb.Limit(int(req.GetCount()))
	}

	// List the payments
	payments, err := s.core.FindPayments(ctx, qb.Build())
	if err != nil {
		return nil, err
	}

	return &rpc.ListPaymentsResponse{
		Entity:   "collection",
		Count:    int64(len(payments)),
		Payments: payments,
	}, nil
}

func toPaymentResponse(payment *rpc.Payment) *rpc.PaymentResponse {
	return &rpc.PaymentResponse{
		Id:          payment.Id,
		ReferenceId: payment.ReferenceId,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Description: payment.Description,
		Status:      payment.Status,
		Payer:       payment.Payer,
		Payee:       payment.Payee,
		CreatedAt:   payment.CreatedAt,
		UpdatedAt:   payment.UpdatedAt,
		Error:       payment.Error,
	}
}
