/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause

   Generated by: https://github.com/swagger-api/swagger-codegen.git */

package manager

type LogicalPortMacTableEntry struct {

	// The MAC address
	MacAddress string `json:"mac_address"`

	// The type of the MAC address
	MacType string `json:"mac_type"`
}