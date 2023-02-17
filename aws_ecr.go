package main

import (
	"encoding/json"
	"fmt"

	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"

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
