//
// SPDX-License-Identifier: BSD-3-Clause
//

package nvidia

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/coreweave/gofish/common"
)

const modeBody = `{
  "@odata.id": "/redfish/v1/Systems/Bluefield/Oem/Nvidia",
  "@odata.type": "#NvidiaMode.v1_0_0.Mode",
  "Id": "Nvidia",
  "Mode": "NicMode",
  "Actions": {
    "#Mode.Set": {
      "target": "/redfish/v1/Systems/Bluefield/Oem/Nvidia/Actions/AdvertisedMode.Set",
      "@Redfish.ActionInfo": "/redfish/v1/Systems/Bluefield/Oem/Nvidia/ModeSetActionInfo"
    }
  }
}`

func TestModeUnmarshal(t *testing.T) {
	var mode Mode
	if err := json.Unmarshal([]byte(modeBody), &mode); err != nil {
		t.Fatal(err)
	}

	if mode.Mode != NICMode {
		t.Fatalf("Mode = %q, want %q", mode.Mode, NICMode)
	}
	if mode.setTarget != "/redfish/v1/Systems/Bluefield/Oem/Nvidia/Actions/AdvertisedMode.Set" {
		t.Fatalf("setTarget = %q", mode.setTarget)
	}
}

func TestModeSetMode(t *testing.T) {
	client := &common.TestClient{
		CustomReturnForActions: map[string][]interface{}{
			http.MethodPost: {
				&http.Response{StatusCode: http.StatusOK, Body: http.NoBody},
			},
		},
	}
	mode := &Mode{setTarget: "/redfish/v1/Systems/Bluefield/Oem/Nvidia/Actions/AdvertisedMode.Set"}
	mode.SetClient(client)

	if err := mode.SetMode(DPUMode); err != nil {
		t.Fatal(err)
	}

	calls := client.CapturedCalls()
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	if calls[0].Action != http.MethodPost {
		t.Fatalf("Action = %q, want %q", calls[0].Action, http.MethodPost)
	}
	if calls[0].URL != "/redfish/v1/Systems/Bluefield/Oem/Nvidia/Actions/AdvertisedMode.Set" {
		t.Fatalf("URL = %q", calls[0].URL)
	}
	if !strings.Contains(calls[0].Payload, "Mode:DpuMode") {
		t.Fatalf("Payload = %q", calls[0].Payload)
	}
}

func TestModeSetModeRequiresActionTarget(t *testing.T) {
	mode := &Mode{}

	err := mode.SetMode(DPUMode)
	if err == nil {
		t.Fatal("SetMode succeeded without an action target")
	}
	if err.Error() != "Mode.Set is not supported by this system" {
		t.Fatalf("error = %q", err)
	}
}
