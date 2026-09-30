package pkceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyCodeChallenge(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	for _, tc := range []struct {
		name, verifier, challenge, method string
		want                              bool
	}{
		{"valid S256", verifier, challenge, "S256", true},
		{"lowercase method rejected", verifier, challenge, "s256", false},
		{"plain method rejected", verifier, verifier, "plain", false},
		{"missing method rejected", verifier, challenge, "", false},
		{"wrong challenge", verifier, challenge + "x", "S256", false},
		{"wrong verifier", "different", challenge, "S256", false},
		{"missing verifier", "", challenge, "S256", false},
		{"missing challenge", verifier, "", "S256", false},
		{"padded challenge rejected", verifier, challenge + "=", "S256", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, VerifyCodeChallenge(tc.verifier, tc.challenge, tc.method))
		})
	}
}
