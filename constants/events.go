package constants

type EventType string

const (
	EventTypeSampayToBankAccountCreateRequested EventType = "sampay.bank.account.create.requested"
	EventTypeBankToSampayBankAccountCreated     EventType = "bank.sampay.account.created"
	EventTypeSampayToBankCreditRequested        EventType = "sampay.bank.wallet.bank.credit.requested"
	EventTypeWalletToBankTransferCompleted      EventType = "bank.sampay.wallet.to.bank.transfer.completed"
	EventTypeBankToBankTrasferRequested         EventType = "sampay.bank.transfer.requested"
	EventTypeBankTransferCompleted              EventType = "bank.sampay.transfer.completed"
)

type EventDirection string

const (
	EventDirectionInbound  EventDirection = "inbound"
	EventDirectionOutbound EventDirection = "outbound"
)

type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusProcessed  EventStatus = "processed"
	EventStatusFailed     EventStatus = "failed"
)
