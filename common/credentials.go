package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/viper"
)

type AwsCredentials struct {
	AwsAccessKeyId string
	AwsSecretKey   string
	AssumeRoleArn  string
	ClientRegion   string
}

func WithAwsCredentials() AwsCredentials {
	return AwsCredentials{
		ClientRegion:   viper.GetString("aws.clientRegion"),
		AwsAccessKeyId: viper.GetString("aws.staticCredentials.awsAccessKeyId"),
		AwsSecretKey:   viper.GetString("aws.staticCredentials.awsSecretKey"),
		AssumeRoleArn:  viper.GetString("aws.assumeRoleArn"),
	}
}

// assume role takes precedence over statc credentials if both set
func (c AwsCredentials) credentialsProvider() (config.LoadOptionsFunc, error) {
	if c.AssumeRoleArn != "" {
		sdkConfig, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
		stsClient := sts.NewFromConfig(sdkConfig)
		credProvider := config.WithCredentialsProvider(
			stscreds.NewAssumeRoleProvider(
				stsClient,
				c.AssumeRoleArn,
			),
		)
		return credProvider, nil
	}
	if c.AwsAccessKeyId != "" && c.AwsSecretKey != "" {
		credProvider := config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AwsAccessKeyId,
				c.AwsSecretKey, "",
			),
		)
		return credProvider, nil
	}
	return config.WithCredentialsProvider(nil), nil // default credential chain will be used
}

func (c AwsCredentials) AwsConfig() (aws.Config, error) {
	credProvider, err := c.credentialsProvider()
	if err != nil {
		return aws.Config{}, err
	}
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(c.ClientRegion), // empty string region will be ignored
		credProvider,
	)
}

func (c AwsCredentials) S3Client() (*s3.Client, error) {
	awsConfig, err := c.AwsConfig()
	if err != nil {
		return nil, fmt.Errorf("Couldn't load configuration. Error: %w\n", err)
	}
	s3Client := s3.NewFromConfig(awsConfig)
	return s3Client, nil
}
