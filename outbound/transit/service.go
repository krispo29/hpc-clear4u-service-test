package transit

import "time"

type OutboundTransitService interface {
}

type service struct {
	selfRepo       OutboundTransitRepository
	contextTimeout time.Duration
}

func NewOutboundTransitService(
	selfRepo OutboundTransitRepository,
	timeout time.Duration,
) OutboundTransitService {
	return &service{
		selfRepo:       selfRepo,
		contextTimeout: timeout,
	}
}
