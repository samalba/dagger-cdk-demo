package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"

	"github.com/aws/jsii-runtime-go"
)

func NewECRStack(id string, containerImage string) (string, error) {
	defer jsii.Close()
	app := cdk.NewApp(nil)

	stack := cdk.NewStack(app, &id, &cdk.StackProps{
		Description: jsii.String("ECR stack for dagger/CDK demo"),
		Env:         nil,
	},
	)

	ecrRepo := ecr.NewRepository(stack, jsii.String("dagger-cdk-demo"), &ecr.RepositoryProps{})

	cdk.NewCfnOutput(stack, jsii.String("LoadBalancerDNS"), &cdk.CfnOutputProps{Value: ecrRepo.RepositoryUri()})

	cloudAsm := app.Synth(nil)
	cfnStack := cloudAsm.GetStackByName(jsii.String("AWSStack"))
	j, err := json.Marshal(cfnStack.Template())
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func deployECR(containerImage string) {
	ctx := context.TODO()
	defer jsii.Close()

	app := cdk.NewApp(nil)

	NewECRStack(app, "AWSStack", containerImage)

	cloudAsm := app.Synth(nil)
	stack := cloudAsm.GetStackByName(jsii.String("AWSStack"))
	j, err := json.Marshal(stack.Template())
	if err != nil {
		panic(err)
	}
	templateBody := string(j)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// override with the region where I ran "cdk bootstrap"
	cfg.Region = "us-west-1"

	client := cloudformation.NewFromConfig(cfg)

	output, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    jsii.String("SamTest"),
		TemplateBody: jsii.String(templateBody),
		Capabilities: []types.Capability{types.CapabilityCapabilityIam},
	})

	if err != nil {
		panic(err)
	}
	fmt.Printf("CreateStack -> %#v\n", output)
}
