package http

import (
	stdErrs "errors"
	"regexp"

	"consul-journey/internal"
	"consul-journey/internal/errors"
	"consul-journey/internal/server/http/handlers"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

type Server struct {
	logger    *zap.Logger
	cfg       *Config
	app       *fiber.App
	startedCh chan struct{}
}

func New(logger *zap.Logger, cfg *Config) (*Server, error) {
	s := &Server{
		logger:    logger.Named("server.http"),
		cfg:       cfg,
		startedCh: make(chan struct{}),
	}
	s.createFiberApp()
	return s, nil
}

func (s *Server) createFiberApp() {
	app := fiber.New(s.fiberConfig())
	app.
		Hooks().
		OnPostStartupMessage(
			func(*fiber.PostStartupMessageData) error {
				close(s.startedCh)
				return nil
			},
		)
	app.Use(recover.New())
	s.app = app
}

func (s *Server) RegisterHandler(path string, handler handlers.Handler) {
	handler.Register(s.app.Group(path).(*fiber.Group))
}

// It panics if the server fails to listen and serve
func (s *Server) Start() {
	s.logger.Info("starting ...")
	go func() {
		if err := s.app.Listen(s.cfg.srvAddr(), s.fiberListenConfig()); err != nil {
			s.logger.Panic("failed to listen and serve", zap.Error(err))
		}
	}()
	<-s.startedCh
	s.logger.Info("started", zap.String("address", s.cfg.srvAddr()))
}

func (s *Server) Stop() {
	s.logger.Info("shutting down ...")
	if err := s.app.ShutdownWithTimeout(s.cfg.ShutdownTimeout); err != nil {
		s.logger.Error("failed to shutdown", zap.Error(err))
		return
	}
	s.logger.Info("shutdown successful")
}

func serverHeader() string {
	return internal.AppName() + " Server " + internal.Version() + " (" + internal.Revision() + ")"
}

func (s *Server) fiberConfig() fiber.Config {
	return fiber.Config{
		ServerHeader:                    serverHeader(),
		StrictRouting:                   false,
		CaseSensitive:                   false,
		DisableHeadAutoRegister:         false,
		Immutable:                       false,
		UnescapePath:                    false,
		BodyLimit:                       s.cfg.BodyLimit,
		MaxRanges:                       1,
		Concurrency:                     s.cfg.Concurrency,
		Views:                           nil,
		ViewsLayout:                     "",
		PassLocalsToViews:               false,
		PassLocalsToContext:             true,
		ReadTimeout:                     s.cfg.ReadTimeout,
		WriteTimeout:                    s.cfg.WriteTimeout,
		IdleTimeout:                     s.cfg.IdleTimeout,
		ReadBufferSize:                  s.cfg.ReadBufferSize,
		WriteBufferSize:                 s.cfg.WriteBufferSize,
		CompressedFileSuffixes:          map[string]string{"gzip": ".gz", "br": ".br", "zstd": ".zst"},
		ProxyHeader:                     s.cfg.ProxyHeader,
		GETOnly:                         false,
		ErrorHandler:                    s.errorHandler,
		DisableKeepalive:                !s.cfg.Keepalive,
		DisableDefaultDate:              true,
		DisableDefaultContentType:       false,
		DisableHeaderNormalizing:        true,
		AppName:                         internal.AppName(),
		SharedStorage:                   nil,
		SharedStatePrefix:               "",
		StreamRequestBody:               s.cfg.StreamRequestBody,
		DisablePreParseMultipartForm:    true,
		ReduceMemoryUsage:               false,
		JSONEncoder:                     sonic.Marshal,
		JSONDecoder:                     sonic.Unmarshal,
		MsgPackEncoder:                  nil,
		MsgPackDecoder:                  nil,
		CBOREncoder:                     nil,
		CBORDecoder:                     nil,
		XMLEncoder:                      nil,
		XMLDecoder:                      nil,
		TrustProxy:                      false,
		TrustProxyConfig:                fiber.DefaultTrustProxyConfig,
		EnableIPValidation:              false,
		ColorScheme:                     fiber.DefaultColors,
		StructValidator:                 nil,
		RequestMethods:                  fiber.DefaultMethods,
		EnableSplittingOnParsers:        false,
		Services:                        nil,
		ServicesStartupContextProvider:  nil,
		ServicesShutdownContextProvider: nil,
		RegexHandler:                    regexp.MustCompile,
	}
}

func (s *Server) fiberListenConfig() fiber.ListenConfig {
	return fiber.ListenConfig{
		ListenerNetwork:       fiber.NetworkTCP,
		ShutdownTimeout:       s.cfg.ShutdownTimeout,
		DisableStartupMessage: true,
		EnablePrintRoutes:     s.cfg.PrintRoutes,
	}
}

func (s *Server) errorHandler(c fiber.Ctx, err error) error {
	var convErr *errors.Error
	if fErr, ok := stdErrs.AsType[*fiber.Error](err); ok {
		convErr = convertFiberError(fErr).Caller(getCaller(c))
		s.logger.Warn(convErr.Error(), convErr.Fields()...)
	} else if aErr, ok := stdErrs.AsType[*errors.Error](err); ok {
		convErr = aErr.Caller(getCaller(c))
		s.logger.Warn(convErr.Error(), convErr.Fields()...)
	} else {
		convErr = errors.NewInternal(getCaller(c), err)
		s.logger.Error(convErr.Error(), convErr.Fields()...)
	}
	return c.Status(convErr.Status()).JSON(convErr)
}

func getCaller(c fiber.Ctx) string {
	return c.Method() + " " + c.OriginalURL()
}

func convertFiberError(fErr *fiber.Error) *errors.Error {
	switch fErr.Code {
	case fiber.StatusBadRequest:
		return ErrBadRequest
	case fiber.StatusUnauthorized:
		return ErrUnauthorized
	case fiber.StatusPaymentRequired:
		return ErrPaymentRequired
	case fiber.StatusForbidden:
		return ErrForbidden
	case fiber.StatusNotFound:
		return ErrPathNotFound
	case fiber.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case fiber.StatusNotAcceptable:
		return ErrNotAcceptable
	case fiber.StatusProxyAuthRequired:
		return ErrProxyAuthRequired
	case fiber.StatusRequestTimeout:
		return ErrRequestTimeout
	case fiber.StatusConflict:
		return ErrConflict
	case fiber.StatusGone:
		return ErrGone
	case fiber.StatusLengthRequired:
		return ErrLengthRequired
	case fiber.StatusPreconditionFailed:
		return ErrPreconditionFailed
	case fiber.StatusRequestEntityTooLarge:
		return ErrRequestEntityTooLarge
	case fiber.StatusRequestURITooLong:
		return ErrRequestURITooLong
	case fiber.StatusUnsupportedMediaType:
		return ErrUnsupportedMediaType
	case fiber.StatusRequestedRangeNotSatisfiable:
		return ErrRequestedRangeNotSatisfiable
	case fiber.StatusExpectationFailed:
		return ErrExpectationFailed
	case fiber.StatusTeapot:
		return ErrTeapot
	case fiber.StatusMisdirectedRequest:
		return ErrMisdirectedRequest
	case fiber.StatusUnprocessableEntity:
		return ErrUnprocessableEntity
	case fiber.StatusLocked:
		return ErrLocked
	case fiber.StatusFailedDependency:
		return ErrFailedDependency
	case fiber.StatusTooEarly:
		return ErrTooEarly
	case fiber.StatusUpgradeRequired:
		return ErrUpgradeRequired
	case fiber.StatusPreconditionRequired:
		return ErrPreconditionRequired
	case fiber.StatusTooManyRequests:
		return ErrTooManyRequests
	case fiber.StatusRequestHeaderFieldsTooLarge:
		return ErrRequestHeaderFieldsTooLarge
	case fiber.StatusUnavailableForLegalReasons:
		return ErrUnavailableForLegalReasons
	case fiber.StatusNotImplemented:
		return ErrNotImplemented
	case fiber.StatusBadGateway:
		return ErrBadGateway
	case fiber.StatusServiceUnavailable:
		return ErrServiceUnavailable
	case fiber.StatusGatewayTimeout:
		return ErrGatewayTimeout
	case fiber.StatusHTTPVersionNotSupported:
		return ErrHTTPVersionNotSupported
	case fiber.StatusVariantAlsoNegotiates:
		return ErrVariantAlsoNegotiates
	case fiber.StatusInsufficientStorage:
		return ErrInsufficientStorage
	case fiber.StatusLoopDetected:
		return ErrLoopDetected
	case fiber.StatusNotExtended:
		return ErrNotExtended
	case fiber.StatusNetworkAuthenticationRequired:
		return ErrNetworkAuthenticationRequired
	default:
		return ErrInternalServerError
	}
}
