package audit

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/audit"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/tfhelper"
)

// ResourceAuditStreamAzureEventGrid schema and implementation for Azure EventHub audit resource
func ResourceAuditStreamSplunk() *schema.Resource {
	r := genBaseAuditStreamResource(flattenAuditStreamSplunk, expandAuditStreamSplunk)

	r.Schema["url"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		DefaultFunc:  schema.EnvDefaultFunc("AZDO_AUDIT_SPLUNK_URL", nil),
		ValidateFunc: validation.IsURLWithHTTPS,
		Description:  "Url for the Splunk instance that will send events to. It should follow format https://<hostname>:<port>",
	}

	r.Schema["collector_token"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Sensitive:        true,
		DefaultFunc:      schema.EnvDefaultFunc("AZDO_AUDIT_SPLUNK_COLLECTOR_TOKEN", nil),
		DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
		ValidateFunc:     validation.StringIsNotWhiteSpace,
		Description:      "The event collector token generated by the Splunk instance",
	}
	// Add a spot in the schema to store the token secretly
	stSecretHashKey, stSecretHashSchema := tfhelper.GenerateSecreteMemoSchema("collector_token")
	r.Schema[stSecretHashKey] = stSecretHashSchema

	return r
}

// Convert internal Terraform data structure to an AzDO data structure
func expandAuditStreamSplunk(d *schema.ResourceData) (*audit.AuditStream, *int, error) {
	auditStream, daysToBackfill := doBaseExpansion(d)
	auditStream.ConsumerType = converter.String("Splunk")
	auditStream.ConsumerInputs = &map[string]string{
		"SplunkUrl":                 d.Get("url").(string),
		"SplunkEventCollectorToken": d.Get("collector_token").(string),
	}

	return auditStream, daysToBackfill, nil
}

// Convert AzDO data structure to internal Terraform data structure
func flattenAuditStreamSplunk(d *schema.ResourceData, auditStream *audit.AuditStream, daysToBackfill *int) {
	doBaseFlattening(d, auditStream, daysToBackfill)

	tfhelper.HelpFlattenSecret(d, "collector_token")

	d.Set("url", (*auditStream.ConsumerInputs)["SplunkUrl"])
	d.Set("collector_token", (*auditStream.ConsumerInputs)["SplunkEventCollectorToken"])
}
