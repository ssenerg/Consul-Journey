package http

import (
	"consul-journey/internal/errors"

	"github.com/gofiber/fiber/v3"
)

const (
	ErrBadRequestCode                    = "BAD_REQUEST"
	ErrUnauthorizedCode                  = "UNAUTHORIZED"
	ErrPaymentRequiredCode               = "PAYMENT_REQUIRED"
	ErrForbiddenCode                     = "FORBIDDEN"
	ErrPathNotFoundCode                  = "PATH_NOT_FOUND"
	ErrMethodNotAllowedCode              = "METHOD_NOT_ALLOWED"
	ErrNotAcceptableCode                 = "NOT_ACCEPTABLE"
	ErrProxyAuthRequiredCode             = "PROXY_AUTH_REQUIRED"
	ErrRequestTimeoutCode                = "REQUEST_TIMEOUT"
	ErrConflictCode                      = "CONFLICT"
	ErrGoneCode                          = "GONE"
	ErrLengthRequiredCode                = "LENGTH_REQUIRED"
	ErrPreconditionFailedCode            = "PRECONDITION_FAILED"
	ErrRequestEntityTooLargeCode         = "REQUEST_ENTITY_TOO_LARGE"
	ErrRequestURITooLongCode             = "REQUEST_URI_TOO_LONG"
	ErrUnsupportedMediaTypeCode          = "UNSUPPORTED_MEDIA_TYPE"
	ErrRequestedRangeNotSatisfiableCode  = "REQUESTED_RANGE_NOT_SATISFIABLE"
	ErrExpectationFailedCode             = "EXPECTATION_FAILED"
	ErrTeapotCode                        = "TEAPOT"
	ErrMisdirectedRequestCode            = "MISDIRECTED_REQUEST"
	ErrUnprocessableEntityCode           = "UNPROCESSABLE_ENTITY"
	ErrLockedCode                        = "LOCKED"
	ErrFailedDependencyCode              = "FAILED_DEPENDENCY"
	ErrTooEarlyCode                      = "TOO_EARLY"
	ErrUpgradeRequiredCode               = "UPGRADE_REQUIRED"
	ErrPreconditionRequiredCode          = "PRECONDITION_REQUIRED"
	ErrTooManyRequestsCode               = "TOO_MANY_REQUESTS"
	ErrRequestHeaderFieldsTooLargeCode   = "REQUEST_HEADER_FIELDS_TOO_LARGE"
	ErrUnavailableForLegalReasonsCode    = "UNAVAILABLE_FOR_LEGAL_REASONS"
	ErrNotImplementedCode                = "NOT_IMPLEMENTED"
	ErrBadGatewayCode                    = "BAD_GATEWAY"
	ErrServiceUnavailableCode            = "SERVICE_UNAVAILABLE"
	ErrGatewayTimeoutCode                = "GATEWAY_TIMEOUT"
	ErrHTTPVersionNotSupportedCode       = "HTTP_VERSION_NOT_SUPPORTED"
	ErrVariantAlsoNegotiatesCode         = "VARIANT_ALSO_NEGOTIATES"
	ErrInsufficientStorageCode           = "INSUFFICIENT_STORAGE"
	ErrLoopDetectedCode                  = "LOOP_DETECTED"
	ErrNotExtendedCode                   = "NOT_EXTENDED"
	ErrNetworkAuthenticationRequiredCode = "NETWORK_AUTHENTICATION_REQUIRED"
)

var (
	ErrBadRequest                    = errors.New(fiber.StatusBadRequest, ErrBadRequestCode, "Bad request received")
	ErrUnauthorized                  = errors.New(fiber.StatusUnauthorized, ErrUnauthorizedCode, "Unauthorized access")
	ErrPaymentRequired               = errors.New(fiber.StatusPaymentRequired, ErrPaymentRequiredCode, "Payment is required")
	ErrForbidden                     = errors.New(fiber.StatusForbidden, ErrForbiddenCode, "Access is forbidden")
	ErrPathNotFound                  = errors.New(fiber.StatusNotFound, ErrPathNotFoundCode, "Path not found")
	ErrMethodNotAllowed              = errors.New(fiber.StatusMethodNotAllowed, ErrMethodNotAllowedCode, "Method not allowed")
	ErrNotAcceptable                 = errors.New(fiber.StatusNotAcceptable, ErrNotAcceptableCode, "Request not acceptable")
	ErrProxyAuthRequired             = errors.New(fiber.StatusProxyAuthRequired, ErrProxyAuthRequiredCode, "Proxy authentication is required")
	ErrRequestTimeout                = errors.New(fiber.StatusRequestTimeout, ErrRequestTimeoutCode, "Request timeout")
	ErrConflict                      = errors.New(fiber.StatusConflict, ErrConflictCode, "Resource conflict")
	ErrGone                          = errors.New(fiber.StatusGone, ErrGoneCode, "Resource is gone")
	ErrLengthRequired                = errors.New(fiber.StatusLengthRequired, ErrLengthRequiredCode, "Length is required")
	ErrPreconditionFailed            = errors.New(fiber.StatusPreconditionFailed, ErrPreconditionFailedCode, "Precondition failed")
	ErrRequestEntityTooLarge         = errors.New(fiber.StatusRequestEntityTooLarge, ErrRequestEntityTooLargeCode, "Request entity too large")
	ErrRequestURITooLong             = errors.New(fiber.StatusRequestURITooLong, ErrRequestURITooLongCode, "Request URI too long")
	ErrUnsupportedMediaType          = errors.New(fiber.StatusUnsupportedMediaType, ErrUnsupportedMediaTypeCode, "Unsupported media type")
	ErrRequestedRangeNotSatisfiable  = errors.New(fiber.StatusRequestedRangeNotSatisfiable, ErrRequestedRangeNotSatisfiableCode, "Requested range not satisfiable")
	ErrExpectationFailed             = errors.New(fiber.StatusExpectationFailed, ErrExpectationFailedCode, "Expectation failed")
	ErrTeapot                        = errors.New(fiber.StatusTeapot, ErrTeapotCode, "I'm a teapot")
	ErrMisdirectedRequest            = errors.New(fiber.StatusMisdirectedRequest, ErrMisdirectedRequestCode, "Misdirected request")
	ErrUnprocessableEntity           = errors.New(fiber.StatusUnprocessableEntity, ErrUnprocessableEntityCode, "Unprocessable entity")
	ErrLocked                        = errors.New(fiber.StatusLocked, ErrLockedCode, "Resource is locked")
	ErrFailedDependency              = errors.New(fiber.StatusFailedDependency, ErrFailedDependencyCode, "Failed dependency")
	ErrTooEarly                      = errors.New(fiber.StatusTooEarly, ErrTooEarlyCode, "Too early")
	ErrUpgradeRequired               = errors.New(fiber.StatusUpgradeRequired, ErrUpgradeRequiredCode, "Upgrade required")
	ErrPreconditionRequired          = errors.New(fiber.StatusPreconditionRequired, ErrPreconditionRequiredCode, "Precondition required")
	ErrTooManyRequests               = errors.New(fiber.StatusTooManyRequests, ErrTooManyRequestsCode, "Too many requests")
	ErrRequestHeaderFieldsTooLarge   = errors.New(fiber.StatusRequestHeaderFieldsTooLarge, ErrRequestHeaderFieldsTooLargeCode, "Request header fields too large")
	ErrUnavailableForLegalReasons    = errors.New(fiber.StatusUnavailableForLegalReasons, ErrUnavailableForLegalReasonsCode, "Unavailable for legal reasons")
	ErrInternalServerError           = errors.New(fiber.StatusInternalServerError, errors.ErrInternalCode, "Internal server error occurred")
	ErrNotImplemented                = errors.New(fiber.StatusNotImplemented, ErrNotImplementedCode, "Not implemented")
	ErrBadGateway                    = errors.New(fiber.StatusBadGateway, ErrBadGatewayCode, "Bad gateway")
	ErrServiceUnavailable            = errors.New(fiber.StatusServiceUnavailable, ErrServiceUnavailableCode, "Service unavailable")
	ErrGatewayTimeout                = errors.New(fiber.StatusGatewayTimeout, ErrGatewayTimeoutCode, "Gateway timeout")
	ErrHTTPVersionNotSupported       = errors.New(fiber.StatusHTTPVersionNotSupported, ErrHTTPVersionNotSupportedCode, "HTTP version not supported")
	ErrVariantAlsoNegotiates         = errors.New(fiber.StatusVariantAlsoNegotiates, ErrVariantAlsoNegotiatesCode, "Variant also negotiates")
	ErrInsufficientStorage           = errors.New(fiber.StatusInsufficientStorage, ErrInsufficientStorageCode, "Insufficient storage")
	ErrLoopDetected                  = errors.New(fiber.StatusLoopDetected, ErrLoopDetectedCode, "Loop detected")
	ErrNotExtended                   = errors.New(fiber.StatusNotExtended, ErrNotExtendedCode, "Not extended")
	ErrNetworkAuthenticationRequired = errors.New(fiber.StatusNetworkAuthenticationRequired, ErrNetworkAuthenticationRequiredCode, "Network authentication required")
)
