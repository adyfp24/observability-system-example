package exception

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	BadRequestError     struct{ Message string }
	UnauthorizedError   struct{ Message string }
	ForbiddenError      struct{ Message string }
	NotFoundError       struct{ Message string }
	InternalServerError struct{ Message string }
	NotImplementedError struct{ Message string }
	ValidationError     struct {
		FailedField string
		Tag         string
		Value       any
	}
)

func (e *BadRequestError) Error() string {
	return e.Message
}
func (e *UnauthorizedError) Error() string {
	return e.Message
}
func (e *ForbiddenError) Error() string {
	return e.Message
}
func (e *NotFoundError) Error() string {
	return e.Message
}
func (e *InternalServerError) Error() string {
	return e.Message
}
func (e *NotImplementedError) Error() string {
	return e.Message
}

func FormatValidationErrors(errs []ValidationError) string {
	var messages []string
	for _, e := range errs {
		messages = append(messages, fmt.Sprintf("%s %s", e.FailedField, e.Tag))
	}
	return strings.Join(messages, ", ")
}

func PanicIfError(err error, ctx ...string) {
	if err != nil {
		if debug, _ := strconv.ParseBool(os.Getenv("APP_DEBUG")); debug {
			panic(&InternalServerError{
				Message: fmt.Sprintf("%v %v", ctx, err),
			})
		} else {
			log.Printf("%v %v", ctx, err)
			panic(&InternalServerError{
				Message: "Internal Server Error",
			})
		}
	}
}
