package main

import (
	"fmt"

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
	"github.com/aws/aws-cdk-go/awscdk/v2/awsses"
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
	wwwDomainNameStr := "www." + domainNameStr

	zone := awsroute53.NewHostedZone(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneProps{
		ZoneName: jsii.String(domainNameStr),
	})

	// 1a. SES Domain Identity
	sesIdentity := awsses.NewEmailIdentity(stack, jsii.String("SesIdentity"), &awsses.EmailIdentityProps{
		Identity: awsses.Identity_Domain(jsii.String(domainNameStr)),
	})

	if sesIdentity.DkimRecords() != nil {
		for i, record := range *sesIdentity.DkimRecords() {
			awsroute53.NewCnameRecord(stack, jsii.String(fmt.Sprintf("DKIM%d", i)), &awsroute53.CnameRecordProps{
				Zone:       zone,
				RecordName: record.Name,
				DomainName: record.Value,
				Ttl:        awscdk.Duration_Minutes(jsii.Number(60)),
			})
		}
	}

	// 1b. MX Records (Google Workspace)
	awsroute53.NewMxRecord(stack, jsii.String("GoogleMX"), &awsroute53.MxRecordProps{
		Zone: zone,
		Values: &[]*awsroute53.MxRecordValue{
			{Priority: jsii.Number(1), HostName: jsii.String("aspmx.l.google.com.")},
			{Priority: jsii.Number(5), HostName: jsii.String("alt1.aspmx.l.google.com.")},
			{Priority: jsii.Number(5), HostName: jsii.String("alt2.aspmx.l.google.com.")},
			{Priority: jsii.Number(10), HostName: jsii.String("alt3.aspmx.l.google.com.")},
			{Priority: jsii.Number(10), HostName: jsii.String("alt4.aspmx.l.google.com.")},
		},
		Ttl: awscdk.Duration_Minutes(jsii.Number(60)),
	})

	// 1c. TXT Record (SPF & Google Verification)
	awsroute53.NewTxtRecord(stack, jsii.String("RootTXT"), &awsroute53.TxtRecordProps{
		Zone: zone,
		Values: jsii.Strings(
			"google-site-verification=_Df0skkdrZbUu8J_u4fePz-mx6RqGF_TdrPkOp0z3A0",
			"v=spf1 include:_spf.google.com include:amazonses.com ~all",
		),
		Ttl: awscdk.Duration_Minutes(jsii.Number(60)),
	})

	// 1d. DMARC Record
	awsroute53.NewTxtRecord(stack, jsii.String("DmarcTXT"), &awsroute53.TxtRecordProps{
		Zone:       zone,
		RecordName: jsii.String("_dmarc"),
		Values: jsii.Strings(
			"v=DMARC1; p=reject; adkim=r; aspf=r; rua=mailto:dmarc_rua@onsecureserver.net;",
		),
		Ttl: awscdk.Duration_Minutes(jsii.Number(60)),
	})

	// 2. SSL CERTIFICATE
	cert := awscertificatemanager.NewCertificate(stack, jsii.String("SiteCert"), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(domainNameStr),
		SubjectAlternativeNames: jsii.Strings(wwwDomainNameStr),
		Validation:              awscertificatemanager.CertificateValidation_FromDns(zone),
	})

	// 3. API GATEWAY CUSTOM DOMAINS
	// Root Domain
	dn := awsapigatewayv2.NewDomainName(stack, jsii.String("DN"), &awsapigatewayv2.DomainNameProps{
		DomainName:  jsii.String(domainNameStr),
		Certificate: cert,
	})

	// WWW Domain
	dnWww := awsapigatewayv2.NewDomainName(stack, jsii.String("DNWww"), &awsapigatewayv2.DomainNameProps{
		DomainName:  jsii.String(wwwDomainNameStr),
		Certificate: cert,
	})

	// 4. LAMBDA FUNCTION
	logGroup := awslogs.NewLogGroup(stack, jsii.String("AppLogs"), &awslogs.LogGroupProps{
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	fn := awslambda.NewFunction(stack, jsii.String("StackFoundryWebsiteRunner"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(jsii.String("../dist"), nil),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(5)),
		Environment: &map[string]*string{
			"GIN_MODE": jsii.String("release"),
		},
		LogGroup: logGroup,
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
			&awsapigatewayv2integrations.HttpLambdaIntegrationProps{
				// Keep Payload 1.0 for HTML rendering compatibility
				PayloadFormatVersion: awsapigatewayv2.PayloadFormatVersion_VERSION_1_0(),
			},
		),
		DefaultDomainMapping: &awsapigatewayv2.DomainMappingOptions{
			DomainName: dn,
		},
	})

	// Explicitly map the WWW domain to the API as well
	awsapigatewayv2.NewApiMapping(stack, jsii.String("WwwMapping"), &awsapigatewayv2.ApiMappingProps{
		Api:        api,
		DomainName: dnWww,
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
					Threshold:          jsii.Number(80),
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

	// 9. DNS RECORDS
	// Root A-Record
	awsroute53.NewARecord(stack, jsii.String("AliasRecord"), &awsroute53.ARecordProps{
		Zone: zone,
		Target: awsroute53.RecordTarget_FromAlias(
			awsroute53targets.NewApiGatewayv2DomainProperties(
				dn.RegionalDomainName(),
				dn.RegionalHostedZoneId(),
			),
		),
	})

	// WWW A-Record
	awsroute53.NewARecord(stack, jsii.String("AliasRecordWww"), &awsroute53.ARecordProps{
		Zone:       zone,
		RecordName: jsii.String("www"),
		Target: awsroute53.RecordTarget_FromAlias(
			awsroute53targets.NewApiGatewayv2DomainProperties(
				dnWww.RegionalDomainName(),
				dnWww.RegionalHostedZoneId(),
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

	stackName := "StackFoundryWebsite-" + stageStr

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
