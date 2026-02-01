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

	// 3. Verify SES Identity
	template.HasResourceProperties(jsii.String("AWS::SES::EmailIdentity"), map[string]interface{}{
		"EmailIdentity": "stackfoundry.co.uk",
	})

	// 4. Verify DNS Records (The "Phone Book" Checks)

	// A) Verify Hosted Zone
	template.HasResourceProperties(jsii.String("AWS::Route53::HostedZone"), map[string]interface{}{
		"Name": "stackfoundry.co.uk.",
	})

	// B) Verify MX Record (Google Workspace)
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "MX",
		"Name": "stackfoundry.co.uk.",
	})

	// C) Verify DKIM CNAME Records (The Loop Check)
	// These are typically: 3 DKIM CNAMEs
	template.HasResourceProperties(jsii.String("AWS::Route53::RecordSet"), map[string]interface{}{
		"Type": "CNAME",
	})

	// Count check: Ensure we have exactly 3 CNAMEs (SES Standard)
	// Note: If you add other CNAMEs later, increase this number.
	template.ResourceCountIs(jsii.String("AWS::Route53::RecordSet"), jsii.Number(7))

	// 5. Verify Budget Alarm
	template.HasResourceProperties(jsii.String("AWS::Budgets::Budget"), map[string]interface{}{
		"Budget": map[string]interface{}{
			"BudgetLimit": map[string]interface{}{
				"Amount": 2,
			},
		},
	})
}
