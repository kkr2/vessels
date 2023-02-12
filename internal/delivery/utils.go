package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kkr2/vessels/internal/errors"
	"github.com/kkr2/vessels/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

// Get request id from echo context
func GetRequestID(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}

// ReqIDCtxKey is a key used for the Request ID in context
type ReqIDCtxKey struct{}

// Get ctx with timeout and request id from echo context
func GetCtxWithReqID(c echo.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request().Context(), time.Second*15)
	ctx = context.WithValue(ctx, ReqIDCtxKey{}, GetRequestID(c))
	return ctx, cancel
}

// Get context  with request id
func GetRequestCtx(c echo.Context) context.Context {
	return context.WithValue(c.Request().Context(), ReqIDCtxKey{}, GetRequestID(c))
}

// Get config path for local or docker
func GetConfigPath(configPath string) string {
	if configPath == "docker" {
		return "./config-docker"
	}
	return "./internal/config/config-local"
}

// Get user ip address
func GetIPAddress(c echo.Context) string {
	return c.Request().RemoteAddr
}

// Error response with logging error for echo context
func ErrResponseWithLog(ctx echo.Context, logger logger.Logger, err error) error {
	logger.Errorf(
		"ErrResponseWithLog, RequestID: %s, IPAddress: %s, Error: %s",
		GetRequestID(ctx),
		GetIPAddress(ctx),
		err,
	)
	return ctx.JSON(ErrorResponse(err))
}

// Error response with logging error for echo context
func LogResponseError(ctx echo.Context, logger logger.Logger, err error) {
	logger.Errorf(
		"ErrResponseWithLog, RequestID: %s, IPAddress: %s, Error: %s",
		GetRequestID(ctx),
		GetIPAddress(ctx),
		err,
	)
}

// Read request body and validate
func ReadRequest(ctx echo.Context, request interface{}) error {
	if err := ctx.Bind(request); err != nil {
		return err
	}
	return validate.StructCtx(ctx.Request().Context(), request)
}

// Read sanitize and validate request
func SanitizeRequest(ctx echo.Context, request interface{}) error {
	operation := errors.Op("utils.SanitizeRequest")
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	defer ctx.Request().Body.Close()

	sanBody, err := SanitizeJSON(body)
	if err != nil {
		return errors.E(operation, errors.KindBadInput, err)
	}

	if err = json.Unmarshal(sanBody, request); err != nil {
		return errors.E(operation, errors.KindBadInput, err)
	}
	if err = validate.StructCtx(ctx.Request().Context(), request); err != nil {
		return errors.E(operation, errors.KindBadInput, err)
	}
	//TODO: bad input
	return nil
}

// VALIDATOR
// Use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate struct fields
func ValidateStruct(ctx context.Context, s interface{}) error {
	return validate.StructCtx(ctx, s)
}

// SANITIZOR
var sanitizer *bluemonday.Policy

func init() {
	sanitizer = bluemonday.UGCPolicy()
}

// Sanitize json
func SanitizeJSON(s []byte) ([]byte, error) {
	d := json.NewDecoder(bytes.NewReader(s))
	d.UseNumber()
	var i interface{}
	err := d.Decode(&i)
	if err != nil {
		return nil, err
	}
	sanitize(i)
	return json.MarshalIndent(i, "", "    ")
}

func sanitize(data interface{}) {
	switch d := data.(type) {
	case map[string]interface{}:
		for k, v := range d {
			switch tv := v.(type) {
			case string:
				d[k] = sanitizer.Sanitize(tv)
			case map[string]interface{}:
				sanitize(tv)
			case []interface{}:
				sanitize(tv)
			case nil:
				delete(d, k)
			}
		}
	case []interface{}:
		if len(d) > 0 {
			switch d[0].(type) {
			case string:
				for i, s := range d {
					d[i] = sanitizer.Sanitize(s.(string))
				}
			case map[string]interface{}:
				for _, t := range d {
					sanitize(t)
				}
			case []interface{}:
				for _, t := range d {
					sanitize(t)
				}
			}
		}
	}
}
