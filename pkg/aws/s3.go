// Package aws contains the ports or interefaces for the s3 module
package aws

import (
	"context"
	"log"
	"selfier/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3(cfg *config.AWSConfig) *s3.Client {

	var configOptions []func(*awsconfig.LoadOptions) error

	configOptions = append(configOptions, awsconfig.WithRegion(cfg.Region))

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
		configOptions = append(configOptions, awsconfig.WithCredentialsProvider(creds))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), configOptions...)
	if err != nil {
		log.Printf("Error loading AWS configuration: %v", err)
		panic(err)
	}

	if cfg.EndpointURL != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.EndpointURL)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Use path-style addressing, required for services like MinIO
		o.UsePathStyle = cfg.S3ForcePathStyle
	})

	return client
}
