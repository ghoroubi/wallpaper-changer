package bing

import (
	"bing/models"
	"bing/repository"
	"context"
)

// This type will implements the binger interface.
type bingAPI struct {
	URL string `json:"url"`
}

// GetImage
// This method implements the fetch of daily bing image.
func (b *bingAPI) GetImage(ctx context.Context, cfg *models.APIConfig) (*models.APIResponse, error) {
	panic("implement me")
}

// NewBingAPI
// Creates an implementer of ing image api.
func NewBingAPI(url string) repository.Binger {
	return &bingAPI{URL: url}
}
