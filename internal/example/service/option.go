package service

import (
	storage "github.com/razorpay/goutils/sqlstorage"

	"github.com/razorpay/go-foundation-v2/internal/example/repo"
)

// Opt is used to set the dependencies and configurations for the service
type Opt func(*Service)

// WithStorage ...
func WithStorage(s storage.Store) Opt {
	return func(p *Service) {
		p.core = NewCore(
			repo.NewPayment(s),
			repo.NewEvent(s),
		)
	}
}
