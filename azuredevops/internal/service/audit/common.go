package audit

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/audit"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/client"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
)

const errMsgUpdateAuditStream = "Error updating audit stream in Azure DevOps: %+v"

type flatFunc func(d *schema.ResourceData, auditStream *audit.AuditStream, daysToBackfill *int, enabled *bool)
type expandFunc func(d *schema.ResourceData) (*audit.AuditStream, *int, *bool)

// genBaseAuditStreamResource creates a Resource with the common parts
// that all Audit Streams require.
func genBaseAuditStreamResource(f flatFunc, e expandFunc) *schema.Resource {
	return &schema.Resource{
		Create: genAuditStreamCreateFunc(f, e),
		Read:   genAuditStreamReadFunc(f),
		Update: genAuditStreamUpdateFunc(f, e),
		Delete: genAuditStreamDeleteFunc(),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"days_to_backfill": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "The number of days of previously recorded audit data that will be replayed into the stream",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// doBaseExpansion performs the expansion for the 'base' attributes that are defined in the schema, above
func doBaseExpansion(d *schema.ResourceData) (*audit.AuditStream, *int, *bool) {
	// an "error" is OK here as it is expected in the case that the ID is not set in the resource data
	var auditStreamId *int
	parsedId, err := strconv.Atoi(d.Id())
	if err == nil {
		auditStreamId = &parsedId
	}

	daysToBackfill := converter.Int(d.Get("days_to_backfill").(int))
	enabled := converter.Bool(d.Get("enabled").(bool))
	auditStream := &audit.AuditStream{
		Id: auditStreamId,
	}

	return auditStream, daysToBackfill, enabled
}

// doBaseFlattening performs the flattening for the 'base' attributes that are defined in the schema, above
func doBaseFlattening(d *schema.ResourceData, auditStream *audit.AuditStream, daysToBackfill *int, enabled *bool) {
	d.SetId(strconv.Itoa(*auditStream.Id))
	d.Set("days_to_backfill", daysToBackfill)
	d.Set("enabled", enabled)
	d.Set("name", auditStream.DisplayName)
}

func genAuditStreamCreateFunc(flatFunc flatFunc, expandFunc expandFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*client.AggregatedClient)
		auditStream, daysToBackfill, enabled := expandFunc(d)

		createdAuditStream, err := createAuditStream(clients, auditStream, daysToBackfill, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return fmt.Errorf("Error creating audit stream in Azure DevOps: %+v", err)
		}

		statefulStream, err := setStreamStatusState(clients, createdAuditStream, *enabled)
		if err != nil {
			return fmt.Errorf(errMsgUpdateAuditStream, err)
		}

		d.SetId(strconv.Itoa(*statefulStream.Id))
		return genAuditStreamReadFunc(flatFunc)(d, m)
	}
}

func genAuditStreamReadFunc(flatFunc flatFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*client.AggregatedClient)
		streamId, err := strconv.Atoi(d.Id())
		if err != nil {
			return fmt.Errorf("Error parsing the audit stream ID from the Terraform resource data: %v", err)
		}

		daysToBackfill := d.Get("days_to_backfill").(int)
		enabled := converter.Bool(d.Get("enabled").(bool))

		auditStream, err := readAuditStream(clients, streamId)
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error looking up audit stream with ID %d. Error: %v", streamId, err)
		}

		if auditStream.Id == nil {
			// e.g. audit stream has been deleted separately without TF
			d.SetId("")
		} else {
			flatFunc(d, auditStream, &daysToBackfill, enabled)
		}
		return nil
	}
}

func genAuditStreamUpdateFunc(flatFunc flatFunc, expandFunc expandFunc) schema.UpdateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*client.AggregatedClient)
		auditStream, daysToBackfill, enabled := expandFunc(d)

		updatedAuditStream, err := updateAuditStream(clients, auditStream)
		if err != nil {
			return fmt.Errorf(errMsgUpdateAuditStream, err)
		}

		statefulStream, err := setStreamStatusState(clients, updatedAuditStream, *enabled)
		if err != nil {
			return fmt.Errorf(errMsgUpdateAuditStream, err)
		}

		flatFunc(d, statefulStream, daysToBackfill, enabled)
		return genAuditStreamReadFunc(flatFunc)(d, m)
	}
}

func genAuditStreamDeleteFunc() schema.DeleteFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*client.AggregatedClient)
		streamId, err := strconv.Atoi(d.Id())
		if err != nil {
			return err
		}

		err = deleteAuditStream(clients, streamId)
		if err != nil {
			return fmt.Errorf("Error deleting audit stream in Azure DevOps: %+v", err)
		}

		return nil
	}
}

func createAuditStream(clients *client.AggregatedClient, stream *audit.AuditStream, daysToBackfill *int, timeoutSeconds time.Duration) (*audit.AuditStream, error) {
	createdAuditStream, err := clients.AuditClient.CreateStream(
		clients.Ctx,
		audit.CreateStreamArgs{
			Stream:         stream,
			DaysToBackfill: daysToBackfill,
		})

	if err != nil {
		return nil, err
	}

	stateConf := &resource.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     5 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending: []string{
			string(audit.AuditStreamStatusValues.Backfilling),
		},
		Target: []string{
			string(audit.AuditStreamStatusValues.Enabled),
			string(audit.AuditStreamStatusValues.DisabledByUser),
			string(audit.AuditStreamStatusValues.DisabledBySystem),
		},
		Refresh: readStreamStatus(clients, *createdAuditStream.Id),
		Timeout: timeoutSeconds,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		// if delErr := deleteServiceEndpoint(clients, projectID, createdServiceEndpoint.Id, d.Timeout(schema.TimeoutDelete)); delErr != nil {
		// 	log.Printf("[DEBUG] Failed to delete the failed service endpoint: %v ", delErr)
		// }
		return nil, fmt.Errorf(" waiting for auditstream ready. %v ", err)
	}

	return createdAuditStream, err
}

func readAuditStream(clients *client.AggregatedClient, streamId int) (*audit.AuditStream, error) {
	auditStream, err := clients.AuditClient.QueryStreamById(
		clients.Ctx,
		audit.QueryStreamByIdArgs{
			StreamId: &streamId,
		})

	return auditStream, err
}

func updateAuditStream(clients *client.AggregatedClient, stream *audit.AuditStream) (*audit.AuditStream, error) {
	updatedAuditStream, err := clients.AuditClient.UpdateStream(
		clients.Ctx,
		audit.UpdateStreamArgs{
			Stream: stream,
		})

	return updatedAuditStream, err
}

func deleteAuditStream(clients *client.AggregatedClient, streamId int) error {
	return clients.AuditClient.DeleteStream(
		clients.Ctx,
		audit.DeleteStreamArgs{
			StreamId: &streamId,
		})
}

func setStreamStatusState(clients *client.AggregatedClient, stream *audit.AuditStream, enabled bool) (*audit.AuditStream, error) {
	var streamStatus *audit.AuditStreamStatus
	streamStatus = &audit.AuditStreamStatusValues.Enabled
	if !enabled {
		streamStatus = &audit.AuditStreamStatusValues.DisabledByUser
	}

	if stream.Status != streamStatus {
		updatedAuditStream, err := clients.AuditClient.UpdateStatus(
			clients.Ctx,
			audit.UpdateStatusArgs{
				StreamId: stream.Id,
				Status:   streamStatus,
			})

		return updatedAuditStream, err
	} else {
		return stream, nil
	}
}

func readStreamStatus(clients *client.AggregatedClient, streamId int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		auditStream, err := readAuditStream(clients, streamId)

		if err != nil {
			return nil, string(audit.AuditStreamStatusValues.Unknown), err
		}

		if *auditStream.Status != audit.AuditStreamStatusValues.Enabled {
			log.Printf("[DEBUG] Waiting for audit stream creation. Operation result %v", auditStream.Status)
		}

		return auditStream, string(*auditStream.Status), nil
	}
}
