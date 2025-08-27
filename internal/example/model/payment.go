package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	storage "github.com/razorpay/goutils/sqlstorage"
)

const (
	// ID ...
	ID = "id"
	// Description ...
	Description = "description"
	// ReferenceID ...
	ReferenceID = "reference_id"
	// Status ...
	Status = "status"
)

// Payment represents the details of a Payment.
type Payment struct {
	storage.ObjectMeta
	// Amount ....
	Amount int64 `json:"amount" gorm:"not null,column:amount"`
	// Currency ...
	Currency string `json:"currency" gorm:"not null,column:currency"`
	// ReferenceID ...
	ReferenceID string `json:"reference_id" gorm:"not null,column:reference_id"`
	// Status ...
	Status string `json:"status" gorm:"not null,column:status"`
	// Description ...
	Description string `json:"description" gorm:"not null,column:description"`
	// ErrorCode ...
	ErrorCode string `json:"error_code" gorm:"column:error_code"`
	// ErrorMessage ...
	ErrorMessage string `json:"error_message" gorm:"column:error_message"`
	// Payer ...
	Payer *PayerDetails `json:"payer" gorm:"not null,column:payer,type:json"`
	// Payee ...
	Payee *PayeeDetails `json:"payee" gorm:"not null,column:payee,type:json"`
}

// PayerDetails represents the details of the payer.
type PayerDetails struct {
	// ID ...
	ID string `json:"id"`
	// Name ...
	Name string `json:"name"`
	// VPA ...
	VPA string `json:"vpa"`
	// Fundsource ...
	Fundsource Fundsource `json:"fundsource"`
}

// PayeeDetails represents the details of the payee.
type PayeeDetails struct {
	// ID ...
	ID string `json:"id"`
	// Name ...
	Name string `json:"name"`
	// VPA ...
	VPA string `json:"vpa"`
	// Fundsource ...
	Fundsource Fundsource `json:"fundsource"`
}

// Fundsource represents the fund source details.
type Fundsource struct {
	// ID ...
	ID string `json:"id"`
	// AccountNumber ...
	AccountNumber string `json:"account_number"`
	// IFSC ...
	IFSC string `json:"ifsc"`
}

// Value returns the JSON value, implementing the driver.Valuer interface.
func (p PayerDetails) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan scans value into PayerDetails, implementing the sql.Scanner interface.
func (p *PayerDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// Value returns the JSON value, implementing the driver.Valuer interface.
func (p PayeeDetails) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan scans value into PayeeDetails, implementing the sql.Scanner interface.
func (p *PayeeDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}
