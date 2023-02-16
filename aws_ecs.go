package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	ecs_patterns "github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewECSStack(scope constructs.Construct, id string, containerImage string) cdk.Stack {
	stack := cdk.NewStack(scope, &id, &cdk.StackProps{
		Description: jsii.String("ECS/Fargate stack for dagger/CDK demo"),
		Env:         nil,
	},
	)

	// Create VPC and Cluster
	vpc := ec2.NewVpc(stack, jsii.String("ALBFargoVpc"), &ec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})

	cluster := ecs.NewCluster(stack, jsii.String("ALBFargoECSCluster"), &ecs.ClusterProps{
		Vpc: vpc,
	})

	res := ecs_patterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("ALBFargoService"), &ecs_patterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster: cluster,
		TaskImageOptions: &ecs_patterns.ApplicationLoadBalancedTaskImageOptions{
			Image: ecs.ContainerImage_FromRegistry(jsii.String(containerImage), &ecs.RepositoryImageProps{}),
		},
	})

	cdk.NewCfnOutput(stack, jsii.String("LoadBalancerDNS"), &cdk.CfnOutputProps{Value: res.LoadBalancer().LoadBalancerDnsName()})

	return stack
}

func deployToECS(containerImage string) {
	ctx := context.TODO()
	defer jsii.Close()

	app := cdk.NewApp(nil)

	NewECSStack(app, "AWSStack", containerImage)

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
