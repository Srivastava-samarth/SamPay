package controllers

import (
    "github.com/Srivastava-samarth/sampay/services"
    "go.temporal.io/sdk/client"
    "gorm.io/gorm"
)

type Controller struct {
    DB             *gorm.DB
    TemporalClient client.Client

    AuthService        *services.AuthService
    BankService        *services.BankService
    MerchantService    *services.MerchantService
    PaymentService     *services.PaymentService
    IdempotencyService *services.IdempotencyService
    PayoutService      *services.PayoutService
    RefundService      *services.RefundService
    UserService        *services.UserService
    VaultService       *services.VaultService
    WalletService      *services.WalletService
    LedgerService      *services.LedgerService
}

func NewController(
    db *gorm.DB,
    temporalClient client.Client,
    authService *services.AuthService,
    bankService *services.BankService,
    merchantService *services.MerchantService,
    paymentService *services.PaymentService,
    idempotencyService *services.IdempotencyService,
    payoutService *services.PayoutService,
    refundService *services.RefundService,
    userService *services.UserService,
    vaultService *services.VaultService,
    walletService *services.WalletService,
    ledgerService *services.LedgerService,
) *Controller {
    return &Controller{
        DB:                 db,
        TemporalClient:     temporalClient,
        AuthService:        authService,
        BankService:        bankService,
        MerchantService:    merchantService,
        PaymentService:     paymentService,
        IdempotencyService: idempotencyService,
        PayoutService:      payoutService,
        RefundService:      refundService,
        UserService:        userService,
        VaultService:       vaultService,
        WalletService:      walletService,
        LedgerService:      ledgerService,
    }
}