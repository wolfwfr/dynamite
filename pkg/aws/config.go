package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func MustLoadAWSConfig(ctx context.Context, region string, profile *string) *aws.Config {
	cfg, err := LoadAWSConfig(ctx, region, profile)
	if err != nil {
		panic(err)
	}
	return cfg
}

func LoadAWSConfig(ctx context.Context, region string, profile *string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading default config at %s: %w", region, err)
	}
	return &cfg, nil
}
