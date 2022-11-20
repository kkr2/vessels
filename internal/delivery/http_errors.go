package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	custtomerrors "github.com/kkr2/vessels/internal/errors"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrServiceUnavailable  = errors.New("external systems not available")
	ErrNotFound            = errors.New("not Found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrBadQueryParams      = errors.New("invalid query params")
	ErrInternalServerError = errors.New("internal Server Error")
)

// Rest error interface
type RestErr interface {
	Status() int
	Error() string
	Causes() interface{}
}

// Rest error struct
type RestError struct {
	ErrStatus int         `json:"status,omitempty"`
	ErrError  string      `json:"error,omitempty"`
	ErrCauses interface{} `json:"cause,omitempty"`
}

// Error  Error() interface method
func (e RestError) Error() string {
	return fmt.Sprintf("status: %d - errors: %s - causes: %v", e.ErrStatus, e.ErrError, e.ErrCauses)
}

// Error status
func (e RestError) Status() int {
	return e.ErrStatus
}

// RestError Causes
func (e RestError) Causes() interface{} {
	return e.ErrCauses
}

// New Rest Error
func NewRestError(status int, err string, causes interface{}) RestErr {
	return RestError{
		ErrStatus: status,
		ErrError:  err,
		ErrCauses: causes,
	}
}

// New Rest Error With Message
func NewRestErrorWithMessage(status int, err string, causes interface{}) RestErr {
	return RestError{
		ErrStatus: status,
		ErrError:  err,
		ErrCauses: causes,
	}
}

// New Rest Error From Bytes
func NewRestErrorFromBytes(bytes []byte) (RestErr, error) {
	var apiErr RestError
	if err := json.Unmarshal(bytes, &apiErr); err != nil {
		return nil, errors.New("invalid json")
	}
	return apiErr, nil
}

// New Bad Request Error
func NewBadRequestError(causes interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusBadRequest,
		ErrError:  ErrBadRequest.Error(),
		ErrCauses: causes,
	}
}

// New Not Found Error
func NewNotFoundError(causes interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusNotFound,
		ErrError:  ErrNotFound.Error(),
		ErrCauses: causes,
	}
}

// New Unauthorized Error
func NewUnauthorizedError(causes interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusUnauthorized,
		ErrError:  ErrUnauthorized.Error(),
		ErrCauses: causes,
	}
}

// New Forbidden Error
func NewForbiddenError(causes interface{}) RestErr {
	return RestError{
		ErrStatus: http.StatusForbidden,
		ErrError:  ErrForbidden.Error(),
		ErrCauses: causes,
	}
}

// New Internal Server Error
func NewInternalServerError(causes interface{}) RestErr {
	result := RestError{
		ErrStatus: http.StatusInternalServerError,
		ErrError:  ErrInternalServerError.Error(),
		ErrCauses: causes,
	}
	return result
}

// Parser of error string messages returns RestError
func ParseErrors(err error) RestErr {
	switch {
	case custtomerrors.IsKind(custtomerrors.KindInternal, err):
		return NewRestError(http.StatusInternalServerError, ErrInternalServerError.Error(), errors.Unwrap(err).Error())
	case custtomerrors.IsKind(custtomerrors.KindBadInput, err):
		return NewRestError(http.StatusBadRequest, ErrBadRequest.Error(), errors.Unwrap(err).Error())
	case custtomerrors.IsKind(custtomerrors.KindExternalRPC, err):
		return NewRestError(http.StatusServiceUnavailable, ErrServiceUnavailable.Error(), errors.Unwrap(err).Error())
	case custtomerrors.IsKind(custtomerrors.KindNotAuthorized, err):
		return NewRestError(http.StatusUnauthorized, ErrUnauthorized.Error(), errors.Unwrap(err).Error())
	case custtomerrors.IsKind(custtomerrors.KindNotAllowed, err):
		return NewRestError(http.StatusForbidden, ErrForbidden.Error(), errors.Unwrap(err).Error())
	default:
		if restErr, ok := err.(RestErr); ok {
			return restErr
		}
		return NewRestError(http.StatusInternalServerError, ErrInternalServerError.Error(), err)
	}
}

// Error response
func ErrorResponse(err error) (int, interface{}) {
	return ParseErrors(err).Status(), ParseErrors(err)
}
