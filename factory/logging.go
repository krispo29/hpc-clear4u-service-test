package factory

import (
	"github.com/go-kit/log"

	"hpc-express-service/auth"
	"hpc-express-service/common"
	inbound "hpc-express-service/inbound/express"
)

type LoggingFactory struct {
}

func InitialLoggingFactory(logger log.Logger, svc *ServiceFactory) {

	svc.AuthSvc = auth.NewLoggingService(log.With(logger, "module", "auth"), svc.AuthSvc)
	svc.CommonSvc = common.NewLoggingService(log.With(logger, "module", "common"), svc.CommonSvc)
	svc.InboundExpressServiceSvc = inbound.NewLoggingService(log.With(logger, "module", "inbound_express"), svc.InboundExpressServiceSvc)

}
