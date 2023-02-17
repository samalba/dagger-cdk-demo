package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"

	"dagger.io/dagger"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
)

func build() {
	ctx := context.Background()

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	nodeCache := client.CacheVolume("node")

	hostSourceDir := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"node_modules/", "ci/"},
	})

	source := client.Container().
		From("node:16").
		WithMountedDirectory("/src", hostSourceDir).
		WithMountedCache("/src/node_modules", nodeCache)

	runner := source.WithWorkdir("/src").
		WithExec([]string{"npm", "install"})

	test := runner.WithExec([]string{"npm", "test", "--", "--watchAll=false"})

	buildDir := test.WithExec([]string{"npm", "run", "build"}).
		Directory("./build")

	ref, err := client.Container().
		From("nginx").
		WithDirectory("/usr/src/nginx", buildDir).
		Publish(ctx, fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000))) //#nosec
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to: %s\n", ref)
}

func main() {
	ctx := context.Background()

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
	fmt.Println("Outputs:", FormatStackOutputs(stack.Outputs))

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
