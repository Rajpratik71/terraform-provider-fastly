package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenGCS(t *testing.T) {
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	cases := []struct {
		remote []*gofastly.GCS
		local  []map[string]any
	}{
		{
			remote: []*gofastly.GCS{
				{
					Name:             "GCS collector",
					User:             "email@example.com",
					Bucket:           "bucketname",
					SecretKey:        secretKey,
					Format:           "log format",
					FormatVersion:    uint(2),
					Period:           3600,
					GzipLevel:        0,
					CompressionCodec: "zstd",
				},
			},
			local: []map[string]any{
				{
					"name":              "GCS collector",
					"user":              "email@example.com",
					"bucket_name":       "bucketname",
					"secret_key":        secretKey,
					"format":            "log format",
					"format_version":    uint(2),
					"period":            3600,
					"gzip_level":        0,
					"compression_codec": "zstd",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenGCS(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_gcslogging(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	gcsName := fmt.Sprintf("gcs %s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigGCS(name, gcsName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesGCS(&service, name, gcsName),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_gcslogging_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	gcsName := fmt.Sprintf("gcs %s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigComputeGCS(name, gcsName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLAttributesGCS(&service, name, gcsName),
				),
			},
		},
	})
}

func TestGcsloggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingGCS(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_gcs"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["secret_key"].Sensitive {
		t.Fatalf("Expected secret_key to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	email := "tf-test@fastly.com"
	secretKey, _ := generateKey()
	resetEnv := setGcsEnv(email, secretKey, t)
	defer resetEnv()

	result1, err1 := loggingResourceSchema["user"].DefaultFunc()
	if err1 != nil {
		t.Fatalf("Unexpected err %#v when calling email DefaultFunc", err1)
	}
	if result1 != email {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", email, result1)
	}

	result2, err2 := loggingResourceSchema["secret_key"].DefaultFunc()
	if err2 != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err2)
	}
	if result2 != secretKey {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", secretKey, result2)
	}
}

func testAccCheckFastlyServiceVCLAttributesGCS(service *gofastly.ServiceDetail, name, gcsName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		gcsList, err := conn.ListGCSs(&gofastly.ListGCSsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up GCSs for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(gcsList) != 1 {
			return fmt.Errorf("gcs missing, expected: 1, got: %d", len(gcsList))
		}

		if gcsList[0].Name != gcsName {
			return fmt.Errorf("gcs name mismatch, expected: %s, got: %#v", gcsName, gcsList[0].Name)
		}

		return nil
	}
}

func testAccServiceVCLConfigGCS(name, gcsName, secretKey string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name = "tf -test backend"
  }

  logging_gcs {
    name = "%s"
    user = "email@example.com"
    bucket_name = "bucketname"
    secret_key = %q
    format = "log format"
    response_condition = ""
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey)
}

func testAccServiceVCLConfigComputeGCS(name, gcsName, secretKey string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test-compute.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name = "tf -test backend"
  }

  logging_gcs {
    name = "%s"
    user = "email@example.com"
    bucket_name = "bucketname"
    secret_key = %q
    compression_codec = "zstd"
  }

 package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey)
}

func testAccServiceVCLConfigGCSEnv(name, gcsName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name = "tf -test backend"
  }

  logging_gcs {
    name = "%s"
    bucket_name = "bucketname"
    format = "log format"
    response_condition = ""
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName)
}

func setGcsEnv(email, secretKey string, t *testing.T) func() {
	e := getGcsEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_GCS_EMAIL", email); err != nil {
		t.Fatalf("Error setting env var FASTLY_GCS_EMAIL: %s", err)
	}
	if err := os.Setenv("FASTLY_GCS_SECRET_KEY", secretKey); err != nil {
		t.Fatalf("Error setting env var FASTLY_GCS_SECRET_KEY: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_GCS_EMAIL", e.Key); err != nil {
			t.Fatalf("Error resetting env var FASTLY_GCS_EMAIL: %s", err)
		}
		if err := os.Setenv("FASTLY_GCS_SECRET_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var FASTLY_GCS_SECRET_KEY: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentGcsEnv struct {
	Key, Secret string
}

func getGcsEnv() *currentGcsEnv {
	// Grab any existing Fastly GCS keys and preserve, in the off chance
	// they're actually set in the enviornment
	return &currentGcsEnv{
		Key:    os.Getenv("FASTLY_GCS_EMAIL"),
		Secret: os.Getenv("FASTLY_GCS_SECRET_KEY"),
	}
}
