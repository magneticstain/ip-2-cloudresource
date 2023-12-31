package awsconnector

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSConnector struct {
	AwsConfig aws.Config
}

func New() (AWSConnector, error) {
	cfg, err := ConnectToAWS("", aws.Config{})

	ac := AWSConnector{AwsConfig: cfg}

	return ac, err
}

func NewAWSConnectorAssumeRole(roleArn string, baseConfig aws.Config) (AWSConnector, error) {
	cfg, err := ConnectToAWS(roleArn, baseConfig)

	ac := AWSConnector{AwsConfig: cfg}

	return ac, err
}

func ConnectToAWS(roleArn string, baseConfig aws.Config) (aws.Config, error) {
	var cfg aws.Config
	var err error

	if baseConfig.Region != "" {
		cfg = baseConfig
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return cfg, err
		}
	}

	// there doesn't appear to be an easy way to set a custom user-agent for aws.Config in v2
	// however, the App ID is apparently always added to the user-agent ( see issue #295 )
	cfg.AppID = "ip-2-cloudresource"

	if roleArn != "" {
		// assume role and override cfg creds with sts creds
		// REF: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/stscreds
		stsSvc := sts.NewFromConfig(cfg)
		roleCreds := stscreds.NewAssumeRoleProvider(stsSvc, roleArn)
		cfg.Credentials = aws.NewCredentialsCache(roleCreds)
	}

	return cfg, nil
}
