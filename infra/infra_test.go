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
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]interface{}{
		"Runtime":    "provided.al2023",
		"Handler":    "bootstrap",
		"MemorySize": 128,
		"Timeout":    5,
		"Architectures": []interface{}{
			"arm64",
		},
	})

	// 2. Verify API Gateway Payload Format
	template.HasResourceProperties(jsii.String("AWS::ApiGatewayV2::Integration"), map[string]interface{}{
		"PayloadFormatVersion": "1.0",
		"IntegrationType":      "AWS_PROXY",
	})

	// 3. Verify SSL Certificate (Root + WWW Support)
	template.HasResourceProperties(jsii.String("AWS::CertificateManager::Certificate"), map[string]interface{}{
		"DomainName": "stackfoundry.co.uk",
		"SubjectAlternativeNames": []interface{}{
			"www.stackfoundry.co.uk",
		},
	})

	// 4. Verify Custom Domains (Root + WWW)
	// We expect 2 domains: stackfoundry.co.uk AND www.stackfoundry.co.uk
	template.ResourceCountIs(jsii.String("AWS::ApiGatewayV2::DomainName"), jsii.Number(2))

	// Verify mappings connect both domains to the API
	template.ResourceCountIs(jsii.String("AWS::ApiGatewayV2::ApiMapping"), jsii.Number(2))

	// 5. Verify SES Identity
	template.HasResourceProperties(jsii.String("AWS::SES::EmailIdentity"), map[string]interface{}{
		"EmailIdentity": "stackfoundry.co.uk",
	})

	// 6. Verify DNS Records (The "Phone Book" Checks)

	// A) Hosted Zone
	template.HasResourceProperties(jsii.String("AWS::Route53::HostedZone"), map[string]interface{}{
		"Name": "stackfoundry.co.uk.",
	})

	// B) MX Record (Google Workspace)
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "MX",
		"Name": "stackfoundry.co.uk.",
	})

	// C) A-Records (Traffic Routing)
	// Root Record (@)
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "A",
		"Name": "stackfoundry.co.uk.",
	})
	// WWW Record
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "A",
		"Name": "www.stackfoundry.co.uk.",
	})

	// D) TXT Records (SPF / DMARC)
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "TXT",
		// We can broadly check that we have TXT records configured
	})

	// 7. Verify Budget Alarm
	template.HasResourceProperties(jsii.String("AWS::Budgets::Budget"), map[string]interface{}{
		"Budget": map[string]interface{}{
			"BudgetLimit": map[string]interface{}{
				"Amount": 2,
				"Unit":   "USD",
			},
			"BudgetType": "COST",
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
}
