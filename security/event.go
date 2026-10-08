package security

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	EventID   uuid.UUID       `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type SignedEvent struct {
	EventID   uuid.UUID       `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	Signature []byte          `json:"signature"`
}

func (s *Signer) SignEvent(event EventEnvelope) (*SignedEvent, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal event for signing: %w",
			err,
		)
	}

	signature, err := s.Sign(data)
	if err != nil {
		return nil, fmt.Errorf(
			"sign event: %w",
			err,
		)
	}

	return &SignedEvent{
		EventID:   event.EventID,
		EventType: event.EventType,
		Payload:   event.Payload,
		Signature: signature,
	}, nil
}

func (v *Verifier) VerifyEvent(event SignedEvent) error {
	envelope := EventEnvelope{
		EventID:   event.EventID,
		EventType: event.EventType,
		Payload:   event.Payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf(
			"marshal event for verification: %w",
			err,
		)
	}

	if err := v.Verify(data, event.Signature); err != nil {
		return fmt.Errorf(
			"verify event: %w",
			err,
		)
	}

	return nil
}