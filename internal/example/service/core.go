package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	logger "github.com/razorpay/goutils/logger/v3"
	storage "github.com/razorpay/goutils/sqlstorage"
	"github.com/razorpay/goutils/sqlstorage/query"

	"github.com/razorpay/go-foundation-v2/internal/example/model"
	"github.com/razorpay/go-foundation-v2/internal/example/repo"
	topic "github.com/razorpay/go-foundation-v2/internal/topic/example/payment"
	gofoundationv2rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1"
	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// Core ...
type Core struct {
	paymentRepo *repo.Payment
	eventRepo   *repo.Event
	logger      *slog.Logger
}

// NewCore creates a new core
func NewCore(paymentRepo *repo.Payment, eventRepo *repo.Event) *Core {
	slogger := slog.New(logger.NewHandler(nil)).
		With(slog.String("service", "example-core"))

	return &Core{
		paymentRepo: paymentRepo,
		eventRepo:   eventRepo,
		logger:      slogger,
	}
}

// CreatePayment creates a new payment record and in the database
func (c *Core) CreatePayment(
	ctx context.Context,
	payment *rpc.Payment,
) (*rpc.Payment, error) {
	paymentModel, err := toDBModel(payment)
	if err != nil {
		return nil, err
	}

	event, err := toEvent(ctx, payment)
	if err != nil {
		return nil, err
	}
	c.logger.InfoContext(ctx, "Event data", "data", string(event.Data))

	err = c.paymentRepo.ExecuteTxn(ctx, func(ctx context.Context) error {
		paymentModel, err = c.paymentRepo.Create(ctx, paymentModel)
		if err != nil {
			return err
		}

		event, err = c.eventRepo.Create(ctx, event)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toDomainModel(paymentModel), nil
}

// FindPaymentByID finds a single payment record in the database by ID.
func (c *Core) FindPaymentByID(
	ctx context.Context,
	id string,
) (*rpc.Payment, error) {
	paymentModel, err := c.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainModel(paymentModel), nil
}

// UpdatePayment updates a payment record in the database.
func (c *Core) UpdatePayment(
	ctx context.Context,
	updatables []string,
	payment *rpc.Payment,
) (*rpc.Payment, error) {
	paymentModel, err := toDBModel(payment)
	if err != nil {
		return nil, err
	}

	event, err := toEvent(ctx, payment)
	if err != nil {
		return nil, err
	}

	err = c.paymentRepo.ExecuteTxn(ctx, func(ctx context.Context) error {
		rowsAffected, err := c.paymentRepo.Update(ctx, updatables, paymentModel)
		if err != nil {
			return err
		} else if rowsAffected == 0 {
			return fmt.Errorf("no payment record found to be updated")
		}

		event, err = c.eventRepo.Create(ctx, event)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toDomainModel(paymentModel), nil
}

// DeletePayment deletes a payment record by ID from the database.
func (c *Core) DeletePayment(ctx context.Context, id string) error {
	// Delete payment from the repository
	err := c.paymentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// FindPayments finds all payment records in the database by the given condition
func (c *Core) FindPayments(
	ctx context.Context, query *query.Query) ([]*rpc.Payment, error) {

	payments, err := c.paymentRepo.FindAllWhere(ctx, query)
	if err != nil {
		return nil, err
	}
	var rpcPayments []*rpc.Payment
	for _, paymentModel := range payments {
		payment := toDomainModel(paymentModel)
		rpcPayments = append(rpcPayments, payment)
	}
	return rpcPayments, nil
}

// toDBModel converts payment domain model to a payment db model.
func toDBModel(payment *rpc.Payment) (*model.Payment, error) {
	objectMeta, err := storage.NewObjectMeta()
	if err != nil {
		return nil, err
	}
	objectMeta.ID = payment.Id

	return &model.Payment{
		ObjectMeta:  *objectMeta,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		ReferenceID: payment.ReferenceId,
		Status:      payment.Status,
		Description: payment.Description,
		Payer: &model.PayerDetails{
			ID:   payment.Payer.Id,
			Name: payment.Payer.Name,
			VPA:  payment.Payer.Vpa,
			Fundsource: model.Fundsource{
				ID:            payment.Payer.Fundsource.Id,
				AccountNumber: payment.Payer.Fundsource.AccountNumber,
				IFSC:          payment.Payer.Fundsource.Ifsc,
			}},
		Payee: &model.PayeeDetails{
			ID:   payment.Payee.Id,
			Name: payment.Payee.Name,
			VPA:  payment.Payee.Vpa,
			Fundsource: model.Fundsource{
				ID:            payment.Payee.Fundsource.Id,
				AccountNumber: payment.Payee.Fundsource.AccountNumber,
				IFSC:          payment.Payee.Fundsource.Ifsc,
			}},
	}, nil
}

// toDomainModel converts a payment db model to a payment domain model.
func toDomainModel(paymentModel *model.Payment) *rpc.Payment {
	return &rpc.Payment{
		Id:          paymentModel.ID,
		ReferenceId: paymentModel.ReferenceID,
		Amount:      paymentModel.Amount,
		Currency:    paymentModel.Currency,
		Description: paymentModel.Description,
		Status:      paymentModel.Status,
		Payer: &rpc.Payer{
			Id:   paymentModel.Payer.ID,
			Name: paymentModel.Payer.Name,
			Vpa:  paymentModel.Payer.VPA,
			Fundsource: &rpc.Fundsource{
				Id:            paymentModel.Payer.Fundsource.ID,
				AccountNumber: paymentModel.Payer.Fundsource.AccountNumber,
				Ifsc:          paymentModel.Payer.Fundsource.IFSC,
			}},
		Payee: &rpc.Payee{
			Id:   paymentModel.Payee.ID,
			Name: paymentModel.Payee.Name,
			Vpa:  paymentModel.Payee.VPA,
			Fundsource: &rpc.Fundsource{
				Id:            paymentModel.Payee.Fundsource.ID,
				AccountNumber: paymentModel.Payee.Fundsource.AccountNumber,
				Ifsc:          paymentModel.Payee.Fundsource.IFSC,
			}},
		CreatedAt: paymentModel.CreatedAt,
		UpdatedAt: paymentModel.UpdatedAt,
		Error: &gofoundationv2rpc.Error{
			Code:    paymentModel.ErrorCode,
			Message: paymentModel.ErrorMessage,
		},
	}
}

// EventMetadata represents metadata for events
//
// TODO: move to proto
type EventMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

// PaymentEvent represents a generic payment event
//
// TODO: move to proto
type PaymentEvent struct {
	Metadata *EventMetadata `json:"metadata"`
	Payment  *rpc.Payment   `json:"payment"`
	Type     string         `json:"type"`
}

func toEvent(_ context.Context, payment *rpc.Payment) (*model.Event, error) {
	var eventModel model.Event
	objectMeta, err := storage.NewObjectMeta()
	if err != nil {
		return nil, err
	}
	eventModel.ObjectMeta = *objectMeta

	var eventType string
	var topicName string

	switch payment.Status {
	case string(model.Created):
		eventType = "payment_created"
		topicName = topic.Created.Name
	case string(model.Success):
		eventType = "payment_success"
		topicName = topic.Success.Name
	case string(model.Failed):
		eventType = "payment_failed"
		topicName = topic.Failed.Name
	default:
		return nil, fmt.Errorf(
			"payment status is not supported for get event: %v",
			payment.Status,
		)
	}

	event := &PaymentEvent{
		Metadata: &EventMetadata{
			Timestamp: time.Now().Unix(),
		},
		Payment: payment,
		Type:    eventType,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	eventModel.Data = eventBytes
	eventModel.Topic = topicName
	return &eventModel, nil
}
