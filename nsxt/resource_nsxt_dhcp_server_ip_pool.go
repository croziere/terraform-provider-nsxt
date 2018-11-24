/* Copyright © 2018 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"net/http"
	"strings"
)

func resourceNsxtDhcpServerIPPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtDhcpServerIPPoolCreate,
		Read:   resourceNsxtDhcpServerIPPoolRead,
		Update: resourceNsxtDhcpServerIPPoolUpdate,
		Delete: resourceNsxtDhcpServerIPPoolDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtDhcpServerIPPoolImport,
		},

		Schema: map[string]*schema.Schema{
			"display_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"logical_dhcp_server_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Id of dhcp server this pool is serving",
				Required:    true,
				ForceNew:    true,
			},
			"dhcp_option_121":     getDhcpOptions121Schema(),
			"dhcp_generic_option": getDhcpGenericOptionsSchema(),
			"ip_range":            getIPRangesSchema(),
			"gateway_ip": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Gateway ip",
				Optional:     true,
				ValidateFunc: validateSingleIP(),
			},
			"lease_time": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Lease time, in seconds",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 4294967295),
				Default:      86400,
			},
			"error_threshold": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Error threshold",
				Optional:     true,
				ValidateFunc: validation.IntBetween(80, 100),
				Default:      100,
			},
			"warning_threshold": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Warning threshold",
				Optional:     true,
				ValidateFunc: validation.IntBetween(50, 80),
				Default:      80,
			},

			"tag":      getTagsSchema(),
			"revision": getRevisionSchema(),
		},
	}
}

func resourceNsxtDhcpServerIPPoolCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	displayName := d.Get("display_name").(string)
	gatewayIP := d.Get("gateway_ip").(string)
	serverID := d.Get("logical_dhcp_server_id").(string)
	leaseTime := int64(d.Get("lease_time").(int))
	errorThreshold := int64(d.Get("error_threshold").(int))
	warningThreshold := int64(d.Get("warning_threshold").(int))

	opt121Routes := getDhcpOptions121(d)
	var opt121 *manager.DhcpOption121
	if opt121Routes != nil {
		opt121 = &manager.DhcpOption121{
			StaticRoutes: opt121Routes,
		}
	}
	tags := getTagsFromSchema(d)
	pool := manager.DhcpIpPool{
		DisplayName: displayName,
		GatewayIp:   gatewayIP,
		Options: &manager.DhcpOptions{
			Option121: opt121,
			Others:    getDhcpGenericOptions(d),
		},
		LeaseTime:        leaseTime,
		ErrorThreshold:   errorThreshold,
		WarningThreshold: warningThreshold,
		AllocationRanges: getIPRangesFromSchema(d),
		Tags:             tags,
	}

	createdPool, resp, err := nsxClient.ServicesApi.CreateDhcpIpPool(nsxClient.Context, serverID, pool)

	if err != nil {
		return fmt.Errorf("Error during DhcpIPPool create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during DhcpIPPool create: %v", resp.StatusCode)
	}

	d.SetId(createdPool.Id)

	return resourceNsxtDhcpServerIPPoolRead(d, m)
}

func resourceNsxtDhcpServerIPPoolRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	serverID := d.Get("logical_dhcp_server_id").(string)
	if id == "" || serverID == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	pool, resp, err := nsxClient.ServicesApi.ReadDhcpIpPool(nsxClient.Context, serverID, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpIPPool read: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpIPPool %s not found", id)
		d.SetId("")
		return nil
	}

	d.Set("revision", pool.Revision)
	d.Set("display_name", pool.DisplayName)
	setTagsInSchema(d, pool.Tags)
	d.Set("logical_dhcp_server_id", serverID)
	d.Set("gateway_ip", pool.GatewayIp)
	setIPRangesInSchema(d, pool.AllocationRanges)
	d.Set("lease_time", pool.LeaseTime)
	d.Set("error_threshold", pool.ErrorThreshold)
	d.Set("warning_threshold", pool.WarningThreshold)

	if pool.Options != nil && pool.Options.Option121 != nil {
		err = setDhcpOptions121InSchema(d, pool.Options.Option121.StaticRoutes)
		if err != nil {
			return fmt.Errorf("Error during DhcpIPPool read option 121: %v", err)
		}
		err = setDhcpGenericOptionsInSchema(d, pool.Options.Others)
		if err != nil {
			return fmt.Errorf("Error during DhcpIPPool read generic options: %v", err)
		}
	} else {
		var emptyDhcpOpt121 []map[string]interface{}
		var emptyDhcpGenOpt []map[string]interface{}
		d.Set("dhcp_option_121", emptyDhcpOpt121)
		d.Set("dhcp_generic_option", emptyDhcpGenOpt)
	}

	return nil
}

func resourceNsxtDhcpServerIPPoolUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	serverID := d.Get("logical_dhcp_server_id").(string)
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	displayName := d.Get("display_name").(string)
	gatewayIP := d.Get("gateway_ip").(string)
	leaseTime := int64(d.Get("lease_time").(int))
	errorThreshold := int64(d.Get("error_threshold").(int))
	warningThreshold := int64(d.Get("warning_threshold").(int))

	opt121Routes := getDhcpOptions121(d)
	var opt121 *manager.DhcpOption121
	if opt121Routes != nil {
		opt121 = &manager.DhcpOption121{
			StaticRoutes: opt121Routes,
		}
	}
	tags := getTagsFromSchema(d)
	pool := manager.DhcpIpPool{
		DisplayName: displayName,
		GatewayIp:   gatewayIP,
		Options: &manager.DhcpOptions{
			Option121: opt121,
			Others:    getDhcpGenericOptions(d),
		},
		LeaseTime:        leaseTime,
		ErrorThreshold:   errorThreshold,
		WarningThreshold: warningThreshold,
		AllocationRanges: getIPRangesFromSchema(d),
		Tags:             tags,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateDhcpIpPool(nsxClient.Context, serverID, id, pool)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during DhcpIPPool update: %v", err)
	}

	return resourceNsxtDhcpServerIPPoolRead(d, m)
}

func resourceNsxtDhcpServerIPPoolDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	serverID := d.Get("logical_dhcp_server_id").(string)
	if id == "" || serverID == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteDhcpIpPool(nsxClient.Context, serverID, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpIPPool delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpIPPool %s not found", id)
		d.SetId("")
	}
	return nil
}

func resourceNsxtDhcpServerIPPoolImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <dhcp-server-id>/<ip-pool-id> as an input")
	}

	d.SetId(s[1])
	d.Set("logical_dhcp_server_id", s[0])

	return []*schema.ResourceData{d}, nil

}