package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"io"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

// Favicons accept the same raster formats as logos, plus ICO containers.
type faviconValidator struct{}

func (v *faviconValidator) Validate(ctx context.Context, req *models.UploadFileRequest) (string, string, error) {
	if req.File.Size > maxProjectLogoSize {
		return "", "", pkgerrors.ErrFileSizeTooLarge
	}
	file, err := req.File.Open()
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.WithContext(ctx).Errorw("Failed to close favicon upload", "error", err)
		}
	}()
	content, err := io.ReadAll(io.LimitReader(file, maxProjectLogoSize+1))
	if err != nil {
		return "", "", err
	}
	if len(content) > maxProjectLogoSize {
		return "", "", pkgerrors.ErrFileSizeTooLarge
	}
	if len(content) >= 4 && content[0] == 0 && content[1] == 0 && content[2] == 1 && content[3] == 0 {
		if !validICO(content) {
			return "", "", errors.ErrInvalidImageType
		}
		return ".ico", "image/x-icon", nil
	}
	return validateImage(ctx, req)
}

// Validate directory and image bounds, including embedded PNG or Windows DIB headers.
func validICO(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count == 0 || 6+count*16 > len(data) {
		return false
	}
	for i := 0; i < count; i++ {
		entry := data[6+i*16 : 6+(i+1)*16]
		size := uint64(binary.LittleEndian.Uint32(entry[8:12]))
		offset := uint64(binary.LittleEndian.Uint32(entry[12:16]))
		if size < 8 || offset < uint64(6+count*16) || offset+size > uint64(len(data)) {
			return false
		}
		payload := data[offset : offset+size]
		width, height := int(entry[0]), int(entry[1])
		if width == 0 {
			width = 256
		}
		if height == 0 {
			height = 256
		}
		if string(payload[:8]) == "\x89PNG\r\n\x1a\n" {
			config, format, err := image.DecodeConfig(bytes.NewReader(payload))
			if err != nil || format != "png" || config.Width != width || config.Height != height {
				return false
			}
		} else {
			if len(payload) < 40 || binary.LittleEndian.Uint32(payload[:4]) < 40 || int(binary.LittleEndian.Uint32(payload[4:8])) != width || int(binary.LittleEndian.Uint32(payload[8:12])) != height*2 {
				return false
			}
			bits := int(binary.LittleEndian.Uint16(payload[14:16]))
			switch bits {
			case 1, 4, 8, 16, 24, 32:
			default:
				return false
			}
			if binary.LittleEndian.Uint16(payload[12:14]) != 1 || binary.LittleEndian.Uint32(payload[16:20]) != 0 {
				return false
			}
			header := int(binary.LittleEndian.Uint32(payload[:4]))
			palette := 0
			if bits <= 8 {
				palette = (1 << bits) * 4
			}
			pixelBytes := ((width*bits + 31) / 32) * 4 * height
			if header+palette+pixelBytes > len(payload) {
				return false
			}
		}
	}
	return true
}
