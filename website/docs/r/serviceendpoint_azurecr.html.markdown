---
layout: "azuredevops"
page_title: "AzureDevops: azuredevops_serviceendpoint_azurecr"
description: |-
  Manages a Azure Container Registry service endpoint within Azure DevOps organization.
---

# azuredevops_serviceendpoint_azurecr

Manages a Azure Container Registry service endpoint within Azure DevOps.

## Example Usage

```hcl
resource "azuredevops_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

# azure container registry service connection
resource "azuredevops_serviceendpoint_azurecr" "example" {
  project_id                = azuredevops_project.example.id
  service_endpoint_name     = "Example AzureCR"
  resource_group            = "example-rg"
  azurecr_spn_tenantid      = "00000000-0000-0000-0000-000000000000"
  azurecr_name              = "ExampleAcr"
  azurecr_subscription_id   = "00000000-0000-0000-0000-000000000000"
  azurecr_subscription_name = "subscription name"
}
```

## Argument Reference

The following arguments are supported:

- `project_id` - (Required) The ID of the project.
- `service_endpoint_name` - (Required) The name you will use to refer to this service connection in task inputs.
- `resource_group` - (Required) The resource group to which the container registry belongs.
- `azurecr_spn_tenantid` - (Required) The tenant id of the service principal.
- `azurecr_name` - (Required) The Azure container registry name.
- `azurecr_subscription_id` - (Required) The subscription id of the Azure targets.
- `azurecr_subscription_name` - (Required) The subscription name of the Azure targets.
- `description` - (Optional) The Service Endpoint description. Defaults to `Managed by Terraform`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the service endpoint.
- `project_id` - The ID of the project.
- `service_endpoint_name` - The Service Endpoint name.
- `service_principal_id` - The service principal ID.

## Relevant Links

- [Azure DevOps Service REST API 6.0 - Service Endpoints](https://docs.microsoft.com/en-us/rest/api/azure/devops/serviceendpoint/endpoints?view=azure-devops-rest-6.0)
- [Azure Container Registry REST API](https://docs.microsoft.com/en-us/rest/api/containerregistry/)

## Import

Azure DevOps Service Endpoint Azure Container Registry can be imported using **projectID/serviceEndpointID** or **projectName/serviceEndpointID**

```sh
terraform import azuredevops_serviceendpoint_azurecr.example 00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000
```
