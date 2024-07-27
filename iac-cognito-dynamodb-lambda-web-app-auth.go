// main.go

package main

import (
	"iac-cognito-dynamodb-lambda-web-app-auth/components"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StackConfigs struct {
	ImageFolder string
	ApiName     string
}

type MyCdkStackProps struct {
	awscdk.StackProps
	stackDetails StackConfigs
}

func NewMyCdkStack(scope constructs.Construct, id string, props *MyCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	imageFolder := props.stackDetails.ImageFolder
	apiName := props.stackDetails.ApiName

	// Call the function to create DynamoDB tables
	components.CreateDynamoDBTables(stack)

	userPool := components.CreateCognitoUserPool(stack)

	components.NewLambdaImageDeployStack(stack, userPool, imageFolder, apiName)

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewMyCdkStack(app, "ProbablyCrater", &MyCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		stackDetails: StackConfigs{
			ImageFolder: "lambda",
			ApiName:     "probablyAPI",
		},
	})

	app.Synth(nil)

}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
