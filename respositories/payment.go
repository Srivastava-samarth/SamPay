package repositories

import "gorm.io/gorm"

type PaymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(
	db *gorm.DB,
) *PaymentRepository{
	return &PaymentRepository{
		DB:db,
	}
}

