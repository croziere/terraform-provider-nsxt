/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause

   Generated by: https://github.com/swagger-api/swagger-codegen.git */

package administration

// Proton service properties
type ProtonServiceProperties struct {

	// Proton logging level
	LoggingLevel string `json:"logging_level"`

	// Package logging levels
	PackageLoggingLevel []ProtonPackageLoggingLevels `json:"package_logging_level,omitempty"`
}