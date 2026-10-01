package mail

import (
	"errors"
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailTemplatesEscapeUserContent(t *testing.T) {
	for _, name := range []string{"reset-password", "activate-member", "verify-email", "invite-user"} {
		t.Run(name, func(t *testing.T) {
			data := map[string]any{"DisplayName": `<img src=x onerror=alert(1)>`, "ProjectName": `<script>project</script>`, "AccountName": `<b>account</b>`, "ProjectLogoURL": `javascript:alert(1)`, "ResetPasswordURL": "https://auth.example/reset?one=1&two=2", "ActivateMemberURL": "https://auth.example/member", "ActivateProjectUserURL": "https://auth.example/user", "VerifyURL": "https://auth.example/verify", "LoginURL": "https://app.example/login"}
			body, err := LoadAndParseTemplate("templates/html/"+name+".html", data)
			require.NoError(t, err)
			require.Contains(t, body, "&lt;script&gt;project&lt;/script&gt;")
			require.NotContains(t, body, "<script>")
			require.NotContains(t, body, "javascript:")
			require.Contains(t, body, "#ZgotmplZ")
			require.Contains(t, body, "https://auth.example/")
		})
	}
}

func TestEmailTemplateFailures(t *testing.T) {
	_, err := LoadTemplate("missing.html")
	require.Error(t, err)
	body, err := LoadAndParseTemplate("missing.html", nil)
	require.Error(t, err)
	require.Empty(t, body)
	failure := errors.New("render failed")
	tpl, err := template.New("failure").Funcs(template.FuncMap{"fail": func() (string, error) { return "", failure }}).Parse("{{fail}}")
	require.NoError(t, err)
	body, err = ParseTemplate(tpl, nil)
	require.ErrorIs(t, err, failure)
	require.Empty(t, body)
}
