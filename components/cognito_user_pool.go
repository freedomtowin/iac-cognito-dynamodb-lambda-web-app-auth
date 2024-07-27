package components

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/jsii-runtime-go"
)

func CreateCognitoUserPool(stack awscdk.Stack) awscognito.UserPool {

	// Create a Cognito User Pool
	userPool := awscognito.NewUserPool(stack, jsii.String("UserPool"), &awscognito.UserPoolProps{
		SelfSignUpEnabled: jsii.Bool(true),
		SignInAliases: &awscognito.SignInAliases{
			Email: jsii.Bool(true),
		},
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:        jsii.Number(8),
			RequireSymbols:   jsii.Bool(true),
			RequireDigits:    jsii.Bool(true),
			RequireUppercase: jsii.Bool(true),
			RequireLowercase: jsii.Bool(true),
		},
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
	})

	readAccessScope := awscognito.NewResourceServerScope(&awscognito.ResourceServerScopeProps{
		ScopeName:        jsii.String("read"),
		ScopeDescription: jsii.String("Read access"),
	})

	writeAccessScope := awscognito.NewResourceServerScope(&awscognito.ResourceServerScopeProps{
		ScopeName:        jsii.String("write"),
		ScopeDescription: jsii.String("Write access"),
	})

	// Create a Resource Server
	resource_name := "probablycraterapi"
	resourceServer := userPool.AddResourceServer(jsii.String("ResourceServer"), &awscognito.UserPoolResourceServerOptions{
		Identifier: jsii.String(resource_name),
		Scopes: &[]awscognito.ResourceServerScope{
			readAccessScope,
			writeAccessScope,
		},
	})

	// Create a User Pool Client
	userPoolClient := userPool.AddClient(jsii.String("UserPoolClient"), &awscognito.UserPoolClientOptions{
		GenerateSecret: jsii.Bool(false),
		AuthFlows: &awscognito.AuthFlow{
			UserPassword: jsii.Bool(true),
			UserSrp:      jsii.Bool(true),
		},
		OAuth: &awscognito.OAuthSettings{
			Flows: &awscognito.OAuthFlows{
				AuthorizationCodeGrant: jsii.Bool(true),
			},
			Scopes: &[]awscognito.OAuthScope{
				awscognito.OAuthScope_OPENID(),
				awscognito.OAuthScope_EMAIL(),
				awscognito.OAuthScope_PROFILE(),
				awscognito.OAuthScope_ResourceServer(resourceServer, readAccessScope),
				awscognito.OAuthScope_ResourceServer(resourceServer, writeAccessScope),
			},
			CallbackUrls: &[]*string{
				jsii.String("https://my-app-domain.com/callback"),
			},
			LogoutUrls: &[]*string{
				jsii.String("https://my-app-domain.com/signout"),
			},
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolID"), &awscdk.CfnOutputProps{
		Value:       userPool.UserPoolId(),
		Description: jsii.String("User Pool Client ID"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolClientID"), &awscdk.CfnOutputProps{
		Value:       userPoolClient.UserPoolClientId(),
		Description: jsii.String("User Pool Client ID"),
	})

	return userPool
}
