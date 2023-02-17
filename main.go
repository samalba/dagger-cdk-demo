package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
)

func build(ctx context.Context, registryURI, registryUsername, registryPassword string) (string, error) {
	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

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
		Contents:    registryPassword,
		Permissions: 0o400,
	}).File("/secret").Secret()

	return client.Container().
		From("nginx").
		WithDirectory("/usr/src/nginx", buildDir).
		WithRegistryAuth("125635003186.dkr.ecr.us-west-1.amazonaws.com", registryUsername, registrySecret).
		Publish(ctx, registryURI)
}

func main() {
	ctx := context.Background()

	ecrAuthToken, err := GetECRAuthorizationToken(ctx, "us-west-1")
	if err != nil {
		panic(err)
	}

	registryUser, registryPasswd, err := ECRTokenToUsernamePassword(ecrAuthToken)
	if err != nil {
		panic(err)
	}

	ecrStack, err := NewECRStack("TestECRStack", "dagger-cdk-demo")
	if err != nil {
		panic(err)
	}

	c, err := NewCfnClient(ctx, "us-west-1")
	if err != nil {
		panic(err)
	}

	stack, err := c.DeployStack(ctx, "TestECRStack", ecrStack, false)
	if err != nil {
		panic(err)
	}

	ecrOutputs := FormatStackOutputs(stack.Outputs)

	ref, err := build(ctx, ecrOutputs["RepositoryUri"], registryUser, registryPasswd)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pushed image to", ref)

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
