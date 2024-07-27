package components

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/jsii-runtime-go"
)

func CreateDynamoDBTables(stack awscdk.Stack) {

	usersTableProps := &awsdynamodb.TablePropsV2{
		TableName: jsii.String("users"),
		Billing:   awsdynamodb.Billing_OnDemand(),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("user_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	}

	// API Keys Table Fields
	apiKeysTableProps := &awsdynamodb.TablePropsV2{
		TableName: jsii.String("api_keys"),
		Billing:   awsdynamodb.Billing_OnDemand(),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("api_key"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
			{
				IndexName: jsii.String("user_id-index"),
				PartitionKey: &awsdynamodb.Attribute{
					Name: jsii.String("user_id"),
					Type: awsdynamodb.AttributeType_STRING,
				},
				ProjectionType: awsdynamodb.ProjectionType_ALL,
			},
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	}

	// Transactions Table Fields
	transactionsTableProps := &awsdynamodb.TablePropsV2{
		TableName: jsii.String("transactions"),
		Billing:   awsdynamodb.Billing_OnDemand(),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("transaction_id"),
			Type: awsdynamodb.AttributeType_NUMBER,
		},
		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
			{
				IndexName: jsii.String("user_id-index"),
				PartitionKey: &awsdynamodb.Attribute{
					Name: jsii.String("user_id"),
					Type: awsdynamodb.AttributeType_STRING,
				},
				ProjectionType: awsdynamodb.ProjectionType_ALL,
			},
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	}

	// Create Tables

	awsdynamodb.NewTableV2(stack, jsii.String("UsersTable"), usersTableProps)
	awsdynamodb.NewTableV2(stack, jsii.String("ApiKeysTable"), apiKeysTableProps)
	awsdynamodb.NewTableV2(stack, jsii.String("TransactionsTable"), transactionsTableProps)

}
