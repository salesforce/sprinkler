// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/viper"
)

type AwsCredentials struct {
	ClientRegion   string
	AwsAccessKeyId string
	AwsSecretKey   string
	SessionToken   string
	AssumeRoleArn  string
}

func WithAwsCredentials() AwsCredentials {
	return AwsCredentials{
		ClientRegion:   viper.GetString("aws.clientRegion"),
		AwsAccessKeyId: viper.GetString("aws.staticCredentials.awsAccessKeyId"),
		AwsSecretKey:   viper.GetString("aws.staticCredentials.awsSecretKey"),
		SessionToken:   viper.GetString("aws.staticCredentials.sessionToken"),
		AssumeRoleArn:  viper.GetString("aws.assumeRoleArn"),
	}
}

func (c AwsCredentials) credentialsProvider() (config.LoadOptionsFunc, error) {
	var credProvider = config.WithCredentialsProvider(nil)
	if c.AwsAccessKeyId != "" && c.AwsSecretKey != "" {
		credProvider = config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AwsAccessKeyId,
				c.AwsSecretKey,
				c.SessionToken, // empty string will be ignored
			),
		)
	}
	if c.AssumeRoleArn != "" {
		awsConfig, err := config.LoadDefaultConfig(
			context.TODO(),
			credProvider, // use static credentials if provided
		)
		if err != nil {
			return nil, err
		}
		stsClient := sts.NewFromConfig(awsConfig)
		assumeRoleProvider := config.WithCredentialsProvider(
			stscreds.NewAssumeRoleProvider(
				stsClient,
				c.AssumeRoleArn,
			),
		)
		return assumeRoleProvider, nil
	}
	return credProvider, nil // default credentials chain if no aws configs set
}

func (c AwsCredentials) AwsConfig() (aws.Config, error) {
	credProvider, err := c.credentialsProvider()
	if err != nil {
		return aws.Config{}, err
	}
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(c.ClientRegion), // empty string will be ignored
		credProvider,
	)
}

func AWSClient[C any](
	cred AwsCredentials,
	createClient func(aws.Config) *C,
) (*C, error) {
	awsConfig, err := cred.AwsConfig()
	if err != nil {
		return nil, fmt.Errorf("Couldn't load configuration. Error: %w\n", err)
	}
	client := createClient(awsConfig)
	return client, nil
}

func (c AwsCredentials) S3Client() (*s3.Client, error) {
	return AWSClient(c, func(cfg aws.Config) *s3.Client { return s3.NewFromConfig(cfg) })
}

func (c AwsCredentials) SNSClient() (*sns.Client, error) {
	return AWSClient(c, func(cfg aws.Config) *sns.Client { return sns.NewFromConfig(cfg) })
}
