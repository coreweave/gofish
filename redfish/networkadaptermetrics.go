//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"context"
	"encoding/json"

	"github.com/coreweave/gofish/common"
)

// NetworkAdapterMetrics shall represent the network metrics for a single network adapter in a Redfish
// implementation.
type NetworkAdapterMetrics struct {
	common.Entity
	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// CPUCorePercent shall contain the device CPU core utilization as a percentage, typically '0' to '100'.
	CPUCorePercent *float64 `json:",omitempty"`
	// Description provides a description of this resource.
	Description string
	// HostBusRXPercent shall contain the host bus, such as PCIe, RX utilization as a percentage, typically '0' to
	// '100', which is calculated by dividing the total bytes received by the theoretical max.
	HostBusRXPercent *float64 `json:",omitempty"`
	// HostBusTXPercent shall contain the host bus, such as PCIe, TX utilization as a percentage, typically '0' to
	// '100', which is calculated by dividing the total bytes transmitted by the theoretical max.
	HostBusTXPercent *float64 `json:",omitempty"`
	// NCSIRXBytes shall contain the total number of NC-SI bytes received since reset, including both passthrough and
	// non-passthrough traffic.
	NCSIRXBytes *int `json:",omitempty"`
	// NCSIRXFrames shall contain the total number of NC-SI frames received since reset, including both passthrough and
	// non-passthrough traffic.
	NCSIRXFrames *int `json:",omitempty"`
	// NCSITXBytes shall contain the total number of NC-SI bytes sent since reset, including both passthrough and non-
	// passthrough traffic.
	NCSITXBytes *int `json:",omitempty"`
	// NCSITXFrames shall contain the total number of NC-SI frames sent since reset, including both passthrough and
	// non-passthrough traffic.
	NCSITXFrames *int `json:",omitempty"`
	// Oem shall contain the OEM extensions. All values for properties that this object contains shall conform to the
	// Redfish Specification-described requirements.
	OEM json.RawMessage `json:"Oem"`
	// RXBytes shall contain the total number of bytes received since reset, including host and remote management
	// passthrough traffic, and inclusive of all protocol overhead.
	RXBytes *int `json:",omitempty"`
	// RXMulticastFrames shall contain the total number of good multicast frames received since reset.
	RXMulticastFrames *int `json:",omitempty"`
	// RXUnicastFrames shall contain the total number of good unicast frames received since reset.
	RXUnicastFrames *int `json:",omitempty"`
	// TXBytes shall contain the total number of bytes transmitted since reset, including host and remote management
	// passthrough traffic, and inclusive of all protocol overhead.
	TXBytes *int `json:",omitempty"`
	// TXMulticastFrames shall contain the total number of good multicast frames transmitted since reset.
	TXMulticastFrames *int `json:",omitempty"`
	// TXUnicastFrames shall contain the total number of good unicast frames transmitted since reset.
	TXUnicastFrames *int `json:",omitempty"`
}

// GetNetworkAdapterMetrics will get a NetworkAdapterMetrics instance from the service.
func GetNetworkAdapterMetrics(c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*NetworkAdapterMetrics, error) {
	return GetNetworkAdapterMetricsWithContext(common.ContextOf(c), c, uri, queryOpts...)
}

// GetNetworkAdapterMetricsWithContext will get a NetworkAdapterMetrics instance from the service.
func GetNetworkAdapterMetricsWithContext(ctx context.Context, c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*NetworkAdapterMetrics, error) {
	return common.GetObjectWithContext[NetworkAdapterMetrics](ctx, c, uri, queryOpts...)
}

// ListReferencedNetworkAdapterMetrics gets the collection of NetworkAdapterMetrics from
// a provided reference.
func ListReferencedNetworkAdapterMetrics(c common.Client, link string, queryOpts ...common.QueryGroupOption) ([]*NetworkAdapterMetrics, error) {
	return ListReferencedNetworkAdapterMetricsWithContext(common.ContextOf(c), c, link, queryOpts...)
}

// ListReferencedNetworkAdapterMetricsWithContext gets the collection of NetworkAdapterMetrics from
// a provided reference.
func ListReferencedNetworkAdapterMetricsWithContext(ctx context.Context, c common.Client, link string, queryOpts ...common.QueryGroupOption) ([]*NetworkAdapterMetrics, error) {
	return common.GetCollectionObjectsWithContext[NetworkAdapterMetrics](ctx, c, link, queryOpts...)
}
