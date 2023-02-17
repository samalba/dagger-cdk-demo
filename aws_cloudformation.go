package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/jsii-runtime-go"
)

type CfnClient struct {
	client *cloudformation.Client
}

func NewCfnClient(ctx context.Context, region string) (*CfnClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	client := &CfnClient{
		client: cloudformation.NewFromConfig(cfg),
	}

	return client, nil
}

// CreateStack deploys a CloudFormation stack from
func (c *CfnClient) CreateStack(ctx context.Context, stackName, templateBody string, iamCapability bool) error {
	caps := []types.Capability{}
	if iamCapability {
		caps = append(caps, types.CapabilityCapabilityIam)
	}

	_, err := c.client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    jsii.String(stackName),
		TemplateBody: jsii.String(templateBody),
		Capabilities: caps,
	})

	if err != nil {
		return err
	}

	return nil
}

// DeployStack creates the stack if needed, wait for it to deploy and returns its outputs
// FIXME: stack update not supported yet
func (c *CfnClient) DeployStack(ctx context.Context, stackName, templateBody string, iamCapability bool) (*types.Stack, error) {
	out, err := c.client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: jsii.String(stackName),
	})

	// Check if the stack already exists
	if err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			return nil, err
		}

		// Create the CFN stack
		log.Printf("Creating CloudFormation stack %q\n", stackName)
		err = c.CreateStack(ctx, stackName, templateBody, iamCapability)
		if err != nil {
			return nil, err
		}

		return c.DeployStack(ctx, stackName, templateBody, iamCapability)
	}

	if len(out.Stacks) < 1 {
		return nil, fmt.Errorf("cannot DescribeStack name %q", stackName)
	}

	stack := out.Stacks[0]
	status := string(stack.StackStatus)
	switch {
	case strings.HasSuffix(status, "_COMPLETE"):
		log.Printf("Stack %q already exists: %s\n", stackName, status)

	case strings.HasSuffix(status, "_IN_PROGRESS"):
		log.Printf("Waiting for stack %q to complete: %s\n", stackName, status)

		// Create a stack waiter
		waiter := cloudformation.NewStackCreateCompleteWaiter(c.client)

		// Wait for the stack create to complete (60 seconds is hardcoded)
		err = waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
			StackName: jsii.String(stackName),
		}, 60*time.Second)
		if err != nil {
			return nil, err
		}

		// Refresh the stack status to get the stack output
		return c.DeployStack(ctx, stackName, templateBody, iamCapability)

	case strings.HasSuffix(status, "_FAILED"):
		log.Printf("Stack %q is in FAILED state %q", stackName, status)
		return nil, fmt.Errorf("Stack %q is in FAILED state %q", stackName, status)

	default:
		log.Printf("Stack %q is in invalid state %q", stackName, status)
		return nil, fmt.Errorf("Stack %q is in invalid state %q", stackName, status)
	}

	return &stack, nil
}

// FormatStackOutputs converts stack outputs into a map of string for easy printing
func FormatStackOutputs(outputs []types.Output) map[string]string {
	outs := map[string]string{}

	for _, o := range outputs {
		outs[*o.OutputKey] = *o.OutputValue
	}

	return outs
}
