package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// Cloudflare Zone ID for stackfoundry.ai
		// Get this from Cloudflare dashboard → stackfoundry.ai → Overview → Zone ID (bottom right)
		cloudflareZoneID := cfg.Require("cloudflareZoneID")
		enableCustomDomain := cfg.GetBool("enableCustomDomain")

		// Deploy AWS resources
		aws, err := deployAWS(ctx, enableCustomDomain)
		if err != nil {
			return err
		}

		// Deploy Cloudflare resources
		if err := deployCloudflare(ctx, cloudflareZoneID, aws); err != nil {
			return err
		}

		return nil
	})
}
