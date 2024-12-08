package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const ACCOUNT_ID = "992382576119"

// const ACCOUNT_ID = "000000000000"

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		createStepFunctionToComplexFlowExample(ctx)
		createStepFunctionToSyncMaltaFlights(ctx)
		return nil
	})
}
