package main

import (
	"encoding/json"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/sfn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createStepFunctionToComplexFlowExample(ctx *pulumi.Context) {
	createLambda(ctx)

	stateMachineDefinition := map[string]interface{}{
		"Comment": "A description of my state machine",
		"StartAt": "Choice",
		"States": map[string]interface{}{
			"Choice": map[string]interface{}{
				"Type": "Choice",
				"Choices": []map[string]interface{}{
					{
						"Variable":     "$.key1",
						"StringEquals": "CASCATA",
						"Next":         "Lambda Invoke (5)",
					},
				},
				"Default": "Parallel",
			},
			"Lambda Invoke (5)": map[string]interface{}{
				"Type":     "Task",
				"Resource": "arn:aws:states:::lambda:invoke",
				"Parameters": map[string]interface{}{
					"Payload.$":    "$",
					"FunctionName": "arn:aws:lambda:us-east-1:" + ACCOUNT_ID + ":function:my_function_3:$LATEST",
				},
				"Retry": []map[string]interface{}{
					{
						"ErrorEquals": []string{
							"Lambda.ServiceException",
							"Lambda.AWSLambdaException",
							"Lambda.SdkClientException",
							"Lambda.TooManyRequestsException",
						},
						"IntervalSeconds": 1,
						"MaxAttempts":     3,
						"BackoffRate":     2,
					},
				},
				"ResultPath": nil,
				"Next":       "SNS Publish (1)",
			},
			"SNS Publish (1)": map[string]interface{}{
				"Type":     "Task",
				"Resource": "arn:aws:states:::sns:publish",
				"Parameters": map[string]interface{}{
					"Message.$": "$",
					"TopicArn":  "arn:aws:sns:us-east-1:" + ACCOUNT_ID + ":my_topic_1",
				},
				"End":        true,
				"ResultPath": nil,
			},
			"Parallel": map[string]interface{}{
				"Type": "Parallel",
				"Branches": []map[string]interface{}{
					{
						"StartAt": "Lambda Invoke",
						"States": map[string]interface{}{
							"Lambda Invoke": map[string]interface{}{
								"Type":     "Task",
								"Resource": "arn:aws:states:::lambda:invoke",
								"Parameters": map[string]interface{}{
									"Payload.$":    "$",
									"FunctionName": "arn:aws:lambda:us-east-1:" + ACCOUNT_ID + ":function:my_function_2:$LATEST",
								},
								"Retry": []map[string]interface{}{
									{
										"ErrorEquals": []string{
											"Lambda.ServiceException",
											"Lambda.AWSLambdaException",
											"Lambda.SdkClientException",
											"Lambda.TooManyRequestsException",
										},
										"IntervalSeconds": 1,
										"MaxAttempts":     3,
										"BackoffRate":     2,
									},
								},
								"ResultPath": nil,
								"Next":       "Success",
							},
							"Success": map[string]interface{}{
								"Type": "Succeed",
							},
						},
					},
					{
						"StartAt": "Lambda Invoke (1)",
						"States": map[string]interface{}{
							"Lambda Invoke (1)": map[string]interface{}{
								"Type":     "Task",
								"Resource": "arn:aws:states:::lambda:invoke",
								"Parameters": map[string]interface{}{
									"Payload.$":    "$",
									"FunctionName": "arn:aws:lambda:us-east-1:" + ACCOUNT_ID + ":function:my_function_3:$LATEST",
								},
								"Retry": []map[string]interface{}{
									{
										"ErrorEquals": []string{
											"Lambda.ServiceException",
											"Lambda.AWSLambdaException",
											"Lambda.SdkClientException",
											"Lambda.TooManyRequestsException",
										},
										"IntervalSeconds": 1,
										"MaxAttempts":     3,
										"BackoffRate":     2,
									},
								},
								"ResultPath": nil,
								"Next":       "Success (1)",
							},
							"Success (1)": map[string]interface{}{
								"Type": "Succeed",
							},
						},
					},
				},
				"Next":       "SNS Publish",
				"ResultPath": nil,
			},
			"SNS Publish": map[string]interface{}{
				"Type":     "Task",
				"Resource": "arn:aws:states:::sns:publish",
				"Parameters": map[string]interface{}{
					"Message.$": "$",
					"TopicArn":  "arn:aws:sns:us-east-1:" + ACCOUNT_ID + ":my_topic_1",
				},
				"End": true,
			},
		},
	}

	definition, err := json.Marshal(stateMachineDefinition)
	if err != nil {
		return
	}

	_, err = sfn.NewStateMachine(ctx, "My_Demo_Machine", &sfn.StateMachineArgs{
		Definition: pulumi.String(definition),
		RoleArn:    pulumi.String("arn:aws:iam::" + ACCOUNT_ID + ":role/service-role/StepFunctions-MyStateMachine-0u3iwvjku-role-mfdufnszh"),
	})
	if err != nil {
		return
	}

}

func createLambda(ctx *pulumi.Context) {
	role, err := iam.NewRole(ctx, "lambdaRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}]
		}`),
	})
	if err != nil {
		return
	}

	policy, err := iam.NewPolicy(ctx, "lambdaPolicy", &iam.PolicyArgs{
		Policy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": [
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Resource": "arn:aws:logs:*:*:*"
			}]
		}`),
	})
	if err != nil {
		return
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "lambdaRoleAttachment", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: policy.Arn,
	})
	if err != nil {
		return
	}

	function, err := lambda.NewFunction(ctx, "myLambdaFunction", &lambda.FunctionArgs{
		Runtime: pulumi.String("provided.al2"),
		Role:    role.Arn,
		Handler: pulumi.String("bootstrap"),
		Code:    pulumi.NewFileArchive("./deployment.zip"),
	})
	if err != nil {
		return
	}

	ctx.Export("lambdaFunctionName", function.Name)
}
