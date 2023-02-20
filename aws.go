package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type AWSClient struct {
	region string
	cCfn   *cloudformation.Client
	cEcr   *ecr.Client
}

func NewAWSClient(ctx context.Context, region string) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	client := &AWSClient{
		region: region,
		cCfn:   cloudformation.NewFromConfig(cfg),
		cEcr:   ecr.NewFromConfig(cfg),
	}

	return client, nil
}

func (c *AWSClient) GetCfnStackOutputs(ctx context.Context, stackName string) (map[string]string, error) {
	panic("not implemented")

	// FIXME: implement
	// out, err := c.client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
	// 	StackName: jsii.String(stackName),
	// })
	// stack := out.Stacks[0]
	// status := string(stack.StackStatus)

	return nil, nil
}

func (c *AWSClient) GetECRAuthorizationToken(ctx context.Context) (string, error) {
	log.Printf("ECR GetAuthorizationToken for region %q", c.region)
	out, err := c.cEcr.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	if len(out.AuthorizationData) < 1 {
		return "", fmt.Errorf("GetECRAuthorizationToken returned empty AuthorizationData")
	}

	authToken := *out.AuthorizationData[0].AuthorizationToken
	return authToken, nil
}

// Converts an ECR auth token to username / password
func ECRTokenToUsernamePassword(token string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}

	split := strings.SplitN(string(decoded), ":", 2)
	if len(split) < 1 {
		return "", "", fmt.Errorf("invalid base64 decoded data")
	}

	return split[0], split[1], nil
}

// FormatStackOutputs converts stack outputs into a map of string for easy printing
func FormatStackOutputs(outputs []types.Output) map[string]string {
	outs := map[string]string{}

	for _, o := range outputs {
		outs[*o.OutputKey] = *o.OutputValue
	}

	return outs
}
