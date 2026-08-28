package web

import (
	"strings"
	"testing"

	"github.com/roledio/roled/internal/models"
)

func TestBuildRedirectURL_PreservesSignupQueryParam(t *testing.T) {
	h := &handler{}
	req := &models.RenderAuthorizeRequest{
		ClientID:            "client-id",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		State:               "state",
		IsSignup:            true,
	}

	redirectURL := h.buildRedirectURL(req, "")

	if !strings.Contains(redirectURL, "signup=true") {
		t.Fatalf("expected signup query param to be preserved, got %q", redirectURL)
	}
}

func TestBuildRedirectURL_OmitsSignupQueryParamWhenDisabled(t *testing.T) {
	h := &handler{}
	req := &models.RenderAuthorizeRequest{
		ClientID:            "client-id",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		State:               "state",
	}

	redirectURL := h.buildRedirectURL(req, "")

	if strings.Contains(redirectURL, "signup=true") {
		t.Fatalf("did not expect signup query param when signup disabled, got %q", redirectURL)
	}
}
