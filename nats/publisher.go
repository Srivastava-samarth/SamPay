package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/security"
)

func (c *NatsClient) Publish(
	ctx context.Context,
	subject constants.EventType,
	event security.EventEnvelope,
) error {
	signedEvent, err := c.signer.SignEvent(event)
	if err != nil {
		return fmt.Errorf("sign event: %w", err)
	}

	data, err := json.Marshal(signedEvent)
	if err != nil {
		return fmt.Errorf("marshal signed event: %w", err)
	}

	_, err = c.js.Publish(ctx, string(subject), data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	return nil
}
