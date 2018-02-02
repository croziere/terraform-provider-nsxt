/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vmware-nsxt"
	"net/http"
	"testing"
)

func TestNSXIcmpTypeNsServiceBasic(t *testing.T) {

	serviceName := fmt.Sprintf("test-nsx-icmp-service")
	updateServiceName := fmt.Sprintf("%s-update", serviceName)
	testResourceName := "nsxt_icmp_type_ns_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIcmpServiceCheckDestroy(state, serviceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIcmpServiceCreateTemplate(serviceName, "ICMPv4", 5, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIcmpServiceExists(serviceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "icmp service"),
					resource.TestCheckResourceAttr(testResourceName, "protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
				),
			},
			{
				Config: testAccNSXIcmpServiceCreateTemplate(updateServiceName, "ICMPv6", 3, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIcmpServiceExists(updateServiceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateServiceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "icmp service"),
					resource.TestCheckResourceAttr(testResourceName, "protocol", "ICMPv6"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func testAccNSXIcmpServiceExists(display_name string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX icmp service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX icmp service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadIcmpTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving icmp service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if icmp service %s exists. HTTP return code was %d", resourceID, responseCode)
		}

		if display_name == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX icmp ns service %s wasn't found", display_name)
	}
}

func testAccNSXIcmpServiceCheckDestroy(state *terraform.State, display_name string) error {

	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_icmp_set_ns_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadIcmpTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving L4 ns service ID %s. Error: %v", resourceID, err)
		}

		if display_name == service.DisplayName {
			return fmt.Errorf("NSX L4 ns service %s still exists", display_name)
		}
	}
	return nil
}

func testAccNSXIcmpServiceCreateTemplate(serviceName string, protocol string, icmp_type int, icmp_code int) string {
	return fmt.Sprintf(`
resource "nsxt_icmp_type_ns_service" "test" {
    description = "icmp service"
    display_name = "%s"
    protocol = "%s"
    icmp_type = "%d"
    icmp_code = "%d"
    tags = [{scope = "scope1"
             tag = "tag1"}
    ]
}`, serviceName, protocol, icmp_type, icmp_code)
}
