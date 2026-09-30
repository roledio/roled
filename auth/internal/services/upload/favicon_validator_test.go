package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/png"
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

func TestICOImageBoundsAndDIBValidation(t *testing.T) {
	wrap := func(width, height byte, payload []byte) []byte {
		data := make([]byte, 22+len(payload))
		copy(data, []byte{0, 0, 1, 0, 1, 0})
		data[6], data[7] = width, height
		binary.LittleEndian.PutUint32(data[14:18], uint32(len(payload)))
		binary.LittleEndian.PutUint32(data[18:22], 22)
		copy(data[22:], payload)
		return data
	}
	dib := func(bits uint16) []byte {
		palette := 0
		if bits <= 8 {
			palette = (1 << bits) * 4
		}
		payload := make([]byte, 40+palette+4)
		binary.LittleEndian.PutUint32(payload[0:4], 40)
		binary.LittleEndian.PutUint32(payload[4:8], 1)
		binary.LittleEndian.PutUint32(payload[8:12], 2)
		binary.LittleEndian.PutUint16(payload[12:14], 1)
		binary.LittleEndian.PutUint16(payload[14:16], bits)
		return payload
	}
	for _, bits := range []uint16{1, 4, 8, 16, 24, 32} {
		require.True(t, validICO(wrap(1, 1, dib(bits))), "valid bit depth %d", bits)
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"short directory", func(d []byte) []byte { return d[:5] }},
		{"zero images", func(d []byte) []byte { d[4] = 0; return d }},
		{"directory extends past file", func(d []byte) []byte { d[4] = 255; return d }},
		{"payload too small", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[14:18], 7); return d }},
		{"payload overlaps directory", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[18:22], 6); return d }},
		{"payload past end", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[18:22], ^uint32(0)); return d }},
		{"truncated DIB", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[14:18], 8); return d[:30] }},
		{"short DIB header", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[22:26], 12); return d }},
		{"incorrect width", func(d []byte) []byte { d[6] = 2; return d }},
		{"incorrect height", func(d []byte) []byte { d[7] = 2; return d }},
		{"unsupported bit depth", func(d []byte) []byte { binary.LittleEndian.PutUint16(d[36:38], 2); return d }},
		{"invalid planes", func(d []byte) []byte { binary.LittleEndian.PutUint16(d[34:36], 2); return d }},
		{"compressed bitmap", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[38:42], 1); return d }},
		{"missing palette", func(d []byte) []byte { binary.LittleEndian.PutUint16(d[36:38], 8); return d }},
		{"header exceeds payload", func(d []byte) []byte { binary.LittleEndian.PutUint32(d[22:26], 100); return d }},
	} {
		t.Run(tc.name, func(t *testing.T) { require.False(t, validICO(tc.mutate(wrap(1, 1, dib(32))))) })
	}
	var small, large bytes.Buffer
	require.NoError(t, png.Encode(&small, image.NewRGBA(image.Rect(0, 0, 1, 1))))
	require.NoError(t, png.Encode(&large, image.NewRGBA(image.Rect(0, 0, 256, 256))))
	require.True(t, validICO(wrap(1, 1, small.Bytes())))
	require.True(t, validICO(wrap(0, 0, large.Bytes())), "zero dimension means 256 pixels")
	require.False(t, validICO(wrap(2, 1, small.Bytes())), "directory must match PNG dimensions")
	require.False(t, validICO(wrap(1, 1, []byte("\x89PNG\r\n\x1a\n"))), "PNG signature alone is not an image")
}
