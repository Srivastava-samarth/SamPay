package nats

import (
	"fmt"

	"github.com/Srivastava-samarth/sampay/config"
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
}

func NewClient(
	config *config.NatsConfig, 
	service *services.Services,
	temporalClient client.Client,
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

	// signer, err := security.NewSigner(privateKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("create event signer: %w", err)
	// }

	// verifier, err := security.NewVerifier(sampayPublicKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("create event verifier: %w", err)
	// }

	return &NatsClient{
		conn: conn,
		js:   js,
		service: service,
		temporalClient: temporalClient,
	}, nil
}

func (c *NatsClient) Close() {
	c.conn.Close()
}
		