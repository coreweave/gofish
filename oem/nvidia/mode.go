//
// SPDX-License-Identifier: BSD-3-Clause
//

package nvidia

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/coreweave/gofish/common"
)

// ModeValue is the operating mode reported by the NVIDIA DPU OEM mode resource.
type ModeValue string

const (
	// DPUMode configures the device as a DPU.
	DPUMode ModeValue = "DpuMode"
	// NICMode configures the device as a NIC.
	NICMode ModeValue = "NicMode"
)

// Mode is the NVIDIA OEM mode resource for a DPU system.
type Mode struct {
	common.Entity

	Mode ModeValue

	setTarget string
}

// UnmarshalJSON unmarshals a Mode object from the raw JSON.
func (m *Mode) UnmarshalJSON(b []byte) error {
	type temp Mode
	var t struct {
		temp
		Actions struct {
			Set common.ActionTarget `json:"#Mode.Set"`
		}
	}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	*m = Mode(t.temp)
	m.setTarget = t.Actions.Set.Target

	return nil
}

// GetMode will get the Mode instance from the service.
func GetMode(c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*Mode, error) {
	return GetModeWithContext(common.ContextOf(c), c, uri, queryOpts...)
}

// GetModeWithContext will get the Mode instance from the service.
func GetModeWithContext(ctx context.Context, c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*Mode, error) {
	return common.GetObjectWithContext[Mode](ctx, c, uri, queryOpts...)
}

// SetMode performs the Mode.Set action.
func (m *Mode) SetMode(mode ModeValue) error {
	return m.SetModeWithContext(common.ContextOf(m.GetClient()), mode)
}

// SetModeWithContext performs the Mode.Set action.
func (m *Mode) SetModeWithContext(ctx context.Context, mode ModeValue) error {
	if m.setTarget == "" {
		return errors.New("Mode.Set is not supported by this system")
	}

	return m.PostWithContext(ctx, m.setTarget, map[string]any{"Mode": mode})
}
