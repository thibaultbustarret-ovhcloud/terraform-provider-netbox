package netbox

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/extras"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	log "github.com/sirupsen/logrus"
)

func TestAccNetboxConfigTemplate_basic(t *testing.T) {
	testSlug := "config_template"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_config_template" "test" {
	name = "%[1]s"
	description = "%[1]s description"
	template_code = "hostname {{ name }}"
	environment_params = jsonencode({"name" = "my-hostname"})
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_config_template.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_config_template.test", "description", fmt.Sprintf("%s description", testName)),
					resource.TestCheckResourceAttr("netbox_config_template.test", "template_code", "hostname {{ name }}"),
					resource.TestCheckResourceAttr("netbox_config_template.test", "environment_params", "{\"name\":\"my-hostname\"}"),
				),
			},
			{
				ResourceName:      "netbox_config_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_config_template", &resource.Sweeper{
		Name:         "netbox_config_template",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*client.NetBoxAPI)
			params := extras.NewExtrasConfigTemplatesListParams()
			res, err := api.Extras.ExtrasConfigTemplatesList(params, nil)
			if err != nil {
				return err
			}
			for _, tmpl := range res.GetPayload().Results {
				if strings.HasPrefix(*tmpl.Name, testPrefix) {
					deleteParams := extras.NewExtrasConfigTemplatesDeleteParams().WithID(tmpl.ID)
					_, err := api.Extras.ExtrasConfigTemplatesDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted a config template")
				}
			}
			return nil
		},
	})
}
