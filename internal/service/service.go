package service

import (
	"context"

	"github.com/gofreego/mediabase/api/mediabase_v1"
	"github.com/gofreego/mediabase/internal/storage"
)

type Config struct {
	StorageConfig       storage.Config
	MaxFileSize         int64    `yaml:"MaxFileSize"`
	AllowedContentTypes []string `yaml:"AllowedContentTypes"`
}

type Service struct {
	storage             storage.Storage
	maxFileSize         int64
	allowedContentTypes map[string]bool
	mediabase_v1.UnimplementedMediabaseServiceServer
}

func NewService(ctx context.Context, cfg *Config, storageProvider storage.Storage) *Service {
	allowedMap := make(map[string]bool)
	for _, ct := range cfg.AllowedContentTypes {
		allowedMap[ct] = true
	}

	return &Service{
		storage:             storageProvider,
		maxFileSize:         cfg.MaxFileSize,
		allowedContentTypes: allowedMap,
	}
}
