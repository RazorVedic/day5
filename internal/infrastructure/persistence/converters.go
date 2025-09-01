package persistence

import (
	"day5/internal/domain/entities"
)

// Product conversions

// ProductToModel converts domain entity to persistence model
func ProductToModel(entity *entities.Product) *Product {
	if entity == nil {
		return nil
	}

	return &Product{
		ID:          entity.ID,
		ProductName: entity.ProductName,
		Price:       entity.Price,
		Quantity:    entity.Quantity,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// ModelToProduct converts persistence model to domain entity
func ModelToProduct(model *Product, entity *entities.Product) {
	if model == nil || entity == nil {
		return
	}

	entity.ID = model.ID
	entity.ProductName = model.ProductName
	entity.Price = model.Price
	entity.Quantity = model.Quantity
	entity.CreatedAt = model.CreatedAt
	entity.UpdatedAt = model.UpdatedAt
}

// Customer conversions

// CustomerToModel converts domain entity to persistence model
func CustomerToModel(entity *entities.Customer) *Customer {
	if entity == nil {
		return nil
	}

	return &Customer{
		ID:        entity.ID,
		Name:      entity.Name,
		Email:     entity.Email,
		Phone:     entity.Phone,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

// ModelToCustomer converts persistence model to domain entity
func ModelToCustomer(model *Customer, entity *entities.Customer) {
	if model == nil || entity == nil {
		return
	}

	entity.ID = model.ID
	entity.Name = model.Name
	entity.Email = model.Email
	entity.Phone = model.Phone
	entity.CreatedAt = model.CreatedAt
	entity.UpdatedAt = model.UpdatedAt
}

// Order conversions

// OrderToModel converts domain entity to persistence model
func OrderToModel(entity *entities.Order) *Order {
	if entity == nil {
		return nil
	}

	return &Order{
		ID:          entity.ID,
		CustomerID:  entity.CustomerID,
		ProductID:   entity.ProductID,
		Quantity:    entity.Quantity,
		UnitPrice:   entity.UnitPrice,
		TotalAmount: entity.TotalAmount,
		OrderDate:   entity.OrderDate,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// ModelToOrder converts persistence model to domain entity
func ModelToOrder(model *Order, entity *entities.Order) {
	if model == nil || entity == nil {
		return
	}

	entity.ID = model.ID
	entity.CustomerID = model.CustomerID
	entity.ProductID = model.ProductID
	entity.Quantity = model.Quantity
	entity.UnitPrice = model.UnitPrice
	entity.TotalAmount = model.TotalAmount
	entity.OrderDate = model.OrderDate
	entity.CreatedAt = model.CreatedAt
	entity.UpdatedAt = model.UpdatedAt

	// Related entities - load if present
	if model.Customer.ID != "" {
		entity.Customer = &entities.Customer{}
		ModelToCustomer(&model.Customer, entity.Customer)
	}

	if model.Product.ID != "" {
		entity.Product = &entities.Product{}
		ModelToProduct(&model.Product, entity.Product)
	}
}

// Transaction conversions

// TransactionToModel converts domain entity to persistence model
func TransactionToModel(entity *entities.Transaction) *Transaction {
	if entity == nil {
		return nil
	}

	return &Transaction{
		ID:            entity.ID,
		OrderID:       entity.OrderID,
		CustomerID:    entity.CustomerID,
		ProductID:     entity.ProductID,
		Type:          string(entity.Type),
		Amount:        entity.Amount,
		Quantity:      entity.Quantity,
		UnitPrice:     entity.UnitPrice,
		Description:   entity.Description,
		TransactionAt: entity.TransactionAt,
		CreatedAt:     entity.CreatedAt,
	}
}

// ModelToTransaction converts persistence model to domain entity
func ModelToTransaction(model *Transaction, entity *entities.Transaction) {
	if model == nil || entity == nil {
		return
	}

	entity.ID = model.ID
	entity.OrderID = model.OrderID
	entity.CustomerID = model.CustomerID
	entity.ProductID = model.ProductID
	entity.Type = entities.TransactionType(model.Type)
	entity.Amount = model.Amount
	entity.Quantity = model.Quantity
	entity.UnitPrice = model.UnitPrice
	entity.Description = model.Description
	entity.TransactionAt = model.TransactionAt
	entity.CreatedAt = model.CreatedAt

	// Related entities - load if present
	if model.Order.ID != "" {
		entity.Order = &entities.Order{}
		ModelToOrder(&model.Order, entity.Order)
	}

	if model.Customer.ID != "" {
		entity.Customer = &entities.Customer{}
		ModelToCustomer(&model.Customer, entity.Customer)
	}

	if model.Product.ID != "" {
		entity.Product = &entities.Product{}
		ModelToProduct(&model.Product, entity.Product)
	}
}

// CustomerCooldown conversions

// CooldownToModel converts domain entity to persistence model
func CooldownToModel(entity *entities.CustomerCooldown) *CustomerCooldown {
	if entity == nil {
		return nil
	}

	return &CustomerCooldown{
		CustomerID:    entity.CustomerID,
		LastOrderTime: entity.LastOrderTime,
		UpdatedAt:     entity.UpdatedAt,
	}
}

// ModelToCooldown converts persistence model to domain entity
func ModelToCooldown(model *CustomerCooldown, entity *entities.CustomerCooldown) {
	if model == nil || entity == nil {
		return
	}

	entity.CustomerID = model.CustomerID
	entity.LastOrderTime = model.LastOrderTime
	entity.UpdatedAt = model.UpdatedAt
}

// Batch conversion helpers

// ModelsToProducts converts slice of models to slice of entities
func ModelsToProducts(models []Product) []*entities.Product {
	products := make([]*entities.Product, len(models))
	for i, model := range models {
		products[i] = &entities.Product{}
		ModelToProduct(&model, products[i])
	}
	return products
}

// ModelsToCustomers converts slice of models to slice of entities
func ModelsToCustomers(models []Customer) []*entities.Customer {
	customers := make([]*entities.Customer, len(models))
	for i, model := range models {
		customers[i] = &entities.Customer{}
		ModelToCustomer(&model, customers[i])
	}
	return customers
}

// ModelsToOrders converts slice of models to slice of entities
func ModelsToOrders(models []Order) []*entities.Order {
	orders := make([]*entities.Order, len(models))
	for i, model := range models {
		orders[i] = &entities.Order{}
		ModelToOrder(&model, orders[i])
	}
	return orders
}

// ModelsToTransactions converts slice of models to slice of entities
func ModelsToTransactions(models []Transaction) []*entities.Transaction {
	transactions := make([]*entities.Transaction, len(models))
	for i, model := range models {
		transactions[i] = &entities.Transaction{}
		ModelToTransaction(&model, transactions[i])
	}
	return transactions
}
