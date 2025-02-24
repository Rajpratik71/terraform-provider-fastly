package fastly

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenDirectors(t *testing.T) {
	cases := []struct {
		remoteDirector []*gofastly.Director
		local          []map[string]any
	}{
		{
			remoteDirector: []*gofastly.Director{
				{
					Name:    "somedirector",
					Type:    3,
					Quorum:  75,
					Retries: 10,
					Backends: []string{
						"somebackend",
					},
				},
			},
			local: []map[string]any{
				{
					"name":     "somedirector",
					"type":     3,
					"quorum":   75,
					"retries":  10,
					"backends": schema.NewSet(schema.HashString, []any{"somebackend"}),
				},
			},
		},
		{
			remoteDirector: []*gofastly.Director{
				{
					Name: "somedirector",
					Backends: []string{
						"somebackend",
						"someotherbackend",
					},
				},
				{
					Name: "someotherdirector",
					Backends: []string{
						"somebackend",
						"someotherbackend",
					},
				},
			},
			local: []map[string]any{
				{
					"name":     "somedirector",
					"backends": schema.NewSet(schema.HashString, []any{"somebackend", "someotherbackend"}),
				},
				{
					"name":     "someotherdirector",
					"backends": schema.NewSet(schema.HashString, []any{"somebackend", "someotherbackend"}),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDirectors(c.remoteDirector)
		// loop, because deepequal wont work with our sets
		expectedCount := len(c.local)
		var found int
		for _, o := range out {
			for _, l := range c.local {
				if o["name"].(string) == l["name"].(string) {
					found++
					if o["backends"] == nil && l["backends"] != nil {
						t.Fatalf("output backends are nil, local are not")
					}

					if o["backends"] != nil {
						oex := o["backends"].(*schema.Set)
						lex := l["backends"].(*schema.Set)
						if !oex.Equal(lex) {
							t.Fatalf("Backends don't match, expected: %#v, got: %#v", lex, oex)
						}
					}
				}
			}
		}

		if found != expectedCount {
			t.Fatalf("Found and expected mismatch: %d / %d", found, expectedCount)
		}
	}
}

// This test validates that two directors are created successfully,
// and in the next Terraform run the first director is updated while
// the second director is unchanged and a third director is added.
// In the final test, the first director is removed while the second
// director is unchanged and one backend for the third director is removed.
func TestAccFastlyServiceVCL_directors_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	directorDeveloper := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_developer",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
		Backends:       []string{"developer"},
	}

	directorApps := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_apps",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
		Backends:       []string{"apps"},
	}

	dbDeveloper := gofastly.DirectorBackend{
		Director: "director_developer",
		Backend:  "developer",
	}

	dbApps := gofastly.DirectorBackend{
		Director: "director_apps",
		Backend:  "apps",
	}

	directorDeveloperUpdated := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_developer",
		Type:           4,
		Quorum:         30,
		Capacity:       100,
		Retries:        10,
		Backends:       []string{"developer_updated"},
	}

	directorWWWDemo := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_www_demo",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
		Backends:       []string{"demo", "www"},
	}

	dbDeveloperUpdated := gofastly.DirectorBackend{
		Director: "director_developer",
		Backend:  "developer_updated",
	}

	dbWWW := gofastly.DirectorBackend{
		Director: "director_www_demo",
		Backend:  "www",
	}

	dbDemo := gofastly.DirectorBackend{
		Director: "director_www_demo",
		Backend:  "demo",
	}

	directorWWWDemo2 := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_www_demo",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
		Backends:       []string{"www"},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLDirectorsConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDirectorsAttributes(&service, []*gofastly.Director{&directorDeveloper, &directorApps}),
					testAccCheckFastlyServiceVCLDirectorBackends(&service, []*gofastly.DirectorBackend{&dbDeveloper, &dbApps}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "director.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLDirectorsConfigUpdate1(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDirectorsAttributes(&service, []*gofastly.Director{&directorDeveloperUpdated, &directorApps, &directorWWWDemo}),
					testAccCheckFastlyServiceVCLDirectorBackends(&service, []*gofastly.DirectorBackend{&dbDeveloperUpdated, &dbApps, &dbWWW, &dbDemo}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "director.#", "3"),
				),
			},

			{
				Config: testAccServiceVCLDirectorsConfigUpdate2(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDirectorsAttributes(&service, []*gofastly.Director{&directorApps, &directorWWWDemo2}),
					testAccCheckFastlyServiceVCLDirectorBackends(&service, []*gofastly.DirectorBackend{&dbApps, &dbWWW}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "director.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLDirectorsAttributes(service *gofastly.ServiceDetail, directors []*gofastly.Director) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		directorList, err := conn.ListDirectors(&gofastly.ListDirectorsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Directors for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(directorList) != len(directors) {
			return fmt.Errorf("director count mismatch, expected (%d), got (%d)", len(directors), len(directorList))
		}

		var found int
		for _, d := range directors {
			for _, ld := range directorList {
				if d.Name == ld.Name {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ID
					d.ServiceVersion = service.ActiveVersion.Number
					ld.CreatedAt = nil
					ld.UpdatedAt = nil
					sort.Strings(d.Backends)
					sort.Strings(ld.Backends)
					if !reflect.DeepEqual(d, ld) {
						return fmt.Errorf("bad match Director match, expected (%#v), got (%#v)", d, ld)
					}
					found++
				}
			}
		}

		if found != len(directors) {
			return fmt.Errorf("error matching Director rules")
		}

		return nil
	}
}

func testAccCheckFastlyServiceVCLDirectorBackends(service *gofastly.ServiceDetail, directorBackends []*gofastly.DirectorBackend) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		directorList, err := conn.ListDirectors(&gofastly.ListDirectorsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Directors for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expectedDirectorBackends := make(map[string][]string)

		for _, director := range directorList {
			matchedDirector := false
			for _, directorBackend := range directorBackends {
				if director.Name == directorBackend.Director {
					matchedDirector = true
					backends := expectedDirectorBackends[director.Name]
					expectedDirectorBackends[director.Name] = append(backends, directorBackend.Backend)
				}
			}
			if !matchedDirector {
				return fmt.Errorf("didn't find the expected director: %s", director.Name)
			}

			for _, directorBackend := range directorBackends {
				if director.Name == directorBackend.Director {
					if len(director.Backends) != len(expectedDirectorBackends[director.Name]) {
						return fmt.Errorf("didn't find the same number of director backends: %s", director.Name)
					}
					for _, expectedBackend := range expectedDirectorBackends[director.Name] {
						matchedBackend := false
						for _, backend := range director.Backends {
							if backend == expectedBackend {
								matchedBackend = true
							}
						}
						if !matchedBackend {
							return fmt.Errorf("didn't find the expected backend: %s", expectedBackend)
						}
					}
				}
			}
		}

		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion, err)
		}

		var directorBackendList []*gofastly.DirectorBackend

		for _, director := range directorList {
			for _, backend := range backendList {
				directorBackendGet, err := conn.GetDirectorBackend(&gofastly.GetDirectorBackendInput{
					ServiceID:      service.ID,
					ServiceVersion: service.ActiveVersion.Number,
					Director:       director.Name,
					Backend:        backend.Name,
				})
				if err == nil {
					directorBackendList = append(directorBackendList, directorBackendGet)
				}
			}
		}

		if len(directorBackends) != len(directorBackendList) {
			return fmt.Errorf("director count mismatch, expected (%d), got (%d)", len(directorBackends), len(directorBackendList))
		}

		var found int
		for _, db := range directorBackends {
			for _, ldb := range directorBackendList {
				if db.Director == ldb.Director && db.Backend == ldb.Backend {
					// we don't know these things ahead of time, so populate them now
					db.ServiceID = service.ID
					db.ServiceVersion = service.ActiveVersion.Number
					ldb.CreatedAt = nil
					ldb.UpdatedAt = nil
					if !reflect.DeepEqual(db, ldb) {
						return fmt.Errorf("bad Director Backend match, expected (%#v), got (%#v)", db, ldb)
					}
					found++
				}
			}
		}

		if found != len(directorBackends) {
			return fmt.Errorf("error matching Director Backend rules, expected (%#v), got (%#v)", len(directorBackendList), found)
		}

		return nil
	}
}

func testAccServiceVCLDirectorsConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "developer.fastly.com"
    name    = "developer"
  }

  backend {
    address = "apps.fastly.com"
    name    = "apps"
    weight  = 1
  }

  director {
    name = "director_developer"
    type = 3
    backends = [ "developer" ]
  }

  director {
    name = "director_apps"
    type = 3
    backends = [ "apps" ]
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLDirectorsConfigUpdate1(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "developer.fastly.com"
    name    = "developer_updated"
  }

  backend {
    address = "apps.fastly.com"
    name    = "apps"
    weight  = 9
  }

  backend {
    address = "www.fastly.com"
    name    = "www"
  }

  backend {
    address = "www.fastlydemo.net"
    name    = "demo"
  }

  director {
    name = "director_developer"
    type = 4
    quorum = 30
    retries = 10
    backends = [ "developer_updated" ]
  }

  director {
    name = "director_apps"
    type = 3
    backends = [ "apps" ]
  }

  director {
    name = "director_www_demo"
    type = 3
    backends = [ "www", "demo" ]
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLDirectorsConfigUpdate2(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "apps.fastly.com"
    name    = "apps"
    weight  = 9
  }

  backend {
    address = "www.fastly.com"
    name    = "www"
  }

  director {
    name = "director_apps"
    type = 3
    backends = [ "apps" ]
  }

  director {
    name = "director_www_demo"
    type = 3
    backends = [ "www" ]
  }

  force_destroy = true
}`, name, domain)
}
