package middlewares

import (
	"fmt"
	"net/url"
	"strings"

	fibernewrelic "github.com/gofiber/contrib/v3/newrelic"
	"github.com/gofiber/fiber/v3"
)

// NewRelicRequestAttributes updates the New Relic transaction name to use parameterized
// route paths (e.g., /api/v1/projects/:project_id) instead of actual IDs.
// It also adds path parameters and query parameters as custom attributes.
// This must run as a route-level middleware (not app.Use) so that Fiber's router has
// already resolved the route and populated c.Route().Params and c.Params() correctly.
func NewRelicRequestAttributes(c fiber.Ctx) error {
	// Snapshot the raw query string now — fasthttp recycles the request buffer
	// after the handler returns, so we must copy before calling c.Next().
	rawQuery := string(c.Request().URI().QueryString())

	// Snapshot path params now — they are valid on the matched route at this point.
	paramKeys := c.Route().Params
	params := make(map[string]string, len(paramKeys))
	for _, key := range paramKeys {
		if v := c.Params(key); v != "" {
			params[key] = v
		}
	}

	err := c.Next()

	txn := fibernewrelic.FromContext(c)
	if txn == nil {
		return err
	}

	// Set the transaction name using the parameterized route pattern.
	if fullPath := c.FullPath(); fullPath != "" {
		txn.SetName(fmt.Sprintf("%s %s", c.Method(), fullPath))
	}

	// Add path params captured before Next().
	for key, value := range params {
		txn.AddAttribute(fmt.Sprintf("path_param.%s", key), value)
	}

	// Parse and add query params from the snapshotted raw query string.
	if rawQuery != "" {
		for pair := range strings.SplitSeq(rawQuery, "&") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
				key, _ := url.QueryUnescape(kv[0])
				value, _ := url.QueryUnescape(kv[1])
				if key != "" && value != "" {
					txn.AddAttribute(fmt.Sprintf("query_param.%s", key), value)
				}
			}
		}
	}

	return err
}
