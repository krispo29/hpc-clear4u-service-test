package auth

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

func (s *loggingService) SignIn(ctx context.Context, username string, password string) (result *SignInResponseModel, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "sign_in",
			"username", username,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.next.SignIn(ctx, username, password)
}
