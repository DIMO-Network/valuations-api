package helpers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger *zerolog.Logger, isProduction bool) error {
	logger = GetLogger(c, logger)

	code := fiber.StatusInternalServerError // Default 500 statuscode

	var e *fiber.Error
	fiberTypeErr := errors.As(err, &e)
	if fiberTypeErr {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	codeStr := strconv.Itoa(code)

	logger.Err(err).Str("httpStatusCode", codeStr).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Msg("caught an error from http request")
	// return an opaque error if we're in a higher level environment and we haven't specified an fiber type err.
	if !fiberTypeErr && isProduction {
		err = fiber.NewError(fiber.StatusInternalServerError, "Internal error")
	}

	return c.Status(code).JSON(ErrorRes{
		Code:    code,
		Message: err.Error(),
	})
}

type ErrorRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
