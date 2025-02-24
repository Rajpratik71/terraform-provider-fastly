---
layout: "fastly"
page_title: "Fastly: fastly_tls_platform_certificate"
sidebar_current: "docs-fastly-datasource-tls_platform_certificate"
description: |-
Get information on Fastly Platform TLS certificate.
---

# fastly_tls_platform_certificate

Use this data source to get information of a Platform TLS certificate for use with other resources.

~> **Warning:** The data source's filters are applied using an **AND** boolean operator, so depending on the combination
of filters, they may become mutually exclusive. The exception to this is `id` which must not be specified in combination
with any of the others.

~> **Note:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single key.

## Example Usage

```terraform
data "fastly_tls_platform_certificate" "example" {
  domains = ["example.com"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **domains** (Set of String) Domains that are listed in any certificate's Subject Alternative Names (SAN) list.
- **id** (String) Unique ID assigned to certificate by Fastly. Conflicts with all the other filters.

### Read-Only

- **configuration_id** (String) ID of TLS configuration used to terminate TLS traffic.
- **created_at** (String) Timestamp (GMT) when the certificate was created.
- **not_after** (String) Timestamp (GMT) when the certificate will expire.
- **not_before** (String) Timestamp (GMT) when the certificate will become valid.
- **replace** (Boolean) A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.
- **updated_at** (String) Timestamp (GMT) when the certificate was last updated.