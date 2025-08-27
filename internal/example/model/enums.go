package model

// PaymentStatus represents the possible statuses for a Payment.
type PaymentStatus string

// Represents a payment lifecycle.
const (
	// Created ...
	Created PaymentStatus = "created"
	// Failed ...
	Failed PaymentStatus = "failed"
	// Success ...
	Success PaymentStatus = "success"
)
