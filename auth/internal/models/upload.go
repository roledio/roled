package models

import "mime/multipart"

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" validate:"required"`
	Type string                `form:"type" validate:"required,oneof=project-logo user-avatar"`
}

type UploadFileResponse struct {
	URL string `json:"url"`
}
