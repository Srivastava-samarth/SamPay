package dto

import "time"

type ComplianceCheckResponse struct {
	MerchantID       string     `json:"merchant_id"`
	KYC              KYCData    `json:"kyc" gorm:"type:jsonb"`
	KYCDate          time.Time `json:"kyc_date"`
	ComplianceStatus string     `json:"compliance_status" gorm:"default:pending;not null"`
	ComplianceDate   time.Time `json:"compliance_date"`
	ComplianceReason string     `json:"compliance_reason"`
}

type KYCData struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
