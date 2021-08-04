package provider

import (
	"context"

	mzcloud "github.com/MaterializeInc/cloud-sdks/go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{
				// TODO(benesch): switch to Frontegg client IDs/secrets, once
				// we've switched to Frontegg.
				"access_token": {
					Description: "The API access token token to authenticate with.",
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("MZCLOUD_ACCESS_TOKEN", nil),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"mzcloud_deployment": resourceMZCloudDeployment(),
			},
			ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
				return &apiClient{
					client:      mzcloud.NewAPIClient(mzcloud.NewConfiguration()),
					accessToken: d.Get("access_token").(string),
				}, nil
			},
		}
	}
}

type apiClient struct {
	client      *mzcloud.APIClient
	accessToken string
}
