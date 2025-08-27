package repo

import (
	"context"
	"errors"
	"fmt"

	storage "github.com/razorpay/goutils/sqlstorage"
	"github.com/razorpay/goutils/sqlstorage/query"
	"github.com/razorpay/goutils/sqlstorage/sql"

	"github.com/razorpay/go-foundation-v2/internal/example/model"
)

// Payment implements the Payment Repo
type Payment struct {
	store storage.Store
}

// NewPayment creates a new instance of paymento
func NewPayment(
	store storage.Store,
) *Payment {

	return &Payment{
		store: store,
	}
}

// Create creates a new payment record in the database.
func (r *Payment) Create(
	ctx context.Context,
	payment *model.Payment) (*model.Payment, error) {
	_, err := r.store.Create(ctx, payment)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// Get retrieves a payment record by ID from the database.
func (r *Payment) Get(ctx context.Context, id string) (*model.Payment, error) {
	payment := &model.Payment{}
	_, err := r.store.FindByID(ctx, payment, id)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// Update updates a payment record in the database
// with the given updateables(columns to update)
func (r *Payment) Update(ctx context.Context,
	updateables []string, payment *model.Payment) (int64, error) {
	var rowsAffected int64
	var ok bool

	query := fmt.Sprintf("%s = ?", "id")

	val, err := r.store.UpdateWithUpdateables(ctx,
		updateables, payment, payment, query, payment.ID)
	if err != nil {
		return 0, err
	}

	if rowsAffected, ok = val.(int64); !ok {
		return 0, errors.New("error in type assertion of rowsAffected to int64")
	}

	return rowsAffected, err
}

// Delete deletes a payment record by ID from the database.
func (r *Payment) Delete(ctx context.Context, id string) error {
	payment := &model.Payment{ObjectMeta: storage.ObjectMeta{ID: id}}
	_, err := r.store.Delete(ctx, payment)
	return err
}

// FindByID fetches a payment using id
func (r *Payment) FindByID(
	ctx context.Context,
	id string) (*model.Payment, error) {

	paymentModel := &model.Payment{}
	_, err := r.store.FindByID(ctx, paymentModel, id)
	if err != nil {
		return nil, err
	}
	return paymentModel, nil
}

// FindAllWhere ...
func (r *Payment) FindAllWhere(ctx context.Context,
	query *query.Query) ([]*model.Payment, error) {

	var models []*model.Payment
	_, err := r.store.FindWithQuery(ctx, &models, query)
	if err != nil {
		if errors.Is(err, sql.ErrBadRequestRecordNotFound) {
			return []*model.Payment{}, nil
		}
		return nil, err
	}
	return models, nil
}

// ExecuteTxn runs multiple commands within a transaction, committing
// changes upon success and rolling them back if an error occurs.
func (r *Payment) ExecuteTxn(
	ctx context.Context,
	fn func(ctx context.Context) error) error {

	return r.store.ExecuteTxn(ctx, fn)
}
