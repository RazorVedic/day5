package builder

import (
	"context"

	"github.com/razorpay/goutils/uniqueid"

	rpc "github.com/razorpay/go-foundation-v2/rpc/gofoundationv2/v1/example/payment"
)

// id sets the id for Payment.
func (b *Builder) id() opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		newID, err := uniqueid.New()
		if err != nil {
			return err
		}
		payment.Id = newID
		return nil
	}
}

// amount sets the amount for Payment.
func (b *Builder) amount(amount int64) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Amount = amount
		return nil
	}
}

// currency sets the currency for Payment.
func (b *Builder) currency(currency string) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Currency = currency
		return nil
	}
}

// referenceID sets the reference ID for Payment.
func (b *Builder) referenceID(referenceID string) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.ReferenceId = referenceID
		return nil
	}
}

// status sets the status for Payment.
func (b *Builder) status(status string) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Status = status
		return nil
	}
}

// description sets the description for Payment.
func (b *Builder) description(description string) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Description = description
		return nil
	}
}

// payer sets the payer details for Payment.
func (b *Builder) payer(payer *rpc.Payer) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Payer = payer
		return nil
	}
}

// payee sets the payee details for Payment.
func (b *Builder) payee(payee *rpc.Payee) opt {
	return func(ctx context.Context, payment *rpc.Payment) error {
		payment.Payee = payee
		return nil
	}
}
