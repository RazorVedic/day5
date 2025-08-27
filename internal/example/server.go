package example

import (
	"context"

	"github.com/razorpay/go-foundation-v2/internal/example/validator"
	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// Service defines the interface for payment service
type Service interface {
	CreatePayment(
		ctx context.Context,
		req *rpc.CreatePaymentRequest) (*rpc.PaymentResponse, error)
	GetPayment(
		ctx context.Context,
		req *rpc.GetPaymentRequest) (*rpc.PaymentResponse, error)
	UpdatePayment(
		ctx context.Context,
		req *rpc.UpdatePaymentRequest) (*rpc.PaymentResponse, error)
	DeletePayment(
		ctx context.Context,
		req *rpc.DeletePaymentRequest) (*rpc.DeletePaymentResponse, error)
	ListPayments(
		ctx context.Context,
		req *rpc.ListPaymentsRequest) (*rpc.ListPaymentsResponse, error)
}

// Server is the implementation of the payment server
type Server struct {
	rpc.UnimplementedPaymentServiceServer
	service Service
}

// NewServer returns a new payment server
func NewServer(srv Service) *Server {
	return &Server{
		service: srv,
	}
}

// CreatePayment handles the creation of a new payment
func (s *Server) CreatePayment(
	ctx context.Context,
	req *rpc.CreatePaymentRequest,
) (*rpc.PaymentResponse, error) {
	err := validator.ValidateCreatePaymentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.service.CreatePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetPayment handles retrieving a payment by ID
func (s *Server) GetPayment(
	ctx context.Context,
	req *rpc.GetPaymentRequest,
) (*rpc.PaymentResponse, error) {
	err := validator.ValidateGetPaymentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.service.GetPayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdatePayment handles updating an existing payment
func (s *Server) UpdatePayment(
	ctx context.Context,
	req *rpc.UpdatePaymentRequest,
) (*rpc.PaymentResponse, error) {
	err := validator.ValidateUpdatePaymentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.service.UpdatePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeletePayment handles deleting a payment by ID
func (s *Server) DeletePayment(
	ctx context.Context,
	req *rpc.DeletePaymentRequest,
) (*rpc.DeletePaymentResponse, error) {
	err := validator.ValidateDeletePaymentRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.service.DeletePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ListPayments handles listing all payments
func (s *Server) ListPayments(
	ctx context.Context,
	req *rpc.ListPaymentsRequest,
) (*rpc.ListPaymentsResponse, error) {
	err := validator.ValidateListPaymentsRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.service.ListPayments(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
