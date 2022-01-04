package testutils

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/audit"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/client"
)

func CheckAuditStreamExists(tfNode string, expectedType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("Did not find an audit stream in the state")
		}

		auditStream, err := getAuditStreamFromState(resourceState)
		if err != nil {
			return err
		}

		if *auditStream.ConsumerType != expectedType {
			return fmt.Errorf("Audit Stream has Type=%s, but expected Type=%s", *auditStream.ConsumerType, expectedType)
		}

		return nil
	}
}

func CheckAuditStreamDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, resource := range s.RootModule().Resources {
			if resource.Type != resourceType {
				continue
			}

			// indicates the resource exists - this should fail the test
			if _, err := getSvcEndpointFromState(resource); err == nil {
				return fmt.Errorf("Unexpectedly found a service endpoint that should have been deleted")
			}
		}

		return nil
	}
}

// given a resource from the state, return a service endpoint (and error)
func getAuditStreamFromState(resource *terraform.ResourceState) (*audit.AuditStream, error) {
	auditStreamDefId, err := strconv.Atoi(resource.Primary.ID)
	if err != nil {
		return nil, err
	}

	clients := GetProvider().Meta().(*client.AggregatedClient)
	return clients.AuditClient.QueryStreamById(clients.Ctx, audit.QueryStreamByIdArgs{
		StreamId: &auditStreamDefId,
	})
}
