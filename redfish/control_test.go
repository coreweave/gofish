//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/coreweave/gofish/common"
)

var controlBody = `{
	"@odata.type": "#Control.v1_5_0.Control",
	"Id": "PowerLimit",
	"Name": "System Power Limit",
	"PhysicalContext": "Chassis",
	"ControlType": "Power",
	"ControlMode": "Automatic",
	"SetPoint": 500,
	"SetPointUnits": "W",
	"AllowableMax": 1000,
	"AllowableMin": 150,
	"Sensor": {
	  "Reading": 374,
	  "DataSourceUri": "/redfish/v1/Chassis/1U/Sensors/TotalPower"
	},
	"Status": {
	  "Health": "OK",
	  "State": "Enabled"
	},
	"@odata.id": "/redfish/v1/Chassis/1U/Controls/PowerLimit"
  }`

// TestControl tests the parsing of Control objects.
func TestControl(t *testing.T) {
	var result Control
	err := json.NewDecoder(strings.NewReader(controlBody)).Decode(&result)

	if err != nil {
		t.Errorf("Error decoding JSON: %s", err)
	}

	assertEquals(t, "PowerLimit", result.ID)
	assertEquals(t, "System Power Limit", result.Name)
	assertEquals(t, "Chassis", string(result.PhysicalContext))
	assertEquals(t, "Power", string(result.ControlType))
	assertEquals(t, "Automatic", string(result.ControlMode))
	assertEquals(t, "W", result.SetPointUnits)
	assertEquals(t, "/redfish/v1/Chassis/1U/Sensors/TotalPower", result.Sensor.DataSourceURI)

	if result.SetPoint == nil {
		t.Error("Expected SetPoint to be set")
	} else if *result.SetPoint != 500 {
		t.Errorf("Unexpected set point, got %.2f", *result.SetPoint)
	}

	if result.AllowableMax == nil {
		t.Error("Expected AllowableMax to be set")
	} else if *result.AllowableMax != 1000 {
		t.Errorf("Unexpected allowable max, got %.2f", *result.AllowableMax)
	}

	if result.AllowableMin == nil {
		t.Error("Expected AllowableMin to be set")
	} else if *result.AllowableMin != 150 {
		t.Errorf("Unexpected allowable min, got %.2f", *result.AllowableMin)
	}

	if *result.Sensor.Reading != 374 {
		t.Errorf("Unexpected sensor reading, got %.2f", *result.Sensor.Reading)
	}
}

func TestControlSingleExcerptSetPointPresence(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantPresent bool
		wantValue   float64
	}{
		{
			name:        "explicit zero",
			body:        `{"SetPoint":0}`,
			wantPresent: true,
			wantValue:   0,
		},
		{
			name:        "explicit value",
			body:        `{"SetPoint":100}`,
			wantPresent: true,
			wantValue:   100,
		},
		{
			name:        "omitted",
			body:        `{}`,
			wantPresent: false,
		},
		{
			name:        "null",
			body:        `{"SetPoint":null}`,
			wantPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ControlSingleExcerpt
			if err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&result); err != nil {
				t.Fatalf("Error decoding JSON: %s", err)
			}

			if gotPresent := result.SetPoint != nil; gotPresent != tt.wantPresent {
				t.Fatalf("SetPoint presence = %t, want %t", gotPresent, tt.wantPresent)
			}
			if result.SetPoint != nil && *result.SetPoint != tt.wantValue {
				t.Fatalf("SetPoint = %.2f, want %.2f", *result.SetPoint, tt.wantValue)
			}
		})
	}
}

// TestControlUpdateSetPoint ensures a pointer SetPoint PATCHes only when it changes, including from absent.
func TestControlUpdateSetPoint(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setPoint    float64
		wantPayload string
	}{
		{name: "changed", body: `{"@odata.id": "/redfish/v1/Controls/1", "SetPoint": 500}`, setPoint: 600, wantPayload: "map[SetPoint:600]"},
		{name: "unchanged", body: `{"@odata.id": "/redfish/v1/Controls/1", "SetPoint": 500}`, setPoint: 500},
		{name: "absent set to zero", body: `{"@odata.id": "/redfish/v1/Controls/1"}`, setPoint: 0, wantPayload: "map[SetPoint:0]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Control
			if err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&result); err != nil {
				t.Fatalf("Error decoding JSON: %s", err)
			}

			testClient := &common.TestClient{}
			result.SetClient(testClient)
			setPoint := tt.setPoint
			result.SetPoint = &setPoint
			if err := result.Update(); err != nil {
				t.Fatalf("Error making Update call: %s", err)
			}

			calls := testClient.CapturedCalls()
			if tt.wantPayload == "" {
				if len(calls) != 0 {
					t.Fatalf("Expected no PATCH, got %d calls", len(calls))
				}
				return
			}
			if len(calls) != 1 || calls[0].Payload != tt.wantPayload {
				t.Fatalf("Expected one PATCH with %s, got %+v", tt.wantPayload, calls)
			}
		})
	}
}
