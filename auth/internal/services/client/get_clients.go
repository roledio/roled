package client

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	"github.com/roledio/roled/pkg/errors"
)

func (s *clientService) GetClients(ctx context.Context, req *models.GetClientsRequest) ([]models.ClientDetails, int, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, 0, err
	}
	clientRepo := s.registry.ClientRepository()
	count, err := clientRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count clients", "error", err)
		return nil, 0, errors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}
	clients, err := clientRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find clients", "error", err)
		return nil, 0, errors.ErrSystemError.WithError(err)
	}
	res := []models.ClientDetails{}
	for _, client := range clients {
		res = append(res, models.ClientDetails{
			ID:          client.ID,
			CreatedAt:   client.CreatedAt,
			UpdatedAt:   client.UpdatedAt,
			Name:        client.Name,
			Description: client.Description,
			IsActive:    client.IsActive,
			IsDefault:   client.IsDefault,
		})
	}
	return res, count, nil
}
