package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Srivastava-samarth/sampay/security"
	"github.com/nats-io/nats.go/jetstream"
)

const BankingResponsesConsumer = "BANKING_RESPONSES"
const BankingEventsStream = "BANKING_EVENTS"

func (c *NatsClient) CreateBankingResponsesConsumer(
	ctx context.Context,
) (jetstream.Consumer, error) {
	consumer, err := c.js.CreateOrUpdateConsumer(
		ctx,
		BankingEventsStream,
		jetstream.ConsumerConfig{
			Durable:       BankingResponsesConsumer,
			FilterSubject: "bank.sampay.>",
			AckPolicy:     jetstream.AckExplicitPolicy,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create banking commands consumer: %w",
			err,
		)
	}

	return consumer, nil
}

func (c *NatsClient) ConsumeBankingCommands(
	ctx context.Context,
	consumer jetstream.Consumer,
) (jetstream.ConsumeContext, error) {
	consumeContext, err := consumer.Consume(
		func(msg jetstream.Msg) {
			var event security.SignedEvent

			if err := json.Unmarshal(msg.Data(), &event); err != nil {
				log.Printf("unmarshal banking event: %v", err)
				if termErr := msg.Term(); termErr != nil {
					log.Printf("terminate malformed event: %v", termErr)
				}
				return
			}

			if err := c.verifier.VerifyEvent(event); err != nil {
				log.Printf(
					"invalid event signature subject=%s: %v",
					msg.Subject(),
					err,
				)
				// Invalid signatures are not transient processing failures.
				if termErr := msg.Term(); termErr != nil {
					log.Printf("terminate invalid event: %v", termErr)
				}
				return
			}

			// Ensure the signed event corresponds to the delivered subject.
			if event.EventType != msg.Subject() {
				log.Printf(
					"event type mismatch: subject=%s event_type=%s",
					msg.Subject(),
					event.EventType,
				)
				if termErr := msg.Term(); termErr != nil {
					log.Printf("terminate mismatched event: %v", termErr)
				}
				return
			}

			if err := c.handleBankingEvents(ctx, &event); err != nil {
				log.Printf("handle banking event: %v", err)
				if nakErr := msg.Nak(); nakErr != nil {
					log.Printf("negative acknowledge banking event: %v", nakErr)
				}
				return
			}

			if err := msg.Ack(); err != nil {
				log.Printf("ack banking event: %v", err)
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("consume banking commands: %w", err)
	}

	return consumeContext, nil
}

func (c *NatsClient) handleBankingEvents(
	ctx context.Context,
	event *security.SignedEvent,
) error {
	switch event.EventType {
	case "bank.sampay.account.created":
		result, err := c.handleBankAccountCreated(ctx, event)
		if err != nil {
			return err
		}

		if result == nil || result.Payload == nil {
			return fmt.Errorf("bank account creation response is empty")
		}
		// Find the original outgoing request.
		originalEvent, err := c.repo.GetEventByEventID(
			ctx,
			result.Payload.RequestedEventID,
		)
		if err != nil {
			return fmt.Errorf("get original bank request: %w", err)
		}

		if originalEvent == nil {
			return fmt.Errorf(
				"original bank request not found: %s",
				result.Payload.RequestedEventID,
			)
		}

		if originalEvent.WorkflowID == "" {
			return fmt.Errorf(
				"workflow ID missing for request event: %s",
				originalEvent.EventID,
			)
		}

		// Signal the workflow waiting for this response.
		err = c.temporalClient.SignalWorkflow(
			ctx,
			originalEvent.WorkflowID,
			"",
			"BankAccountCreationResult",
			result,
		)
		if err != nil {
			return fmt.Errorf(
				"signal merchant onboarding workflow: %w",
				err,
			)
		}

		return nil

	case "bank.sampay.wallet.to.bank.transfer.completed":
		result, err := c.handleBankAccountCredit(ctx, event)
		if err != nil {
			return err
		}

		if result == nil || result.Payload == nil {
			return fmt.Errorf("wallet to bank transfe response is empty")
		}
		// Find the original outgoing request.
		originalEvent, err := c.repo.GetEventByEventID(
			ctx,
			result.Payload.RequestedEventID,
		)
		if err != nil {
			return fmt.Errorf("get original bank request: %w", err)
		}

		if originalEvent == nil {
			return fmt.Errorf(
				"original wallet to bank transfer request not found: %s",
				result.Payload.RequestedEventID,
			)
		}

		if originalEvent.WorkflowID == "" {
			return fmt.Errorf(
				"workflow ID missing for request event: %s",
				originalEvent.EventID,
			)
		}

		// Signal the workflow waiting for this response.
		err = c.temporalClient.SignalWorkflow(
			ctx,
			originalEvent.WorkflowID,
			"",
			"WalletToBankTransferResult",
			result,
		)
		if err != nil {
			return fmt.Errorf(
				"signal merchant onboarding workflow: %w",
				err,
			)
		}

		return nil

	case "bank.sampay.bank.to.wallet.transfer.completed":
		result, err := c.handleBankAccountDebit(ctx, event)
		if result == nil || result.Payload == nil {
			return fmt.Errorf("wallet to bank transfe response is empty")
		}
		// Find the original outgoing request.
		originalEvent, err := c.repo.GetEventByEventID(
			ctx,
			result.Payload.RequestedEventID,
		)
		if err != nil {
			return fmt.Errorf("get original bank request: %w", err)
		}

		if originalEvent == nil {
			return fmt.Errorf(
				"original wallet to bank transfer request not found: %s",
				result.Payload.RequestedEventID,
			)
		}

		if originalEvent.WorkflowID == "" {
			return fmt.Errorf(
				"workflow ID missing for request event: %s",
				originalEvent.EventID,
			)
		}

		// Signal the workflow waiting for this response.
		err = c.temporalClient.SignalWorkflow(
			ctx,
			originalEvent.WorkflowID,
			"",
			"BankToWalletTransferResult",
			result,
		)

		if err != nil {
			return fmt.Errorf(
				"signal merchant onboarding workflow: %w",
				err,
			)
		}

		return nil

	case "bank.sampay.transfer.completed":
		result, err := c.handleBankTransfer(ctx, event)
		if err != nil {
			return err
		}

		if result == nil || result.Payload == nil {
			return fmt.Errorf("wallet to bank transfe response is empty")
		}
		// Find the original outgoing request.
		originalEvent, err := c.repo.GetEventByEventID(
			ctx,
			result.Payload.RequestedEventID,
		)
		if err != nil {
			return fmt.Errorf("get original bank request: %w", err)
		}

		if originalEvent == nil {
			return fmt.Errorf(
				"original wallet to bank transfer request not found: %s",
				result.Payload.RequestedEventID,
			)
		}

		if originalEvent.WorkflowID == "" {
			return fmt.Errorf(
				"workflow ID missing for request event: %s",
				originalEvent.EventID,
			)
		}

		// Signal the workflow waiting for this response.
		err = c.temporalClient.SignalWorkflow(
			ctx,
			originalEvent.WorkflowID,
			"",
			"BankToBankTransferResult",
			result,
		)
		if err != nil {
			return fmt.Errorf(
				"signal merchant onboarding workflow: %w",
				err,
			)
		}

		return nil

    case "bank.sampay.vault.topup.completed":
          result, err := c.handleVaultTopup(ctx, event)
		if err != nil {
			return err
		}

		if result == nil || result.Payload == nil {
			return fmt.Errorf("wallet to bank transfe response is empty")
		}
		// Find the original outgoing request.
		originalEvent, err := c.repo.GetEventByEventID(
			ctx,
			result.Payload.RequestedEventID,
		)
		if err != nil {
			return fmt.Errorf("get original bank request: %w", err)
		}

		if originalEvent == nil {
			return fmt.Errorf(
				"original wallet to bank transfer request not found: %s",
				result.Payload.RequestedEventID,
			)
		}

		if originalEvent.WorkflowID == "" {
			return fmt.Errorf(
				"workflow ID missing for request event: %s",
				originalEvent.EventID,
			)
		}

		// Signal the workflow waiting for this response.
		err = c.temporalClient.SignalWorkflow(
			ctx,
			originalEvent.WorkflowID,
			"",
			"CompanyBankToVaultTransferResult",
			result,
		)
		if err != nil {
			return fmt.Errorf(
				"signal merchant onboarding workflow: %w",
				err,
			)
		}

		return nil  

	default:
		return fmt.Errorf(
			"unsupported banking event type: %s",
			event.EventType,
		)
	}
}
