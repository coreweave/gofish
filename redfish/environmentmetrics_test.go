//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

var environmentMetricsBody = `{
	"@odata.type": "#EnvironmentMetrics.v1_3_1.EnvironmentMetrics",
	"ID": "Metrics1",
	"Name": "Processor Environment Metrics",
	"TemperatureCelsius": {
	  "DataSourceUri": "/redfish/v1/Chassis/1U/Sensors/CPU1Temp",
	  "Reading": 44
	},
	"PowerWatts": {
	  "DataSourceUri": "/redfish/v1/Chassis/1U/Sensors/CPU1Power",
	  "Reading": 12.87
	},
	"FanSpeedsPercent": [
	  {
		"DataSourceUri": "/redfish/v1/Chassis/1U/Sensors/CPUFan1",
		"DeviceName": "CPU #1 Fan Speed",
		"Reading": 80
	  }
	],
	"@odata.id": "/redfish/v1/Systems/437XR1138R2/Processors/1/EnvironmentMetrics"
  }`

// TestEnvironmentMetrics tests the parsing of EnvironmentMetrics objects.
func TestEnvironmentMetrics(t *testing.T) {
	var result EnvironmentMetrics
	err := json.NewDecoder(strings.NewReader(environmentMetricsBody)).Decode(&result)

	if err != nil {
		t.Errorf("Error decoding JSON: %s", err)
	}

	assertEquals(t, "Metrics1", result.ID)
	assertEquals(t, "Processor Environment Metrics", result.Name)
	assertEquals(t, "/redfish/v1/Chassis/1U/Sensors/CPU1Temp", result.TemperatureCelsius.DataSourceURI)
	assertEquals(t, "/redfish/v1/Chassis/1U/Sensors/CPU1Power", result.PowerWatts.DataSourceURI)
	assertEquals(t, "/redfish/v1/Chassis/1U/Sensors/CPUFan1", result.FanSpeedsPercent[0].DataSourceURI)
	assertEquals(t, "CPU #1 Fan Speed", result.FanSpeedsPercent[0].DeviceName)

	if *result.PowerWatts.Reading != 12.87 {
		t.Errorf("Unexpected PowerWatts reading: %.2f", *result.PowerWatts.Reading)
	}
}

// hgxGPUEnvironmentMetricsBody is shaped like NVIDIA HGX GPU EnvironmentMetrics on Dell GB200
// firmware 25.07.4001500, which returns EnergykWh as an empty array.
var hgxGPUEnvironmentMetricsBody = `{
	"@odata.type": "#EnvironmentMetrics.v1_3_0.EnvironmentMetrics",
	"@odata.id": "/redfish/v1/Chassis/HGX_GPU_0/EnvironmentMetrics",
	"Id": "EnvironmentMetrics",
	"Name": "GPU Environment Metrics",
	"EnergykWh": [],
	"PowerLimitWatts": {
	  "AllowableMax": 1200,
	  "DefaultSetPoint": 1200
	}
  }`

// TestEnvironmentMetricsEmptyEnergykWh tests that an empty EnergykWh array does not fail the resource.
func TestEnvironmentMetricsEmptyEnergykWh(t *testing.T) {
	var result EnvironmentMetrics
	if err := json.Unmarshal([]byte(hgxGPUEnvironmentMetricsBody), &result); err != nil {
		t.Fatalf("Error decoding JSON: %s", err)
	}

	if result.PowerLimitWatts.AllowableMax != 1200 {
		t.Errorf("Unexpected PowerLimitWatts AllowableMax: %.0f", result.PowerLimitWatts.AllowableMax)
	}
	if result.PowerLimitWatts.DefaultSetPoint != 1200 {
		t.Errorf("Unexpected PowerLimitWatts DefaultSetPoint: %.0f", result.PowerLimitWatts.DefaultSetPoint)
	}
	if result.EnergykWh.Reading != nil {
		t.Errorf("Expected nil EnergykWh reading, got %.2f", *result.EnergykWh.Reading)
	}
}

// TestEnvironmentMetricsEnergykWhShapes tests which EnergykWh shapes decode and which still fail.
func TestEnvironmentMetricsEnergykWhShapes(t *testing.T) {
	reading := 4.2
	tests := []struct {
		name        string
		energykWh   string
		wantReading *float64
		wantErr     bool
	}{
		{name: "object", energykWh: `{"Reading": 4.2}`, wantReading: &reading},
		{name: "null", energykWh: `null`},
		{name: "empty array", energykWh: `[]`},
		{name: "non-empty array", energykWh: `[{"Reading": 4.2}]`, wantErr: true},
		{name: "string", energykWh: `"4.2"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result EnvironmentMetrics
			err := json.Unmarshal([]byte(`{"Id": "Metrics1", "EnergykWh": `+tt.energykWh+`}`), &result)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected decode error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Error decoding JSON: %s", err)
			}
			if !reflect.DeepEqual(result.EnergykWh.Reading, tt.wantReading) {
				t.Errorf("Unexpected EnergykWh reading: got %v, want %v", result.EnergykWh.Reading, tt.wantReading)
			}
		})
	}
}
