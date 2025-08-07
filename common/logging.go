package common

import (
	"context"
	"time"

	"github.com/go-kit/log"
)

type loggingService struct {
	logger log.Logger
	next   Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) GetAllExchangeRates(ctx context.Context) (result []*GetExchangeRateModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_all_exchange_rates",
			"total", len(result),
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetAllExchangeRates(ctx)
}

func (s *loggingService) GetAllConvertTemplates(ctx context.Context, param string) (result []*GetAllConvertTemplateModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "get_all_convert_templates",
			"total", len(result),
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.GetAllConvertTemplates(ctx, param)
}
