package flashutil

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

const (
	cookiePrefix = "flash_"
)

type FlashData struct {
	ErrorMessage string
	ErrorDebug   string
}

func (f *FlashData) WithError(err error) *FlashData {
	if err == nil {
		return f
	}
	var ce pkgerrors.CustomError
	if !errors.As(err, &ce) {
		// This error is not of type pkgerrors.CustomError, it is an unexpected error
		ce = pkgerrors.ErrSystemError.WithError(err)
	}
	f.ErrorMessage = ce.Msg
	if ce.Err != nil {
		f.ErrorDebug = ce.Err.Error()
	} else if ce.DebugMessage != "" {
		f.ErrorDebug = ce.DebugMessage
	}
	return f
}

func SetData(c fiber.Ctx, data any) {
	if data == nil {
		log.WithContext(c.Context()).Debug("Empty flash data, skip writing flash data")
		return
	}
	m, err := json.Marshal(data)
	if err != nil {
		log.WithContext(c.Context()).Errorw("Failed to marshal flash data to json", "err", err)
		return
	}
	// Store JSON as URL-escaped string to ensure cookie value is valid.
	escaped := url.QueryEscape(string(m))
	Write(c, "data", escaped)
}

func ReadData(c fiber.Ctx, data any) error {
	val := Read(c, "data")
	if val == "" {
		return nil
	}
	// Unescape stored value before unmarshalling JSON
	unesc, err := url.QueryUnescape(val)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(unesc), data)
}

func Write(c fiber.Ctx, key, value string) {
	c.Cookie(&fiber.Cookie{
		Name:    sanitizeKey(key),
		Value:   value,
		Expires: time.Now().Add(10 * time.Second), // Expires in ten seconds
	})
}

func Read(c fiber.Ctx, key string) string {
	cookieName := sanitizeKey(key)
	val := c.Cookies(cookieName)
	if val != "" {
		c.ClearCookie(cookieName)
	}
	return val
}

func sanitizeKey(key string) string {
	// Pattern to match one or more non-word characters (\W)
	// \W matches any character that is not a letter, number, or underscore.
	// \W is a shorthand for [^a-zA-Z0-9_]
	// + is a quantifier that matches one or more occurrences of the preceding element.
	// This ensures that consecutive special characters are replaced by only a single underscore
	// e.g., "user@name!!" becomes "user_name_" instead of "user_name__"
	pattern := regexp.MustCompile(`[\W]+`)
	return cookiePrefix + pattern.ReplaceAllString(key, "_")
}
