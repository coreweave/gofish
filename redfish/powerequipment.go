//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"context"
	"encoding/json"

	"github.com/coreweave/gofish/common"
)

type PowerEquipmentLinks struct {
	ManagedBy      common.Links
	ManagedByCount int `json:"ManagedBy@odata.count"`
	Oem            json.RawMessage
}

type PowerEquipmentActions struct {
	Oem json.RawMessage // OEM actions will be stored here
}

// PowerEquipment shall be used to represent
// the set of power equipment for a Redfish implementation.
type PowerEquipment struct {
	common.Resource

	// The status and health of the resource and its subordinate or dependent resources.
	Status common.Status

	ElectricalBusesLink  common.Link `json:"ElectricalBuses"`
	FloorPDUsLink        common.Link `json:"FloorPDUs"`
	PowerShelvesLink     common.Link `json:"PowerShelves"`
	RackPDUsLink         common.Link `json:"RackPDUs"`
	SwitchgearLink       common.Link `json:"Switchgear"`
	TransferSwitchesLink common.Link `json:"TransferSwitches"`

	Links   PowerEquipmentLinks   `json:"Links"`
	Actions PowerEquipmentActions `json:"Actions"`
}

// GetPowerEquipment will get a PowerEquipment instance from the Redfish service.
func GetPowerEquipment(c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*PowerEquipment, error) {
	return GetPowerEquipmentWithContext(common.ContextOf(c), c, uri, queryOpts...)
}

// GetPowerEquipmentWithContext will get a PowerEquipment instance from the Redfish service.
func GetPowerEquipmentWithContext(ctx context.Context, c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*PowerEquipment, error) {
	return common.GetObjectWithContext[PowerEquipment](ctx, c, uri, queryOpts...)
}

// ManagedBy gets the collection of managers of this PowerEquipment
func (powerEquipment *PowerEquipment) ManagedBy(queryOpts ...common.QueryGroupOption) ([]*Manager, error) {
	return powerEquipment.ManagedByWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// ManagedByWithContext gets the collection of managers of this PowerEquipment
func (powerEquipment *PowerEquipment) ManagedByWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*Manager, error) {
	if len(powerEquipment.Links.ManagedBy) == 0 {
		return nil, nil
	}

	return common.GetObjectsWithContext[Manager](ctx, powerEquipment.GetClient(), powerEquipment.Links.ManagedBy.ToStrings(), queryOpts...)
}

// ElectricalBuses gets the collection that contains a set of electrical bus units.
func (powerEquipment *PowerEquipment) ElectricalBuses(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.ElectricalBusesWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// ElectricalBusesWithContext gets the collection that contains a set of electrical bus units.
func (powerEquipment *PowerEquipment) ElectricalBusesWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.ElectricalBusesLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.ElectricalBusesLink.String(), queryOpts...)
}

// FloorPDUs gets the collection that contains a set of floor power distribution units.
func (powerEquipment *PowerEquipment) FloorPDUs(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.FloorPDUsWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// FloorPDUsWithContext gets the collection that contains a set of floor power distribution units.
func (powerEquipment *PowerEquipment) FloorPDUsWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.FloorPDUsLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.FloorPDUsLink.String(), queryOpts...)
}

// PowerShelves gets the collection that contains a set of power shelves.
func (powerEquipment *PowerEquipment) PowerShelves(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.PowerShelvesWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// PowerShelvesWithContext gets the collection that contains a set of power shelves.
func (powerEquipment *PowerEquipment) PowerShelvesWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.PowerShelvesLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.PowerShelvesLink.String(), queryOpts...)
}

// RackPDUs gets the collection that contains a set of rack-level power distribution units.
func (powerEquipment *PowerEquipment) RackPDUs(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.RackPDUsWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// RackPDUsWithContext gets the collection that contains a set of rack-level power distribution units.
func (powerEquipment *PowerEquipment) RackPDUsWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.RackPDUsLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.RackPDUsLink.String(), queryOpts...)
}

// Switchgear gets the collection that contains a set of switchgear.
func (powerEquipment *PowerEquipment) Switchgear(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.SwitchgearWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// SwitchgearWithContext gets the collection that contains a set of switchgear.
func (powerEquipment *PowerEquipment) SwitchgearWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.SwitchgearLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.SwitchgearLink.String(), queryOpts...)
}

// TransferSwitches gets the collection that contains a set of transfer switches.
func (powerEquipment *PowerEquipment) TransferSwitches(queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	return powerEquipment.TransferSwitchesWithContext(common.ContextOf(powerEquipment.GetClient()), queryOpts...)
}

// TransferSwitchesWithContext gets the collection that contains a set of transfer switches.
func (powerEquipment *PowerEquipment) TransferSwitchesWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*PowerDistribution, error) {
	if powerEquipment.TransferSwitchesLink.IsZero() {
		return nil, nil
	}
	return ListReferencedPowerDistributionUnitsWithContext(ctx, powerEquipment.GetClient(), powerEquipment.TransferSwitchesLink.String(), queryOpts...)
}
