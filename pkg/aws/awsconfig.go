package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

func MustLoadAWSConfig(ctx context.Context, region string, profile *string) *aws.Config {
	cfg, err := LoadAWSConfig(ctx, region, profile)
	if err != nil {
		panic(err)
	}
	return cfg
}

func LoadAWSConfig(ctx context.Context, region string, profile *string) (*aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0, 1)
	opts = append(opts, config.WithRegion(region))
	if profile != nil && *profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(*profile))
		opts = append(opts, config.WithAssumeRoleCredentialOptions(func(roleOpts *stscreds.AssumeRoleOptions) {
			roleOpts.TokenProvider = stscreds.StdinTokenProvider
		}))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading default config at %s: %w", region, err)
	}
	return &cfg, nil
}
