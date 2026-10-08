package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

func (c *NatsClient) handleBankAccountCreated(
	ctx context.Context,
	msg jetstream.Msg,
)