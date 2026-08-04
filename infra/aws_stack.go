package main

import (
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/budgets"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AWSResources struct {
	APIEndpoint         pulumi.StringOutput
	RegionalDomainName  pulumi.StringOutput
	CertValidationName  pulumi.StringPtrOutput
	CertValidationValue pulumi.StringPtrOutput
	CertValidationType  pulumi.StringPtrOutput
}

func deployAWS(ctx *pulumi.Context, enableCustomDomain bool) (*AWSResources, error) {
	// 1. ACM CERTIFICATE for stackfoundry.ai + wildcard
	cert, err := acm.NewCertificate(ctx, "sf-ai-cert", &acm.CertificateArgs{
		DomainName: pulumi.String("stackfoundry.ai"),
		SubjectAlternativeNames: pulumi.StringArray{
			pulumi.String("*.stackfoundry.ai"),
		},
		ValidationMethod: pulumi.String("DNS"),
	})
	if err != nil {
		return nil, err
	}

	// Export the validation record so we can add it to Cloudflare
	certValidation := cert.DomainValidationOptions.Index(pulumi.Int(0))
	certValidationName := certValidation.ResourceRecordName()
	certValidationValue := certValidation.ResourceRecordValue()
	certValidationType := certValidation.ResourceRecordType()
	ctx.Export("certValidationName", certValidationName)
	ctx.Export("certValidationValue", certValidationValue)
	ctx.Export("certValidationType", certValidationType)

	// 2. IAM ROLE for Lambda
	assumeRole, err := iam.NewRole(ctx, "sf-lambda-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"Service": "lambda.amazonaws.com"},
				"Action": "sts:AssumeRole"
			}]
		}`),
	})
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "sf-lambda-basic", &iam.RolePolicyAttachmentArgs{
		Role:      assumeRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicy(ctx, "sf-lambda-ses", &iam.RolePolicyArgs{
		Role: assumeRole.Name,
		Policy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Action": ["ses:SendEmail", "ses:SendRawEmail"],
				"Resource": "*"
			}]
		}`),
	})
	if err != nil {
		return nil, err
	}

	// 3. CLOUDWATCH LOG GROUP
	logGroup, err := cloudwatch.NewLogGroup(ctx, "sf-ai-logs", &cloudwatch.LogGroupArgs{
		RetentionInDays: pulumi.Int(7),
	})
	if err != nil {
		return nil, err
	}

	// 4. LAMBDA FUNCTION
	fn, err := lambda.NewFunction(ctx, "sf-ai-runner", &lambda.FunctionArgs{
		Runtime:       pulumi.String("provided.al2023"),
		Architectures: pulumi.StringArray{pulumi.String("arm64")},
		Handler:       pulumi.String("bootstrap"),
		Role:          assumeRole.Arn,
		Code:          pulumi.NewFileArchive("../dist"),
		MemorySize:    pulumi.Int(128),
		Timeout:       pulumi.Int(5),
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"GIN_MODE": pulumi.String("release"),
			},
		},
		LoggingConfig: &lambda.FunctionLoggingConfigArgs{
			LogGroup: logGroup.Name,
		},
	})
	if err != nil {
		return nil, err
	}

	// 5. API GATEWAY
	api, err := apigatewayv2.NewApi(ctx, "sf-ai-api", &apigatewayv2.ApiArgs{
		ProtocolType: pulumi.String("HTTP"),
	})
	if err != nil {
		return nil, err
	}

	integration, err := apigatewayv2.NewIntegration(ctx, "sf-ai-integration", &apigatewayv2.IntegrationArgs{
		ApiId:                api.ID(),
		IntegrationType:      pulumi.String("AWS_PROXY"),
		IntegrationUri:       fn.Arn,
		PayloadFormatVersion: pulumi.String("1.0"),
	})
	if err != nil {
		return nil, err
	}

	_, err = apigatewayv2.NewRoute(ctx, "sf-ai-route", &apigatewayv2.RouteArgs{
		ApiId:    api.ID(),
		RouteKey: pulumi.String("$default"),
		Target:   pulumi.Sprintf("integrations/%s", integration.ID()),
	})
	if err != nil {
		return nil, err
	}

	stage, err := apigatewayv2.NewStage(ctx, "sf-ai-stage", &apigatewayv2.StageArgs{
		ApiId:      api.ID(),
		Name:       pulumi.String("$default"),
		AutoDeploy: pulumi.Bool(true),
		DefaultRouteSettings: &apigatewayv2.StageDefaultRouteSettingsArgs{
			ThrottlingBurstLimit: pulumi.Int(50),
			ThrottlingRateLimit:  pulumi.Float64(10),
		},
	})
	if err != nil {
		return nil, err
	}

	// Lambda permission for API Gateway
	_, err = lambda.NewPermission(ctx, "sf-ai-apigw-permission", &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  fn.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.Sprintf("%s/*/*", stage.ExecutionArn),
	})
	if err != nil {
		return nil, err
	}

	regionalDomainName := api.ApiEndpoint.ApplyT(func(endpoint string) string {
		// API endpoint includes scheme; Cloudflare CNAME needs only hostname.
		clean := strings.TrimPrefix(endpoint, "https://")
		return strings.TrimPrefix(clean, "http://")
	}).(pulumi.StringOutput)

	// 6. API GATEWAY CUSTOM DOMAIN for stackfoundry.ai
	// Guard this behind config so first deploy can focus on cert issuance.
	if enableCustomDomain {
		domain, err := apigatewayv2.NewDomainName(ctx, "sf-ai-domain", &apigatewayv2.DomainNameArgs{
			DomainName: pulumi.String("stackfoundry.ai"),
			DomainNameConfiguration: &apigatewayv2.DomainNameDomainNameConfigurationArgs{
				CertificateArn: cert.Arn,
				EndpointType:   pulumi.String("REGIONAL"),
				SecurityPolicy: pulumi.String("TLS_1_2"),
			},
		})
		if err != nil {
			return nil, err
		}

		_, err = apigatewayv2.NewApiMapping(ctx, "sf-ai-mapping", &apigatewayv2.ApiMappingArgs{
			ApiId:      api.ID(),
			DomainName: domain.ID(),
			Stage:      stage.ID(),
		})
		if err != nil {
			return nil, err
		}

		regionalDomainName = domain.DomainNameConfiguration.ApplyT(func(c apigatewayv2.DomainNameDomainNameConfiguration) string {
			return *c.TargetDomainName
		}).(pulumi.StringOutput)
	}

	// 7. BUDGET
	_, err = budgets.NewBudget(ctx, "sf-ai-budget", &budgets.BudgetArgs{
		BudgetType:  pulumi.String("COST"),
		TimeUnit:    pulumi.String("MONTHLY"),
		LimitAmount: pulumi.String("2"),
		LimitUnit:   pulumi.String("USD"),
		Notifications: budgets.BudgetNotificationArray{
			&budgets.BudgetNotificationArgs{
				ComparisonOperator:       pulumi.String("GREATER_THAN"),
				NotificationType:         pulumi.String("ACTUAL"),
				Threshold:                pulumi.Float64(80),
				ThresholdType:            pulumi.String("PERCENTAGE"),
				SubscriberEmailAddresses: pulumi.StringArray{pulumi.String("joe@stackfoundry.co.uk")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ctx.Export("apiEndpoint", api.ApiEndpoint)
	ctx.Export("regionalDomainName", regionalDomainName)

	return &AWSResources{
		APIEndpoint: api.ApiEndpoint,
		RegionalDomainName:  regionalDomainName,
		CertValidationName:  certValidationName,
		CertValidationValue: certValidationValue,
		CertValidationType:  certValidationType,
	}, nil
}
