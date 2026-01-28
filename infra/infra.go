package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsbudgets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StackFoundryWebsiteStackProps struct {
	awscdk.StackProps
}

func NewStackFoundryWebsiteStack(scope constructs.Construct, id string, props *StackFoundryWebsiteStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// 1. DOMAIN & DNS
	domainNameStr := "stackfoundry.co.uk"
	zone := awsroute53.NewHostedZone(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneProps{
		ZoneName: jsii.String(domainNameStr),
	})

	// 2. SSL CERTIFICATE
	cert := awscertificatemanager.NewCertificate(stack, jsii.String("SiteCert"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(domainNameStr),
		Validation: awscertificatemanager.CertificateValidation_FromDns(zone),
	})

	// 3. API GATEWAY CUSTOM DOMAIN
	dn := awsapigatewayv2.NewDomainName(stack, jsii.String("DN"), &awsapigatewayv2.DomainNameProps{
		DomainName:  jsii.String(domainNameStr),
		Certificate: cert,
	})

	// 4. LAMBDA FUNCTION
	fn := awslambda.NewFunction(stack, jsii.String("StackFoundryApp"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(jsii.String("../"), nil),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(10)),
		Environment: &map[string]*string{
			"GIN_MODE": jsii.String("release"),
		},
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
	})

	// 5. PERMISSIONS (SES)
	fn.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("ses:SendEmail", "ses:SendRawEmail"),
		Resources: jsii.Strings("*"),
	}))

	// 6. API GATEWAY (HTTP API)
	api := awsapigatewayv2.NewHttpApi(stack, jsii.String("StackFoundryAPI"), &awsapigatewayv2.HttpApiProps{
		DefaultIntegration: awsapigatewayv2integrations.NewHttpLambdaIntegration(
			jsii.String("LambdaIntegration"),
			fn,
			nil,
		),
		DefaultDomainMapping: &awsapigatewayv2.DomainMappingOptions{
			DomainName: dn,
		},
	})

	// 7. RATE LIMITING
	if api.DefaultStage() != nil && api.DefaultStage().Node() != nil && api.DefaultStage().Node().DefaultChild() != nil {
		cfnStage := api.DefaultStage().Node().DefaultChild().(awsapigatewayv2.CfnStage)
		cfnStage.SetDefaultRouteSettings(&awsapigatewayv2.CfnStage_RouteSettingsProperty{
			ThrottlingBurstLimit: jsii.Number(50),
			ThrottlingRateLimit:  jsii.Number(10),
		})
	}

	// 8. BILLING ALARM
	awsbudgets.NewCfnBudget(stack, jsii.String("LowCostBudget"), &awsbudgets.CfnBudgetProps{
		Budget: &awsbudgets.CfnBudget_BudgetDataProperty{
			BudgetType: jsii.String("COST"),
			TimeUnit:   jsii.String("MONTHLY"),
			BudgetLimit: &awsbudgets.CfnBudget_SpendProperty{
				Amount: jsii.Number(2), // $2.00 Limit
				Unit:   jsii.String("USD"),
			},
		},
		NotificationsWithSubscribers: []interface{}{
			&awsbudgets.CfnBudget_NotificationWithSubscribersProperty{
				Notification: &awsbudgets.CfnBudget_NotificationProperty{
					NotificationType:   jsii.String("ACTUAL"),
					ComparisonOperator: jsii.String("GREATER_THAN"),
					Threshold:          jsii.Number(80), // 80% of $2.00
				},
				Subscribers: []interface{}{
					&awsbudgets.CfnBudget_SubscriberProperty{
						SubscriptionType: jsii.String("EMAIL"),
						Address:          jsii.String("joe@stackfoundry.co.uk"),
					},
				},
			},
		},
	})

	// 9. DNS A-RECORD
	awsroute53.NewARecord(stack, jsii.String("AliasRecord"), &awsroute53.ARecordProps{
		Zone: zone,
		Target: awsroute53.RecordTarget_FromAlias(
			awsroute53targets.NewApiGatewayv2DomainProperties(
				dn.RegionalDomainName(),
				dn.RegionalHostedZoneId(),
			),
		),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApiEndpoint"), &awscdk.CfnOutputProps{
		Value: api.ApiEndpoint(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("Nameservers"), &awscdk.CfnOutputProps{
		Value: jsii.String("Check Route53 Console for NS records!"),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	// One Account dev and prod website for future changes
	val := app.Node().TryGetContext(jsii.String("stage"))
	var stageStr string
	if val == nil {
		stageStr = "prod"
	} else {
		if str, ok := val.(string); ok {
			stageStr = str
		} else {
			stageStr = "prod"
		}
	}

	stackName := "StackFoundry-" + stageStr

	NewStackFoundryWebsiteStack(app, stackName, &StackFoundryWebsiteStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("eu-west-2"),
			},
			Tags: &map[string]*string{
				"Environment": jsii.String(stageStr),
			},
		},
	})

	app.Synth(nil)
}
