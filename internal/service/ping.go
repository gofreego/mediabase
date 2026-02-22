package service

import (
	"context"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/mediabase/api/mediabase_v1"
)

func (s *Service) Ping(ctx context.Context, req *mediabase_v1.PingRequest) (*mediabase_v1.PingResponse, error) {
	logger.Debug(ctx, "Ping request received, %v", req.Message)
	return &mediabase_v1.PingResponse{
		Message: "Its fine here...!",
	}, nil
}
