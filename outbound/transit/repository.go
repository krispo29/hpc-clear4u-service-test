package transit

import "time"

type OutboundTransitRepository interface {
}

type repository struct {
	contextTimeout time.Duration
}

func NewRepository(
	timeout time.Duration,
) OutboundTransitRepository {
	return &repository{
		contextTimeout: timeout,
	}
}
