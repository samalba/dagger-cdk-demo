package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"

	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)

	NewECRStack(app, "DaggerDemoECRStack", "dagger-cdk-demo")

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
// func cdkEnvironment() *awscdk.Environment {
// 	return &awscdk.Environment{
// 		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
// 		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
// 	}
// }
