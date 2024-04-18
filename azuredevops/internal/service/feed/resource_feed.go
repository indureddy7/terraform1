package feed

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/client"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils"
)

func ResourceFeed() *schema.Resource {
	return &schema.Resource{
		Create: resourceFeedCreate,
		Read:   resourceFeedRead,
		Update: resourceFeedUpdate,
		Delete: resourceFeedDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
				ForceNew:     true,
			},
			"project_id": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ForceNew:     true,
			},
			"permanent_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"restored": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceFeedCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)

	if isFeedRestorable(d, m) {
		err := restoreFeed(d, m)

		if err != nil {
			return fmt.Errorf("restoring feed. Name: %s, Error: %+v", name, err)
		}

		return resourceFeedRead(d, m)
	}

	err := createFeed(d, m)

	if err != nil {
		return fmt.Errorf("creating new feed. Name: %s, Error: %+v", name, err)
	}

	return resourceFeedRead(d, m)
}

func resourceFeedRead(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)

	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)

	getFeed, err := clients.FeedClient.GetFeed(clients.Ctx, feed.GetFeedArgs{
		FeedId:  &name,
		Project: &projectId,
	})

	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading feed during read: %+v", err)
	}

	if getFeed != nil {
		d.SetId(getFeed.Id.String())
		d.Set("name", getFeed.Name)
		if getFeed.Project != nil {
			d.Set("project_id", getFeed.Project.Id.String())
		}
	}

	return nil
}

func resourceFeedUpdate(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)

	_, err := clients.FeedClient.UpdateFeed(clients.Ctx, feed.UpdateFeedArgs{
		Feed:    &feed.FeedUpdate{},
		FeedId:  &name,
		Project: &projectId,
	})

	if err != nil {
		return err
	}

	return resourceFeedRead(d, m)
}

func resourceFeedDelete(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)
	permanentDelete := d.Get("permanent_delete").(bool)

	err := clients.FeedClient.DeleteFeed(clients.Ctx, feed.DeleteFeedArgs{
		FeedId:  &name,
		Project: &projectId,
	})

	if err != nil {
		return err
	}

	if permanentDelete {
		err = clients.FeedClient.PermanentDeleteFeed(clients.Ctx, feed.PermanentDeleteFeedArgs{
			FeedId:  &name,
			Project: &projectId,
		})

		if err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

func isFeedRestorable(d *schema.ResourceData, m interface{}) bool {
	clients := m.(*client.AggregatedClient)
	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)

	change, err := clients.FeedClient.GetFeedChange(clients.Ctx, feed.GetFeedChangeArgs{
		FeedId:  &name,
		Project: &projectId,
	})

	return err == nil && *((*change).ChangeType) == feed.ChangeTypeValues.Delete
}

func createFeed(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)

	createFeed := feed.Feed{
		Name: &name,
	}

	_, err := clients.FeedClient.CreateFeed(clients.Ctx, feed.CreateFeedArgs{
		Feed:    &createFeed,
		Project: &projectId,
	})

	if err != nil {
		return err
	}

	d.Set("restored", false)

	return nil
}

func restoreFeed(d *schema.ResourceData, m interface{}) error {
	clients := m.(*client.AggregatedClient)
	name := d.Get("name").(string)
	projectId := d.Get("project_id").(string)

	path := "/isDeleted"

	patchJsons := []webapi.JsonPatchOperation{{
		From:  nil,
		Path:  &path,
		Op:    &webapi.OperationValues.Replace,
		Value: false,
	}}

	err := clients.FeedClient.RestoreDeletedFeed(clients.Ctx, feed.RestoreDeletedFeedArgs{
		FeedId:    &name,
		Project:   &projectId,
		PatchJson: &patchJsons,
	})

	if err != nil {
		return err
	}

	d.Set("restored", true)

	return nil
}
