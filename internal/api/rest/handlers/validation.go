package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	validate       = newValidator()
	errInvalidJSON = errors.New("invalid json")
)

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return false
		}
		_, err := uuid.Parse(value)
		return err == nil
	})
	_ = v.RegisterValidation("rfc3339", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return false
		}
		_, err := time.Parse(time.RFC3339, value)
		return err == nil
	})
	return v
}

func decodeAndValidateJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errInvalidJSON
	}
	if err := validate.Struct(dst); err != nil {
		return formatValidationError(err)
	}
	return nil
}

func formatValidationError(err error) error {
	var verr validator.ValidationErrors
	if !errors.As(err, &verr) || len(verr) == 0 {
		return err
	}
	field := verr[0].Field()
	if field == "" {
		field = "field"
	}
	switch verr[0].Tag() {
	case "required":
		return fmt.Errorf("%s is required", field)
	case "email", "uuid", "rfc3339", "oneof", "min", "max":
		return fmt.Errorf("invalid %s", field)
	default:
		return fmt.Errorf("invalid %s", field)
	}
}
