package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenBackendCompute(t *testing.T) {
	cases := []struct {
		serviceMetadata ServiceMetadata
		remote          []*gofastly.Backend
		local           []map[string]any
	}{
		{
			serviceMetadata: ServiceMetadata{
				serviceType: ServiceTypeCompute,
			},
			remote: []*gofastly.Backend{
				{
					Name:                "test.notexample.com",
					Address:             "www.notexample.com",
					OverrideHost:        "origin.example.com",
					Port:                uint(80),
					AutoLoadbalance:     true,
					BetweenBytesTimeout: uint(10000),
					ConnectTimeout:      uint(1000),
					ErrorThreshold:      uint(0),
					FirstByteTimeout:    uint(15000),
					MaxConn:             uint(200),
					HealthCheck:         "",
					UseSSL:              false,
					SSLCheckCert:        true,
					SSLHostname:         "",
					SSLCACert:           "",
					SSLCertHostname:     "",
					SSLSNIHostname:      "",
					SSLClientKey:        "",
					SSLClientCert:       "",
					MaxTLSVersion:       "",
					MinTLSVersion:       "",
					SSLCiphers:          "foo:bar:baz",
					Shield:              "lga-ny-us",
					Weight:              uint(100),
				},
			},
			local: []map[string]any{
				{
					"name":                  "test.notexample.com",
					"address":               "www.notexample.com",
					"override_host":         "origin.example.com",
					"port":                  80,
					"auto_loadbalance":      true,
					"between_bytes_timeout": 10000,
					"connect_timeout":       1000,
					"error_threshold":       0,
					"first_byte_timeout":    15000,
					"max_conn":              200,
					"healthcheck":           "",
					"use_ssl":               false,
					"ssl_check_cert":        true,
					"ssl_hostname":          "",
					"ssl_ca_cert":           "",
					"ssl_cert_hostname":     "",
					"ssl_sni_hostname":      "",
					"ssl_client_key":        "",
					"ssl_client_cert":       "",
					"max_tls_version":       "",
					"min_tls_version":       "",
					"ssl_ciphers":           "foo:bar:baz",
					"shield":                "lga-ny-us",
					"weight":                100,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBackend(c.remote, c.serviceMetadata)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceCompute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test1.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "version_comment", ""),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "backend.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				ResourceName:      "fastly_service_compute.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename", "imported"},
			},
		},
	})
}

func testAccCheckServiceComputeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("error listing services when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccServiceComputeConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }
  force_destroy = true
  activate = false
}`, name, domain)
}
