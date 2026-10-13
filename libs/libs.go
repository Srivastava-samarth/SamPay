package libs

import "github.com/Srivastava-samarth/sampay/config"

type Libs struct {
	Config *config.InternalConfig
}

func NewLibs(
	cfg *config.InternalConfig,
) *Libs{
	return &Libs{
		Config: cfg,
	}
}