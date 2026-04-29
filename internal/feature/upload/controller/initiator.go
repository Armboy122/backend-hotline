package controller

import (
	"backend-hotlines3/internal/config"
	"backend-hotlines3/pkg/s3"
)

type Controller struct {
	r2Client *s3.R2Client
}

func NewController(cfg *config.Config) (*Controller, error) {
	r2Client, err := s3.NewR2Client(s3.R2Config{
		AccountID:       cfg.Cloudflare.R2.AccountID,
		AccessKeyID:     cfg.Cloudflare.R2.AccessKeyID,
		SecretAccessKey: cfg.Cloudflare.R2.SecretAccessKey,
		BucketName:      cfg.Cloudflare.R2.BucketName,
		PublicURL:       cfg.Cloudflare.R2.PublicURL,
	})
	if err != nil {
		return nil, err
	}

	return &Controller{r2Client: r2Client}, nil
}
