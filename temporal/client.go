package temporal

import (
	"go.temporal.io/sdk/client"

	"github.com/Srivastava-samarth/sampay/config"
)

func NewClient(cfg *config.TemporalConfig) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort: cfg.Host,
	})
}
