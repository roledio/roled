package upload

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/stretchr/testify/require"
)

func TestFaviconValidator(t *testing.T) {
	icon, err := os.ReadFile("../../views/assets/static/favicon.ico")
	require.NoError(t, err)
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"valid ICO", icon, true},
		{"truncated ICO", icon[:20], false},
		{"html disguised as icon", []byte("<script>alert(1)</script>"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("file", "favicon.ico")
			require.NoError(t, err)
			_, err = part.Write(tc.data)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			req := httptest.NewRequest("POST", "/upload", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			require.NoError(t, req.ParseMultipartForm(3*1024*1024))
			defer func() { require.NoError(t, req.MultipartForm.RemoveAll()) }()
			validator, err := newUploadTypeValidator("favicon")
			require.NoError(t, err)
			ext, mime, err := validator.Validate(context.Background(), &models.UploadFileRequest{Type: "favicon", File: req.MultipartForm.File["file"][0]})
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, ".ico", ext)
			require.Equal(t, "image/x-icon", mime)
		})
	}
}
