package constants

type EventType string

const (
	EventTypeSampayToBankAccountCreateRequested EventType = "sampay.bank.account.create.requested"
	EventTypeBankToSampayBankAccountCreated EventType = "bank.sampay.account.created"
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