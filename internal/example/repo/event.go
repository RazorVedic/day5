package repo

import (
	"context"

	storage "github.com/razorpay/goutils/sqlstorage"

	"github.com/razorpay/go-foundation-v2/internal/example/model"
)

// Event implements the Event Repo
type Event struct {
	store storage.Store
}

// NewEvent creates a new Repo for Event
func NewEvent(store storage.Store) *Event {
	return &Event{
		store: store,
	}
}

// Create creates a new record in the event database.
func (r *Event) Create(
	ctx context.Context,
	event *model.Event,
) (*model.Event, error) {
	_, err := r.store.Create(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}
