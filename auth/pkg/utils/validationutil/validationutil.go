package validationutil

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gofiber/fiber/v3/log"
	"github.com/govalues/decimal"
)

const (
	ValidatorTagNotBlank               = "notblank"
	ValidatorTagBase64RawURL           = "base64rawurl"
	ValidatorTagAlphanumUnderscore     = "alphanum_underscore"
	ValidatorTagAlphanumDash           = "alphanum_dash"
	ValidatorTagAlphanumDashUnderscore = "alphanum_dash_underscore"
)

var validate, translator = setupValidator()

func GetValidator() *validator.Validate {
	return validate
}

func ValidateStruct(ctx context.Context, s any) error {
	err := validate.StructCtx(ctx, s)
	if err != nil {
		trans := err.(validator.ValidationErrors).Translate(translator)
		messages := []string{}
		for _, msg := range trans {
			messages = append(messages, msg)
		}
		return errors.New(strings.Join(messages, ", "))
	}
	return err
}

// IsValidEmail checks if the provided email has a valid format.
func IsValidEmail(email string) bool {
	err := validate.Var(email, "required,email")
	return err == nil
}

func setupValidator() (*validator.Validate, ut.Translator) {
	v := validator.New()
	v.RegisterCustomTypeFunc(decimalValidator, decimal.Decimal{})
	translator := registerTranslation(v)
	registerCustomValidationAndTranslation(v, translator)
	return v, translator
}

func decimalValidator(field reflect.Value) any {
	if dec, ok := field.Interface().(decimal.Decimal); ok {
		f, _ := dec.Float64()
		return f
	}
	return nil
}

func registerTranslation(val *validator.Validate) ut.Translator {
	english := en.New()
	uni := ut.New(english, english)
	t, found := uni.GetTranslator("en")
	if !found {
		log.Warn("Validation translation not found, will use default message format")
	}
	err := en_translations.RegisterDefaultTranslations(val, t)
	if err != nil {
		log.Warn("Register default translation error:", err)
	}
	return t
}

func registerCustomValidationAndTranslation(val *validator.Validate, translator ut.Translator) {
	validations := []struct {
		tag          string
		validatorFn  validator.Func
		errorMessage string
	}{
		{
			tag:          ValidatorTagNotBlank,
			validatorFn:  notBlankValidator,
			errorMessage: "{0} must not be empty or contains only whitespace characters",
		},
		{
			tag:          ValidatorTagBase64RawURL,
			validatorFn:  nil, // Already exists, just adding the translation
			errorMessage: "{0} must be a valid Base64 raw URL encoded string",
		},
		{
			tag: ValidatorTagAlphanumDash,
			validatorFn: func(fl validator.FieldLevel) bool {
				return IsAlphanumDash(fl.Field().String())
			},
			errorMessage: "{0} must contain only alphanumeric and dash characters",
		},
		{
			tag: ValidatorTagAlphanumUnderscore,
			validatorFn: func(fl validator.FieldLevel) bool {
				return IsAlphanumUnderscore(fl.Field().String())
			},
			errorMessage: "{0} must contain only alphanumeric and underscore characters",
		},
		{
			tag: ValidatorTagAlphanumDashUnderscore,
			validatorFn: func(fl validator.FieldLevel) bool {
				return IsAlphanumDashUnderscore(fl.Field().String())
			},
			errorMessage: "{0} must contain only alphanumeric, underscore, and dash characters",
		},
	}
	for _, v := range validations {
		if v.validatorFn != nil {
			err := val.RegisterValidation(v.tag, v.validatorFn)
			if err != nil {
				log.Warnf("Register validation '%s' error: %s", v.tag, err)
				return
			}
		}
		err := val.RegisterTranslation(v.tag, translator,
			func(ut ut.Translator) error {
				return ut.Add(v.tag, v.errorMessage, true)
			},
			func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T(v.tag, fe.Field())
				return t
			})
		if err != nil {
			log.Warnf("Register '%s' translation error: %s", v.tag, err)
		}
	}
}

func IsAlphanumDashUnderscore(val string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(val)
}

func IsAlphanumDash(val string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(val)
}

func IsAlphanumUnderscore(val string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(val)
}

// notBlankValidator validates that a string field is not blank (not empty/whitespace only).
// Unlike validators.NotBlank, this passes on empty strings - letting 'required' handle emptiness.
// It only validates when there's actual content, similar to how 'min' and 'max' work.
func notBlankValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		s := field.String()
		// Pass on empty string - let 'required' handle it
		if s == "" {
			return true
		}
		// Check if string contains only whitespace
		for _, r := range s {
			if !unicode.IsSpace(r) {
				return true
			}
		}
		return false
	default:
		return true
	}
}
