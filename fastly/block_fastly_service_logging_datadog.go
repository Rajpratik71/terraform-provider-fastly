package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DatadogServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DatadogServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingDatadog returns a new resource.
func NewServiceLoggingDatadog(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&DatadogServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_datadog",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *DatadogServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *DatadogServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Datadog logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "US",
			Description: "The region that log data will be sent to. One of `US` or `EU`. Defaults to `US` if undefined",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The API key from your Datadog account",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *DatadogServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Datadog logging addition opts: %#v", opts)

	return createDatadog(conn, opts)
}

// Read refreshes the resource.
func (h *DatadogServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Datadog logging endpoints for (%s)", d.Id())
		datadogList, err := conn.ListDatadog(&gofastly.ListDatadogInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Datadog logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dll := flattenDatadog(datadogList)

		for _, element := range dll {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), dll); err != nil {
			log.Printf("[WARN] Error setting Datadog logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *DatadogServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateDatadogInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between any we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.String(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Datadog Opts: %#v", opts)
	_, err := conn.UpdateDatadog(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *DatadogServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Datadog logging endpoint removal opts: %#v", opts)

	return deleteDatadog(conn, opts)
}

func createDatadog(conn *gofastly.Client, i *gofastly.CreateDatadogInput) error {
	_, err := conn.CreateDatadog(i)
	return err
}

func deleteDatadog(conn *gofastly.Client, i *gofastly.DeleteDatadogInput) error {
	err := conn.DeleteDatadog(i)

	errRes, ok := err.(*gofastly.HTTPError)
	if !ok {
		return err
	}

	// 404 response codes don't result in an error propagating because a 404 could
	// indicate that a resource was deleted elsewhere.
	if !errRes.IsNotFound() {
		return err
	}

	return nil
}

func flattenDatadog(datadogList []*gofastly.Datadog) []map[string]any {
	var dsl []map[string]any
	for _, dl := range datadogList {
		// Convert Datadog logging to a map for saving to state.
		ndl := map[string]any{
			"name":               dl.Name,
			"token":              dl.Token,
			"region":             dl.Region,
			"format":             dl.Format,
			"format_version":     dl.FormatVersion,
			"placement":          dl.Placement,
			"response_condition": dl.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range ndl {
			if v == "" {
				delete(ndl, k)
			}
		}

		dsl = append(dsl, ndl)
	}

	return dsl
}

func (h *DatadogServiceAttributeHandler) buildCreate(datadogMap any, serviceID string, serviceVersion int) *gofastly.CreateDatadogInput {
	df := datadogMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateDatadogInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		Token:             df["token"].(string),
		Region:            df["region"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *DatadogServiceAttributeHandler) buildDelete(datadogMap any, serviceID string, serviceVersion int) *gofastly.DeleteDatadogInput {
	df := datadogMap.(map[string]any)

	return &gofastly.DeleteDatadogInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
