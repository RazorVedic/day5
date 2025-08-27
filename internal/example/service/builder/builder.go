package builder

import (
	"context"

	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// opt is used to build the model request for service
type opt func(context.Context, *rpc.Payment) error

// Builder to build payment domain model from request
type Builder struct{}

// New returns the payment builder with required grpc clients
func New() *Builder {
	return &Builder{}
}

// BuildFromCreatePayment builds the payment from create payment request
func (b *Builder) BuildFromCreatePayment(
	ctx context.Context,
	req *rpc.CreatePaymentRequest,
) (*rpc.Payment, error) {
	var opts []opt

	opts = append(opts,
		b.id(),
		b.amount(req.Amount),
		b.currency(req.Currency),
		b.referenceID(req.ReferenceId),
		b.status("created"),
		b.description(req.Description),
		b.payer(req.Payer),
		b.payee(req.Payee),
	)

	payment := &rpc.Payment{
		Payer: &rpc.Payer{},
		Payee: &rpc.Payee{},
	}

	for _, o := range opts {
		err := o(ctx, payment)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}
