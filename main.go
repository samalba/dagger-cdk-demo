package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
)

func build(ctx context.Context, client *dagger.Client, registry *RegistryInfo) (string, error) {
	nodeCache := client.CacheVolume("node")

	// // Read the source code from local directory
	// sourceDir := client.Host().Directory(".", dagger.HostDirectoryOpts{
	// 	Exclude: []string{"node_modules/", "ci/"},
	// })

	// Read the source code from a remote git repository
	sourceDir := client.Git("https://github.com/dagger/hello-dagger.git").
		Commit("5343dfee12cfc59013a51886388a7cacee3f16b9").
		Tree().
		Directory(".")

	source := client.Container().
		From("node:16").
		WithMountedDirectory("/src", sourceDir).
		WithMountedCache("/src/node_modules", nodeCache)

	runner := source.WithWorkdir("/src").
		WithExec([]string{"npm", "install"})

	test := runner.WithExec([]string{"npm", "test", "--", "--watchAll=false"})

	buildDir := test.WithExec([]string{"npm", "run", "build"}).
		Directory("./build")

	// FIXME: This is a workaround until there is a better way to create a secret from the API
	registrySecret := client.Container().WithNewFile("/secret", dagger.ContainerWithNewFileOpts{
		Contents:    registry.password,
		Permissions: 0o400,
	}).File("/secret").Secret()

	return client.Container().
		From("nginx").
		WithDirectory("/usr/src/nginx", buildDir).
		WithRegistryAuth("125635003186.dkr.ecr.us-west-1.amazonaws.com", registry.username, registrySecret).
		Publish(ctx, registry.uri)
}

func main() {
	ctx := context.Background()

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	awsClient, err := NewAWSClient(ctx, "us-west-1")
	if err != nil {
		panic(err)
	}

	registry := initRegistry(ctx, client, awsClient)
	imageRef, err := build(ctx, client, registry)
	if err != nil {
		panic(err)
	}

	fmt.Println("Published image to", imageRef)

	// // initialize Dagger client
	// client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	// if err != nil {
	// 	panic(err)
	// }
	// defer client.Close()

	// alpine := client.Pipeline("TestPipeline").Container().From("alpine:3.17").WithExec([]string{"sh", "-c", "echo this is a test"})

	// stdout, err := alpine.Stdout(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(stdout)

	// deployToECS("amazon/amazon-ecs-sample")

}
