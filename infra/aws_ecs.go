package main

import (
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	ecs_patterns "github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewECSStack(scope constructs.Construct, id string, containerImage string) cdk.Stack {
	stack := cdk.NewStack(scope, &id, &cdk.StackProps{
		Description: jsii.String("ECS/Fargate stack for dagger/CDK demo"),
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
