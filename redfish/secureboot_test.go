//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/coreweave/gofish/common"
)

type secureBootUpdateClient struct {
	common.TestClient
	patchBody string
}

func (c *secureBootUpdateClient) PatchWithHeadersWithContext(ctx context.Context, uri string, payload interface{}, headers map[string]string) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	c.patchBody = string(body)
	return c.TestClient.PatchWithHeadersWithContext(ctx, uri, payload, headers)
}

func TestSecureBootSettingsTargetPreservesDirectUpdate(t *testing.T) {
	for _, advertised := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			for _, changed := range []bool{false, true} {
				t.Run(fmt.Sprintf("settings=%t/enabled=%t/changed=%t", advertised, enabled, changed), func(t *testing.T) {
					const active = "/redfish/v1/SecureBoot"
					target := ""
					settings := ""
					if advertised {
						target = active + "/Settings"
						settings = fmt.Sprintf(`,"@Redfish.Settings":{"SettingsObject":{"@odata.id":%q}}`, target)
					}
					var resource SecureBoot
					if err := json.Unmarshal([]byte(fmt.Sprintf(`{"@odata.id":%q,"@odata.etag":"active","SecureBootEnable":%t%s}`, active, enabled, settings)), &resource); err != nil {
						t.Fatal(err)
					}
					client := &secureBootUpdateClient{}
					if got := resource.Settings.SettingsObject.String(); got != target {
						t.Fatalf("Settings.SettingsObject = %q, want %q", got, target)
					}
					resource.SetClient(client)
					if changed {
						resource.SecureBootEnable = !enabled
					}
					if err := resource.Update(); err != nil {
						t.Fatal(err)
					}
					calls := client.CapturedCalls()
					if !changed {
						if len(calls) != 0 {
							t.Fatalf("unchanged resource made requests: %+v", calls)
						}
						return
					}
					if len(calls) != 1 {
						t.Fatalf("requests: %+v", calls)
					}
					patch := calls[len(calls)-1]
					if patch.Action != http.MethodPatch || patch.URL != active || patch.CustomHeaders["If-Match"] != "active" {
						t.Fatalf("PATCH: %+v", patch)
					}
					if want := fmt.Sprintf(`{"SecureBootEnable":%t}`, !enabled); client.patchBody != want {
						t.Fatalf("body = %s, want %s", client.patchBody, want)
					}
					if resource.ODataID != active || resource.ODataEtag != "active" {
						t.Fatal("active resource identity or ETag changed")
					}
				})
			}
		}
	}
}

var secureBootBody = `{
		"@odata.context": "/redfish/v1/$metadata#SecureBoot.SecureBoot",
		"@odata.type": "#SecureBoot.v1_0_5.SecureBoot",
		"@odata.id": "/redfish/v1/SecureBoot",
		"Id": "SecureBoot-1",
		"Name": "SecureBootOne",
		"Description": "SecureBoot One",
		"SecureBootCurrentBoot": "Enabled",
		"SecureBootEnable": true,
		"SecureBootMode": "UserMode",
		"Actions": {
			"#SecureBoot.ResetKeys": {
				"target": "/redfish/v1/SecureBoot/Actions/SecureBoot.ResetKeys"
			}
		}
	}`

// TestSecureBoot tests the parsing of SecureBoot objects.
func TestSecureBoot(t *testing.T) {
	var result SecureBoot
	err := json.NewDecoder(strings.NewReader(secureBootBody)).Decode(&result)

	if err != nil {
		t.Errorf("Error decoding JSON: %s", err)
	}

	if result.ID != "SecureBoot-1" {
		t.Errorf("Received invalid ID: %s", result.ID)
	}

	if result.Name != "SecureBootOne" {
		t.Errorf("Received invalid name: %s", result.Name)
	}

	if result.SecureBootCurrentBoot != EnabledSecureBootCurrentBootType {
		t.Errorf("Invalid SecureBootCurrentBoot: %s", result.SecureBootCurrentBoot)
	}

	if !result.SecureBootEnable {
		t.Error("SecureBootEnable should be true")
	}

	if result.SecureBootMode != UserModeSecureBootModeType {
		t.Errorf("Invalid SecureBootMode: %s", result.SecureBootMode)
	}

	if result.Actions.ResetKeys.Target != "/redfish/v1/SecureBoot/Actions/SecureBoot.ResetKeys" {
		t.Errorf("Invalid ResetKeys target: %s", result.Actions.ResetKeys.Target)
	}
}

// TestSecureBootUpdate tests the Update call.
func TestSecureBootUpdate(t *testing.T) {
	var result SecureBoot
	err := json.NewDecoder(strings.NewReader(secureBootBody)).Decode(&result)

	if err != nil {
		t.Errorf("Error decoding JSON: %s", err)
	}

	testClient := &common.TestClient{}
	result.SetClient(testClient)

	result.SecureBootEnable = false
	err = result.Update()

	if err != nil {
		t.Errorf("Error making Update call: %s", err)
	}

	calls := testClient.CapturedCalls()

	if !strings.Contains(calls[0].Payload, "SecureBootEnable:false") {
		t.Errorf("Unexpected SecureBootEnable update payload: %s", calls[0].Payload)
	}
}
