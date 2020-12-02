package repository

import (
	"bing/models"
	"context"
)



// Binger
// Describes the bing api behaviours.
type Binger interface {
	GetImage(ctx context.Context,cfg *models.APIConfig)(*models.APIResponse,error)
}