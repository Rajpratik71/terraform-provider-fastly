---
layout: "fastly"
page_title: "Fastly: tls_certificate"
sidebar_current: "docs-fastly-resource-tls_certificate"
description: |-
Uploads a custom TLS certificate
---

# fastly_tls_certificate

Uploads a custom TLS certificate to Fastly to be used to terminate TLS traffic.

-> Each TLS certificate **must** have its corresponding private key uploaded _prior_ to uploading the certificate. This
can be achieved in Terraform using [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html)

## Example Usage

Basic usage:

```terraform
resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "cert" {
  key_algorithm   = tls_private_key.key.algorithm
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example.com"
  }

  is_ca_certificate     = true
  validity_period_hours = 360

  allowed_uses = [
    "cert_signing",
    "server_auth",
  ]

  dns_names = ["example.com"]
}

resource "fastly_tls_private_key" "key" {
  key_pem = tls_private_key.key.private_key_pem
  name    = "tf-demo"
}

resource "fastly_tls_certificate" "example" {
  name = "tf-demo"
  certificate_body = tls_self_signed_cert.cert.cert_pem
  depends_on = [fastly_tls_private_key.key] // The private key has to be present before the certificate can be uploaded
}
```

## Updating certificates

There are three scenarios for updating a certificate:

1. The certificate is about to expire but the private key stays the same.
2. The certificate is about to expire but the private key is changing.
3. The domains on the certificate are changing.

In the first scenario you only need to update the `certificate_body` attribute of the `fastly_tls_certificate` resource, while the other scenarios require a new private key (`fastly_tls_private_key`) and certificate (`fastly_tls_certificate`) to be generated.

When updating both the `fastly_tls_private_key` and `fastly_tls_certificate` resources, they should be done in multiple plan/apply steps to avoid potential downtime. The new certificate and associated private key must first be created so they exist alongside the currently active resources. Once the new resources have been created, then the `fastly_tls_activation` can be updated to point to the new certificate. Finally, the original key/certificate resources can be deleted.

## Import

A certificate can be imported using its Fastly certificate ID, e.g.

```sh
$ terraform import fastly_tls_certificate.demo xxxxxxxxxxx
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **certificate_body** (String) PEM-formatted certificate, optionally including any intermediary certificates.

### Optional

- **id** (String) The ID of this resource.
- **name** (String) Human-readable name used to identify the certificate. Defaults to the certificate's Common Name or first Subject Alternative Name entry.

### Read-Only

- **created_at** (String) Timestamp (GMT) when the certificate was created.
- **domains** (Set of String) All the domains (including wildcard domains) that are listed in the certificate's Subject Alternative Names (SAN) list.
- **issued_to** (String) The hostname for which a certificate was issued.
- **issuer** (String) The certificate authority that issued the certificate.
- **replace** (Boolean) A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.
- **serial_number** (String) A value assigned by the issuer that is unique to a certificate.
- **signature_algorithm** (String) The algorithm used to sign the certificate.
- **updated_at** (String) Timestamp (GMT) when the certificate was last updated.
