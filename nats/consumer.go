package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Srivastava-samarth/sampay/security"
	"github.com/nats-io/nats.go/jetstream"
)

const BankingCommandsConsumer = "BANKING_COMMANDS"
const BankingEventsStream = "BANKING_EVENTS"

func (c *NatsClient) CreateBankingCommandsConsumer(
	ctx context.Context,
) (jetstream.Consumer, error) {
	consumer, err := c.js.CreateOrUpdateConsumer(
		ctx,
		BankingEventsStream,
		jetstream.ConsumerConfig{
			Durable:       BankingCommandsConsumer,
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
				return
			}

			if err := c.verifier.VerifyEvent(event); err != nil {
				log.Printf(
					"invalid event signature subject=%s: %v",
					msg.Subject(),
					err,
				)
				return
			}
			if err := c.handleBankingEvents(ctx, msg); err != nil {
				log.Printf("handle banking event: %v", err)
				return
			}

			if err := msg.Ack(); err != nil {
				log.Printf("ack message: %v", err)
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"consume banking commands: %w",
			err,
		)
	}

	return consumeContext, nil
}

func (c *NatsClient) handleBankingEvents(
	ctx context.Context,
	msg jetstream.Msg,
) error {
	log.Printf(
		"received banking event subject=%s data=%s",
		msg.Subject(),
		string(msg.Data()),
	)

	switch msg.Subject() {
	case "bank.sampay.account.created":
		return nil
	case "bank.sampay.account.create.failed":
		return nil
	default:
		return fmt.Errorf(
			"unsupported banking event subject: %s",
			msg.Subject(),
		)
	}
}
