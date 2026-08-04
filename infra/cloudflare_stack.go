package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v5/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployCloudflare(ctx *pulumi.Context, zoneID string, aws *AWSResources) error {
	// 1. ACM VALIDATION RECORD
	// This validates your SSL cert — must be created before cert is usable
	_, err := cloudflare.NewRecord(ctx, "sf-acm-validation", &cloudflare.RecordArgs{
		ZoneId:  pulumi.String(zoneID),
		Type:    aws.CertValidationType.Elem(),
		Name:    aws.CertValidationName.Elem(),
		Value:   aws.CertValidationValue.Elem(),
		Ttl:     pulumi.Int(60),
		Proxied: pulumi.Bool(false), // Must be DNS only for ACM validation
	})
	if err != nil {
		return err
	}

	// 2. ROOT DOMAIN → API Gateway regional domain
	_, err = cloudflare.NewRecord(ctx, "sf-ai-root", &cloudflare.RecordArgs{
		ZoneId:  pulumi.String(zoneID),
		Type:    pulumi.String("CNAME"),
		Name:    pulumi.String("@"),
		Value:   aws.RegionalDomainName,
		Ttl:     pulumi.Int(1), // Auto when proxied
		Proxied: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	// 3. WWW → API Gateway regional domain
	_, err = cloudflare.NewRecord(ctx, "sf-ai-www", &cloudflare.RecordArgs{
		ZoneId:  pulumi.String(zoneID),
		Type:    pulumi.String("CNAME"),
		Name:    pulumi.String("www"),
		Value:   aws.RegionalDomainName,
		Ttl:     pulumi.Int(1),
		Proxied: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	// 4. SSL MODE — Full (not strict, API Gateway has its own AWS cert)
	_, err = cloudflare.NewZoneSettingsOverride(ctx, "sf-ai-zone-settings", &cloudflare.ZoneSettingsOverrideArgs{
		ZoneId: pulumi.String(zoneID),
		Settings: cloudflare.ZoneSettingsOverrideSettingsArgs{
			Ssl:            pulumi.StringPtr("full"),
			AlwaysUseHttps: pulumi.StringPtr("on"),
		},
	})
	if err != nil {
		return err
	}

	// 6. EMAIL RECORDS (Google Workspace MX)
	mxRecords := []struct {
		name     string
		value    string
		priority int
	}{
		{"sf-mx-1", "aspmx.l.google.com", 1},
		{"sf-mx-2", "alt1.aspmx.l.google.com", 5},
		{"sf-mx-3", "alt2.aspmx.l.google.com", 5},
		{"sf-mx-4", "alt3.aspmx.l.google.com", 10},
		{"sf-mx-5", "alt4.aspmx.l.google.com", 10},
	}

	for _, mx := range mxRecords {
		_, err = cloudflare.NewRecord(ctx, mx.name, &cloudflare.RecordArgs{
			ZoneId:   pulumi.String(zoneID),
			Type:     pulumi.String("MX"),
			Name:     pulumi.String("@"),
			Value:    pulumi.String(mx.value),
			Priority: pulumi.Int(mx.priority),
			Ttl:      pulumi.Int(3600),
		})
		if err != nil {
			return err
		}
	}

	// 7. SPF RECORD
	_, err = cloudflare.NewRecord(ctx, "sf-spf", &cloudflare.RecordArgs{
		ZoneId: pulumi.String(zoneID),
		Type:   pulumi.String("TXT"),
		Name:   pulumi.String("@"),
		Value:  pulumi.String("v=spf1 include:_spf.google.com include:amazonses.com ~all"),
		Ttl:    pulumi.Int(3600),
	})
	if err != nil {
		return err
	}

	// 8. DMARC RECORD
	_, err = cloudflare.NewRecord(ctx, "sf-dmarc", &cloudflare.RecordArgs{
		ZoneId: pulumi.String(zoneID),
		Type:   pulumi.String("TXT"),
		Name:   pulumi.String("_dmarc"),
		Value:  pulumi.String("v=DMARC1; p=reject; adkim=r; aspf=r; rua=mailto:dmarc_rua@onsecureserver.net;"),
		Ttl:    pulumi.Int(3600),
	})
	if err != nil {
		return err
	}

	return nil
}
