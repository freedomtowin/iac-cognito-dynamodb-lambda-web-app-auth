package components

import (
	"os"
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/jsii-runtime-go"
)

type StackConfigs struct {
	ImageFolder string
	ApiName     string
}

type LambdaImageStackProps struct {
	awscdk.StackProps
	stackDetails StackConfigs
}

func NewLambdaImageDeployStack(stack awscdk.Stack, userPool awscognito.UserPool, imageFolder string, apiName string) {

	dir, _ := os.Getwd()

	ecr_image := awslambda.EcrImageCode_FromAssetImage(jsii.String(filepath.Join(dir, imageFolder)),
		&awslambda.AssetImageCodeProps{},
	)

	// create AmazonDynamoDBFullAccess role
	dynamoDBRole := awsiam.NewRole(stack, aws.String("myDynamoDBFullAccessRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(aws.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromManagedPolicyArn(stack, aws.String("AmazonDynamoDBFullAccess"), aws.String("arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess")),
		},
	})

	// Create Lambda function
	lambdaFn := awslambda.NewFunction(stack, jsii.String("lambdaFromImage"), &awslambda.FunctionProps{
		Code: ecr_image,
		// Handler and Runtime must be *FROM_IMAGE* when provisioning Lambda from container.
		Handler:      awslambda.Handler_FROM_IMAGE(),
		Runtime:      awslambda.Runtime_FROM_IMAGE(),
		FunctionName: jsii.String(apiName),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Role:         dynamoDBRole,
	})

	lambdaFn.AddAlias(jsii.String("Live"), &awslambda.AliasOptions{})

	// Create a Cognito Authorizer
	authorizer := awsapigateway.NewCognitoUserPoolsAuthorizer(stack, jsii.String("Authorizer"), &awsapigateway.CognitoUserPoolsAuthorizerProps{
		CognitoUserPools: &[]awscognito.IUserPool{userPool},
	})

	// Create an API Gateway
	restApi := awsapigateway.NewRestApi(stack, jsii.String("myRESTApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String(apiName),
		Deploy:      jsii.Bool(true),
	})

	// Add MethodResponse to MethodOptions
	methodOptions := &awsapigateway.MethodOptions{
		MethodResponses: &[]*awsapigateway.MethodResponse{
			{
				StatusCode: jsii.String("200"), // Specify the HTTP status code for the response
				ResponseParameters: &map[string]*bool{
					"method.response.header.Access-Control-Allow-Origin": jsii.Bool(true), // Add the CORS header
				},
				ResponseModels: &map[string]awsapigateway.IModel{
					"application/json": awsapigateway.Model_EMPTY_MODEL(), // Specify JSON as the response content type
				},
			},
		},
		AuthorizationType: awsapigateway.AuthorizationType_COGNITO,
		Authorizer:        authorizer,
	}

	integrationResponse := &[]*awsapigateway.IntegrationResponse{
		{
			StatusCode: jsii.String("200"),
			ResponseParameters: &map[string]*string{
				"method.response.header.Access-Control-Allow-Origin": jsii.String("'*'"),
			},
		},
	}

	integrationOptions := &awsapigateway.LambdaIntegrationOptions{
		IntegrationResponses: integrationResponse,
		Proxy:                jsii.Bool(false),
	}

	// Create a resource and add method
	corrEndpoint := restApi.Root().AddResource(jsii.String("correlation"), &awsapigateway.ResourceOptions{
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		},
	})

	corrEndpoint.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(lambdaFn, integrationOptions), methodOptions)

	awscdk.NewCfnOutput(stack, jsii.String("myRESTApiEndpoint"), &awscdk.CfnOutputProps{
		Value:       restApi.Url(),
		Description: jsii.String("REST API Endpoint"),
	})

}
