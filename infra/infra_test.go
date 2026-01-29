package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestStackFoundryWebsiteStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewStackFoundryWebsiteStack(app, "MyTestStack", &StackFoundryWebsiteStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("eu-west-2"),
			},
		},
	})

	// THEN
	template := assertions.Template_FromStack(stack, nil)

	// 1. Verify Lambda Function Configuration
	// Updated expectation: Timeout is now 5 seconds
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
		"Runtime":    "provided.al2023",
		"Handler":    "bootstrap",
		"MemorySize": 128,
		"Timeout":    5, // FIX: Updated from 10 to 5 to match main.go
		"Architectures": []interface{}{
			"arm64",
		},
	})

	// Verify Log Group exists with correct retention
	template.HasResourceProperties(jsii.String("AWS::Logs::LogGroup"), map[string]interface{}{
		"RetentionInDays": 7,
	})

	// 2. Verify Lambda IAM Permissions (SES)
	// Check that we have a policy allowing ses:SendEmail
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []interface{}{
						"ses:SendEmail",
						"ses:SendRawEmail",
					},
					"Effect":   "Allow",
					"Resource": "*",
				},
			},
		},
	})

	// 3. Verify API Gateway (HTTP API)
	template.HasResourceProperties(jsii.String("AWS::ApiGatewayV2::Api"), map[string]interface{}{
		"ProtocolType": "HTTP",
		"Name":         "StackFoundryAPI",
	})

	// 4. Verify Budget (The Cost Tripwire)
	// Critical: Ensure the limit is exactly $2.00 and notifies the correct email
	template.HasResourceProperties(jsii.String("AWS::Budgets::Budget"), map[string]interface{}{
		"Budget": map[string]interface{}{
			"BudgetLimit": map[string]interface{}{
				"Amount": 2,
				"Unit":   "USD",
			},
			"BudgetType": "COST",
			"TimeUnit":   "MONTHLY",
		},
		"NotificationsWithSubscribers": []interface{}{
			map[string]interface{}{
				"Subscribers": []interface{}{
					map[string]interface{}{
						"Address":          "joe@stackfoundry.co.uk",
						"SubscriptionType": "EMAIL",
					},
				},
			},
		},
	})

	// 5. Verify Hosted Zone
	template.HasResourceProperties(jsii.String("AWS::Route53::HostedZone"), map[string]interface{}{
		"Name": "stackfoundry.co.uk.",
	})
}
