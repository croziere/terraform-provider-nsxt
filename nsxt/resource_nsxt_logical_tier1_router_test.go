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

func TestNSXLogicalTier1RouterBasic(t *testing.T) {

	name := fmt.Sprintf("test-nsx-logical-tier1-router")
	updateName := fmt.Sprintf("%s-update", name)
	testResourceName := "nsxt_logical_tier1_router.test"
	failoverMode := "PREEMPTIVE"
	haMode := "ACTIVE_STANDBY"
	edgeClusterName := EdgeClusterDefaultName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalTier1RouterCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalTier1RouterCreateTemplate(name, failoverMode, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier1RouterExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "enable_router_advertisement", "true"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_connected_routes", "true"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_static_routes", "true"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_nat_routes", "true"),
				),
			},
			{
				Config: testAccNSXLogicalTier1RouterUpdateTemplate(updateName, failoverMode, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier1RouterExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "enable_router_advertisement", "false"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_connected_routes", "false"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_static_routes", "false"),
					resource.TestCheckResourceAttr(testResourceName, "advertise_nat_routes", "false"),
				),
			},
		},
	})
}

func testAccNSXLogicalTier1RouterExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical tier1 router resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical tier1 router resource ID not set in resources")
		}

		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical tier1 router ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking verifying logical tier1 router existence. HTTP returned %d", resourceID, responseCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical tier1 router %s not found", displayName)
	}
}

func testAccNSXLogicalTier1RouterCheckDestroy(state *terraform.State, displayName string) error {

	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_tier1_router" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical tier1 router ID %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX logical tier1 router %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalTier1RouterCreateTemplate(name string, failoverMode string, haMode string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
     display_name = "%s"
}

resource "nsxt_logical_tier1_router" "test" {
	display_name = "%s"
	description = "Acceptance Test"
	failover_mode = "%s"
	high_availability_mode = "%s"
	edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
	enable_router_advertisement = "true"
	advertise_connected_routes = "true"
	advertise_static_routes = "true"
	advertise_nat_routes = "true"
	tags = [{scope = "scope1"
         tag = "tag1"}, 
        {scope = "scope2"
    	 tag = "tag2"}
	]

}`, edgeClusterName, name, failoverMode, haMode)
}

func testAccNSXLogicalTier1RouterUpdateTemplate(name string, failoverMode string, haMode string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
     display_name = "%s"
}

resource "nsxt_logical_tier1_router" "test" {
	display_name = "%s"
	description = "Acceptance Test Update"
	failover_mode = "%s"
	high_availability_mode = "%s"
	edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
	enable_router_advertisement = "false"
	advertise_connected_routes = "false"
	advertise_static_routes = "false"
	advertise_nat_routes = "false"
	tags = [{scope = "scope3"
	    	 tag = "tag3"}
	]
}`, edgeClusterName, name, failoverMode, haMode)
}