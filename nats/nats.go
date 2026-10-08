package nats

import (
	"fmt"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/security"
	"github.com/Srivastava-samarth/sampay/services"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.temporal.io/sdk/client"
)

type NatsClient struct {
	conn           *nats.Conn
	js             jetstream.JetStream
	service        *services.Services
	temporalClient client.Client

	signer   *security.Signer
	verifier *security.Verifier
}

func NewClient(
	config *config.NatsConfig,
	service *services.Services,
	temporalClient client.Client,
	privateKey []byte,
	bankPublicKey []byte,
) (*NatsClient, error) {
	conn, err := nats.Connect(config.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create jetstream client: %w", err)
	}

	signer, err := security.NewSigner(privateKey)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create event signer: %w", err)
	}

	verifier, err := security.NewVerifier(bankPublicKey)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create event verifier: %w", err)
	}

	return &NatsClient{
		conn:           conn,
		js:             js,
		service:        service,
		temporalClient: temporalClient,
		signer:         signer,
		verifier:       verifier,
	}, nil
}

func (c *NatsClient) Close() {
	c.conn.Close()
}