package errors

import (
	"errors"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
)

const (
	ErrInternalCode = "INTERNAL_SERVER"
)

type Error struct {
	status  int
	code    string
	message string
	details map[string]any
	callers []string
	err     error
}

func NewInternal(caller string, err error) *Error {
	res := New(500, ErrInternalCode, "Internal server error occurred").Caller(caller)
	res.err = err
	return res
}

func New(status int, code, message string) *Error {
	return &Error{
		status:  status,
		code:    code,
		message: message,
		details: make(map[string]any),
		callers: make([]string, 0),
		err:     nil,
	}
}

func (e *Error) keepSafeNewInstance() *Error {
	if len(e.callers) != 0 || len(e.details) != 0 {
		return e
	}
	return &Error{
		status:  e.status,
		code:    e.code,
		message: e.message,
		details: make(map[string]any),
		callers: make([]string, 0),
		err:     e.err,
	}
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Fields() []zap.Field {
	fields := make([]zap.Field, 0, len(e.details)+4)
	fields = append(fields, zap.Error(e.err))
	fields = append(fields, zap.Int("status", e.status))
	fields = append(fields, zap.String("code", e.code))
	for k, v := range e.details {
		fields = append(fields, zap.Any(k, v))
	}
	fields = append(fields, zap.Strings("callers", e.callers))
	return fields
}

func (e *Error) Caller(caller string) *Error {
	if caller == "" {
		return e
	}
	e = e.keepSafeNewInstance()
	e.callers = append(e.callers, caller)
	return e
}

func (e *Error) Detail(key string, value any) *Error {
	if key == "" {
		return e
	}
	e = e.keepSafeNewInstance()
	e.details[key] = value
	return e
}

func (e *Error) Code() string {
	return e.code
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(
		presentableError{
			Code:    e.code,
			Message: e.message,
			Details: e.details,
		},
	)
}

func Pipe(caller string, err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Caller(caller)
	}
	return NewInternal(caller, err)
}

type presentableError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}
