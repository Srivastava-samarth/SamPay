package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (a *Registry) GetUnsettledPayment(
	ctx context.Context,
) ([]*models.Payment, error) {
	return a.PaymentService.PaymentRepo.GetPaymentsBySettlementStatus(constants.LedgerSettlementPending)
}

func (a *Registry) ExecuteSettlementByMerchant(
	ctx context.Context,
	merchantID uuid.UUID,
	payments []*models.Payment,
) error {
	return a.DB.Transaction(func(tx *gorm.DB) error {
		walletRepo := a.WalletService.WalletRepo.WithTx(tx)
		paymentRepo := a.PaymentService.PaymentRepo.WithTx(tx)

		settlementAmount := decimal.Zero

		for _, payment := range payments {
			settlementAmount = settlementAmount.Add(payment.Amount)
		}

		wallet, errMW := walletRepo.GetWalletByMerchantId(merchantID)
		if errMW != nil {
			return errMW
		}

		if wallet == nil {
			return errors.New("merchant wallet not found")
		}

		if wallet.ReservedBalance.LessThan(settlementAmount) {
			return errors.New("insufficient reserved balance for settlement")
		}

		_, errUWB := walletRepo.UpdateWalletBalance(
			merchantID,
			&dto.UpdateWallletBalanceRequest{
				AvailableBalance: wallet.AvailableBalance.Add(settlementAmount),
				ReservedBalance:  wallet.ReservedBalance.Sub(settlementAmount),
			},
		)
		if errUWB != nil {
			return fmt.Errorf("failed to update merchant wallet: %w", errUWB)
		}

		for _, payment := range payments {
			if payment == nil {
				continue
			}
			settlementStatus := constants.LedgerSettlementSettled
			_, errUSS := paymentRepo.UpdateSettlementStatusByID(
				payment.ID,
				&settlementStatus,
			)
			if errUSS != nil {
				return fmt.Errorf(
					"failed to update settlement status for payment %s: %w",
					payment.ID,
					errUSS,
				)
			}
		}

		return nil
	})
}
