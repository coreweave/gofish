//
// SPDX-License-Identifier: BSD-3-Clause
//

package common

import (
	"testing"
)

const (
	envelopeBody = `{"error":{"code":"Base.1.10.GeneralError",` +
		`"message":"A general error has occurred. See ExtendedInfo for more information.",` +
		`"@Message.ExtendedInfo":[{"MessageId":"Base.1.10.AccessUnauthorized",` +
		`"Message":"While attempting to establish a connection, the service denied access.",` +
		`"MessageArgs":["admin"],"Resolution":"Verify credentials and resubmit the request."}]}}`

	barePropertyBody = `{"VerifyRemoteServerCertificate@Message.ExtendedInfo":[{` +
		`"@odata.type":"#Message.v1_1_1.Message",` +
		`"Message":"The property VerifyRemoteServerCertificate is a required property and must be included in the request.",` +
		`"MessageArgs":["VerifyRemoteServerCertificate"],` +
		`"MessageId":"Base.1.18.1.PropertyMissing",` +
		`"Resolution":"Ensure that the property is in the request body and has a valid value."}]}`

	bareUnscopedBody = `{"@Message.ExtendedInfo":[{"MessageId":"Base.1.10.ResourceAtUriUnauthorized",` +
		`"Message":"While accessing the resource, the service denied access."}]}`

	envelopeWithSiblingBody = `{"error":{"code":"Base.1.10.GeneralError","message":"failed",` +
		`"@Message.ExtendedInfo":[{"MessageId":"Base.1.10.GeneralError"}]},` +
		`"Attributes@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyNotWritable",` +
		`"MessageArgs":["Attributes"]}]}`

	envelopeWithNestedBody = `{"error":{"code":"Base.1.10.GeneralError","message":"failed",` +
		`"@Message.ExtendedInfo":[{"MessageId":"Base.1.10.GeneralError"}],` +
		`"Boot@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyValueNotInList",` +
		`"MessageArgs":["Boot"]}]}}`

	plainBody = "unable to execute request, no target provided"
)

func asRedfishError(t *testing.T, err error) *Error {
	t.Helper()
	rfErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("ConstructError returned %T, want *Error", err)
	}
	return rfErr
}

func messageIDs(infos []ErrExtendedInfo) []string {
	ids := make([]string, 0, len(infos))
	for _, info := range infos {
		ids = append(ids, info.MessageID)
	}
	return ids
}

func containsID(infos []ErrExtendedInfo, want string) bool {
	for _, id := range messageIDs(infos) {
		if id == want {
			return true
		}
	}
	return false
}

type constructErrorCase struct {
	name string
	body string
	code int

	wantCode       string
	wantMessage    string
	wantExtended   []string          // MessageIds in ExtendedInfos
	wantProperties map[string]string // annotated property -> expected MessageId
}

var constructErrorCases = []constructErrorCase{
	{
		name:         "envelope populates the unscoped messages",
		body:         envelopeBody,
		code:         401,
		wantCode:     "Base.1.10.GeneralError",
		wantMessage:  "A general error has occurred. See ExtendedInfo for more information.",
		wantExtended: []string{"Base.1.10.AccessUnauthorized"},
	},
	{
		name:           "bare property-scoped body is no longer dropped",
		body:           barePropertyBody,
		code:           400,
		wantMessage:    barePropertyBody,
		wantProperties: map[string]string{"VerifyRemoteServerCertificate": "Base.1.18.1.PropertyMissing"},
	},
	{
		name:         "bare unscoped body is no longer dropped",
		body:         bareUnscopedBody,
		code:         401,
		wantMessage:  bareUnscopedBody,
		wantExtended: []string{"Base.1.10.ResourceAtUriUnauthorized"},
	},
	{
		name:           "annotation beside the error object",
		body:           envelopeWithSiblingBody,
		code:           400,
		wantCode:       "Base.1.10.GeneralError",
		wantMessage:    "failed",
		wantExtended:   []string{"Base.1.10.GeneralError"},
		wantProperties: map[string]string{"Attributes": "Base.1.10.PropertyNotWritable"},
	},
	{
		name:           "annotation inside the error object",
		body:           envelopeWithNestedBody,
		code:           400,
		wantCode:       "Base.1.10.GeneralError",
		wantMessage:    "failed",
		wantExtended:   []string{"Base.1.10.GeneralError"},
		wantProperties: map[string]string{"Boot": "Base.1.10.PropertyValueNotInList"},
	},
	{
		name:        "non-JSON body falls back to the raw message",
		body:        plainBody,
		code:        0,
		wantMessage: plainBody,
	},
	{
		name:        "empty body",
		body:        "",
		code:        500,
		wantMessage: "",
	},
	{
		name:        "malformed JSON does not panic",
		body:        `{"VerifyRemoteServerCertificate@Message.ExtendedInfo":[{`,
		code:        400,
		wantMessage: `{"VerifyRemoteServerCertificate@Message.ExtendedInfo":[{`,
	},
	{
		name: "annotation whose value is not an array is skipped",
		body: `{"Foo@Message.ExtendedInfo":"not an array",` +
			`"Bar@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyMissing"}]}`,
		wantMessage: `{"Foo@Message.ExtendedInfo":"not an array",` +
			`"Bar@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyMissing"}]}`,
		code:           400,
		wantProperties: map[string]string{"Bar": "Base.1.10.PropertyMissing"},
	},
}

func TestConstructError(t *testing.T) {
	for i := range constructErrorCases {
		tt := &constructErrorCases[i]
		t.Run(tt.name, func(t *testing.T) {
			assertConstructError(t, tt)
		})
	}
}

func assertConstructError(t *testing.T, tt *constructErrorCase) {
	t.Helper()
	rfErr := asRedfishError(t, ConstructError(tt.code, []byte(tt.body)))

	if rfErr.HTTPReturnedStatusCode != tt.code {
		t.Errorf("status = %d, want %d", rfErr.HTTPReturnedStatusCode, tt.code)
	}
	if string(rfErr.RawData) != tt.body {
		t.Errorf("RawData = %q, want %q", rfErr.RawData, tt.body)
	}
	if rfErr.Code != tt.wantCode {
		t.Errorf("Code = %q, want %q", rfErr.Code, tt.wantCode)
	}
	if rfErr.Message != tt.wantMessage {
		t.Errorf("Message = %q, want %q", rfErr.Message, tt.wantMessage)
	}

	gotExtended := messageIDs(rfErr.ExtendedInfos)
	if len(gotExtended) != len(tt.wantExtended) {
		t.Errorf("ExtendedInfos = %v, want %v", gotExtended, tt.wantExtended)
	}
	for _, want := range tt.wantExtended {
		if !containsID(rfErr.ExtendedInfos, want) {
			t.Errorf("ExtendedInfos = %v, missing %q", gotExtended, want)
		}
	}

	if len(rfErr.PropertyExtendedInfos) != len(tt.wantProperties) {
		t.Errorf("PropertyExtendedInfos = %v, want %v", rfErr.PropertyExtendedInfos, tt.wantProperties)
	}
	for property, wantID := range tt.wantProperties {
		infos, ok := rfErr.PropertyExtendedInfos[property]
		if !ok {
			t.Errorf("PropertyExtendedInfos missing property %q", property)
			continue
		}
		if !containsID(infos, wantID) {
			t.Errorf("property %q = %v, want %q", property, messageIDs(infos), wantID)
		}
		// Whatever is reachable per-property must also be reachable in the flat view.
		if !containsID(rfErr.AllExtendedInfos(), wantID) {
			t.Errorf("AllExtendedInfos = %v, missing %q", messageIDs(rfErr.AllExtendedInfos()), wantID)
		}
	}
}

func TestConstructErrorPropertyArgs(t *testing.T) {
	rfErr := asRedfishError(t, ConstructError(400, []byte(barePropertyBody)))

	infos := rfErr.PropertyExtendedInfos["VerifyRemoteServerCertificate"]
	if len(infos) != 1 {
		t.Fatalf("got %d messages, want 1", len(infos))
	}
	if got := infos[0].MessageArgs; len(got) != 1 || got[0] != "VerifyRemoteServerCertificate" {
		t.Errorf("MessageArgs = %v, want [VerifyRemoteServerCertificate]", got)
	}
	if infos[0].Resolution == "" {
		t.Error("Resolution was dropped")
	}
}

func TestAllExtendedInfos(t *testing.T) {
	t.Run("returns the same slice when there is nothing scoped", func(t *testing.T) {
		rfErr := asRedfishError(t, ConstructError(401, []byte(envelopeBody)))
		got := rfErr.AllExtendedInfos()
		if len(got) != len(rfErr.ExtendedInfos) {
			t.Fatalf("len = %d, want %d", len(got), len(rfErr.ExtendedInfos))
		}
		// The fast path must not copy: callers hit this on every predicate check.
		if len(got) > 0 && &got[0] != &rfErr.ExtendedInfos[0] {
			t.Error("AllExtendedInfos copied the slice on the unscoped fast path")
		}
	})

	t.Run("merges both scopes", func(t *testing.T) {
		rfErr := asRedfishError(t, ConstructError(400, []byte(envelopeWithSiblingBody)))
		got := rfErr.AllExtendedInfos()
		if len(got) != 2 {
			t.Fatalf("AllExtendedInfos = %v, want 2 entries", messageIDs(got))
		}
		for _, want := range []string{"Base.1.10.GeneralError", "Base.1.10.PropertyNotWritable"} {
			if !containsID(got, want) {
				t.Errorf("AllExtendedInfos = %v, missing %q", messageIDs(got), want)
			}
		}
	})

	t.Run("empty error", func(t *testing.T) {
		var rfErr Error
		if got := rfErr.AllExtendedInfos(); got != nil {
			t.Errorf("AllExtendedInfos = %v, want nil", got)
		}
	})
}

func TestHasPropertyScopedExtendedInfo(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"unscoped annotation only", envelopeBody, false},
		{"bare unscoped annotation", bareUnscopedBody, false},
		{"no annotation at all", `{"error":{"code":"Base.1.10.GeneralError"}}`, false},
		{"not JSON", plainBody, false},
		{"empty", "", false},
		{"property-scoped", barePropertyBody, true},
		{"property-scoped beside the envelope", envelopeWithSiblingBody, true},
		{"property-scoped inside the envelope", envelopeWithNestedBody, true},
		{"unscoped first, property-scoped second", bareUnscopedBody + barePropertyBody, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasPropertyScopedExtendedInfo([]byte(tt.body)); got != tt.want {
				t.Errorf("hasPropertyScopedExtendedInfo = %v, want %v", got, tt.want)
			}
		})
	}
}
