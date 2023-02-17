package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"

	"github.com/aws/aws-sdk-go-v2/config"
	sdk_ecr "github.com/aws/aws-sdk-go-v2/service/ecr"

	"github.com/aws/jsii-runtime-go"
)

func NewECRStack(id string, repositoryName string) (string, error) {
	defer jsii.Close()
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, &id, &cdk.StackProps{
		Description: jsii.String(fmt.Sprintf("ECR stack for repository %s", repositoryName)),
		Env:         nil,
	},
	)

	ecrRepo := ecr.NewRepository(stack, jsii.String(repositoryName), &ecr.RepositoryProps{
		RepositoryName: jsii.String(repositoryName),
	})
	cdk.NewCfnOutput(stack, jsii.String("RepositoryUri"), &cdk.CfnOutputProps{Value: ecrRepo.RepositoryUri()})

	cloudAsm := app.Synth(nil)
	cfnStack := cloudAsm.GetStackByName(jsii.String(id))
	j, err := json.Marshal(cfnStack.Template())
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func GetECRAuthorizationToken(ctx context.Context, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}

	cfg.Region = region
	client := sdk_ecr.NewFromConfig(cfg)

	log.Printf("ECR GetAuthorizationToken for region %q", region)
	out, err := client.GetAuthorizationToken(ctx, &sdk_ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	if len(out.AuthorizationData) < 1 {
		return "", fmt.Errorf("GetECRAuthorizationToken returned empty AuthorizationData")
	}

	authToken := *out.AuthorizationData[0].AuthorizationToken
	return authToken, nil
}
