package provider

import (
	"context"
	"fmt"
	"time"

	mzcloud "github.com/MaterializeInc/cloud-sdks/go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMZCloudDeployment() *schema.Resource {
	return &schema.Resource{
		Description: `Configures a Materialize Cloud deployment.`,

		CreateContext: resourceMZCloudDeploymentCreate,
		ReadContext:   resourceMZCloudDeploymentRead,
		UpdateContext: resourceMZCloudDeploymentUpdate,
		DeleteContext: resourceMZCloudDeploymentDelete,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "The internal cluster ID of the deployment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hostname": {
				Description: "The hostname at which the deployment is accessible.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"mz_version": {
				Description: "The Materialize version to deploy.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The human-readable name of the deployment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The size of the deployment.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceMZCloudDeploymentSerialize(d *schema.ResourceData) mzcloud.DeploymentRequest {
	size := d.Get("size").(string)
	mz_version := d.Get("mz_version").(string)
	return mzcloud.DeploymentRequest{
		Size:      &size,
		MzVersion: mz_version,
	}
}

func resourceMZCloudDeploymentDeserialize(d *schema.ResourceData, r mzcloud.Deployment) error {
	d.SetId(r.Id)
	if err := d.Set("hostname", r.Hostname); err != nil {
		return err
	}
	if err := d.Set("cluster_id", r.ClusterId); err != nil {
		return err
	}
	if err := d.Set("mz_version", r.MzVersion); err != nil {
		return err
	}
	if err := d.Set("name", r.Name); err != nil {
		return err
	}
	if err := d.Set("size", r.Size); err != nil {
		return err
	}
	return nil
}

func resourceMZCloudDeploymentWaitForUpdate(
	ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration,
) diag.Diagnostics {
	client := meta.(*apiClient)
	err := resource.Retry(timeout, func() *resource.RetryError {
		deployment, _, err := client.client.DeploymentsApi.DeploymentsRetrieve(ctx, d.Id()).Execute()
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("failed to retrieve deployment: %s", err))
		}
		if deployment.FlaggedForUpdate || deployment.StatefulsetStatus != "OK" {
			return resource.RetryableError(fmt.Errorf(
				"expected deployment to be ready but got flagged_for_update=%t status=%s",
				deployment.FlaggedForUpdate, deployment.StatefulsetStatus,
			))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceMZCloudDeploymentRead(ctx, d, meta)
}

func resourceMZCloudDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	ctx = context.WithValue(ctx, mzcloud.ContextAccessToken, client.accessToken)

	in := resourceMZCloudDeploymentSerialize(d)
	deployment, _, err := client.client.DeploymentsApi.DeploymentsCreate(ctx).DeploymentRequest(in).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(deployment.Id)
	return resourceMZCloudDeploymentWaitForUpdate(ctx, d, meta, d.Timeout(schema.TimeoutCreate))
}

func resourceMZCloudDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	ctx = context.WithValue(ctx, mzcloud.ContextAccessToken, client.accessToken)

	deployment, _, err := client.client.DeploymentsApi.DeploymentsRetrieve(ctx, d.Id()).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	if err := resourceMZCloudDeploymentDeserialize(d, deployment); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceMZCloudDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	ctx = context.WithValue(ctx, mzcloud.ContextAccessToken, client.accessToken)

	in := resourceMZCloudDeploymentSerialize(d)
	_, _, err := client.client.DeploymentsApi.DeploymentsUpdate(ctx, d.Id()).DeploymentRequest(in).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceMZCloudDeploymentWaitForUpdate(ctx, d, meta, d.Timeout(schema.TimeoutUpdate))
}

func resourceMZCloudDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	ctx = context.WithValue(ctx, mzcloud.ContextAccessToken, client.accessToken)

	_, err := client.client.DeploymentsApi.DeploymentsDestroy(ctx, d.Id()).Execute()
	if err != nil {
		return diag.FromErr(err)
	}
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		deployment, res, err := client.client.DeploymentsApi.DeploymentsRetrieve(ctx, d.Id()).Execute()
		if res.StatusCode == 404 {
			return nil
		} else if err != nil {
			return resource.NonRetryableError(fmt.Errorf("failed to retrieve deployment: %s", err))
		} else {
			return resource.RetryableError(fmt.Errorf(
				"expected deployment to be ready but got flagged_for_deletion=%t status=%s",
				deployment.FlaggedForDeletion, deployment.StatefulsetStatus,
			))
		}
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
