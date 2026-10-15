//
// SPDX-License-Identifier: BSD-3-Clause
//

package gofish

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coreweave/gofish/common"
)

const (
	errMsg = `{
      "code": "Base.1.0.GeneralError",
      "message": "A general error has occurred. See ExtendedInfo for more information.",
      "@Message.ExtendedInfo": [
          {
              "MessageId": "Base.1.0.PropertyValueNotInList",
              "Message": "The value Red for the property IndicatorLED is not in the list of acceptable values",
              "MessageArgs": [
                  "RED",
                  "IndicatorLED"
              ],
              "Severity": "Warning",
              "Resolution": "Remove the property from the request body and resubmit the request if the operation failed"
          },
          {
              "MessageId": "Base.1.0.PropertyNotWriteable",
              "Message": "The property SKU is a read only property and cannot be assigned a value",
              "MessageArgs": [
                  "SKU"
              ],
              "Severity": "Warning",
              "Resolution": "Remove the property from the request body and resubmit the request if the operation failed"
          }
      ]
  }`
	expectErrorStatus         = `{"error": ` + errMsg + "}"
	nonErrorStructErrorStatus = "Internal Server Error"
)

func testError(code int, t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(expectErrorStatus)) //nolint
	}))
	defer ts.Close()

	_, err := Connect(ClientConfig{Endpoint: ts.URL, HTTPClient: ts.Client()})
	if err == nil {
		t.Error("Update call should fail")
	}
	errStruct, ok := err.(*common.Error)
	if !ok {
		t.Errorf("%d should return known error type: %v", code, err)
	}
	if errStruct.HTTPReturnedStatusCode != code {
		t.Errorf("The error code is different from %d", code)
	}
	errBody, err := json.MarshalIndent(errStruct, "  ", "    ")
	if err != nil {
		t.Errorf("Marshall error %v got: %s", errStruct, err)
	}
	if errMsg != string(errBody) {
		t.Errorf("Expect:\n%s\nGot:\n%s", errMsg, string(errBody))
	}
}

// TestError400 tests the parsing of error reply.
func TestError400(t *testing.T) {
	testError(400, t)
}

// TestError404 tests the parsing of error reply.
func TestError404(t *testing.T) {
	testError(404, t)
}

// TestErrorOther tests failures that do not return an Error struct
func TestErrorOther(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(nonErrorStructErrorStatus)) //nolint
	}))
	defer ts.Close()

	_, err := Connect(ClientConfig{Endpoint: ts.URL, HTTPClient: ts.Client()})
	if err == nil {
		t.Error("connect should fail")
	}
	errStruct, ok := err.(*common.Error)
	if !ok {
		t.Errorf("call should return known error type: %v", err)
	}
	if errStruct.HTTPReturnedStatusCode != 500 {
		t.Errorf("The error code is different from 500")
	}
	if err.Error() != "500: Internal Server Error" {
		t.Errorf("Expected '500: %s', got '%s'", nonErrorStructErrorStatus, err.Error())
	}
}

// TestConnectContextTimeout
func TestConnectContextTimeout(t *testing.T) {
	// ctx will timeout very quickly
	ctx, cancel := context.WithTimeout(
		context.Background(),
		1*time.Microsecond)
	defer cancel()

	_, err := ConnectContext(
		ctx,
		ClientConfig{
			Endpoint: "https://testContextTimeout.com",
		})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("Context should timeout")
	}
}

func TestServiceGetter(t *testing.T) {
	type serviceGetter interface {
		GetService() *Service
	}

	var sg serviceGetter = &APIClient{}
	if sg.GetService() != nil {
		t.Errorf("Empty client should return a nil service")
	}
}

// TestConnectContextCancel
func TestConnectContextCancel(t *testing.T) {
	// ctx will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ConnectContext(
		ctx,
		ClientConfig{
			Endpoint: "https://testContextCancel.com",
		})

	if !errors.Is(err, context.Canceled) {
		t.Error("Context should be cancelled")
	}
}

// TestConnectDefaultContextTimeout
func TestConnectDefaultContextTimeout(t *testing.T) {
	// ctx will timeout very quickly
	ctx, cancel := context.WithTimeout(
		context.Background(),
		1*time.Microsecond)
	defer cancel()

	_, err := ConnectDefaultContext(
		ctx,
		"https://testContextTimeout.com",
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("Context should timeout")
	}
}

// TestConnectDefaultContextCancel
func TestConnectDefaultContextCancel(t *testing.T) {
	// ctx will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ConnectDefaultContext(
		ctx,
		"https://testContextCancel.com",
	)

	if !errors.Is(err, context.Canceled) {
		t.Error("Context should be cancelled")
	}
}

func TestClientRunRawRequestNoURL(t *testing.T) {
	client := APIClient{sem: make(chan bool, 1)}

	_, err := client.runRawRequest("", "", nil, "") //nolint:bodyclose
	if err == nil {
		t.Error("Request without relative path should have failed")
	}

	if err.Error() != "unable to execute request, no target provided" {
		t.Errorf("Unexpected error response: %s", err.Error())
	}
}

func genTLSCert(t *testing.T) tls.Certificate {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	template := x509.Certificate{
		// serial doesn't matter to us here and old golang needs it
		// newer golang versions can autogen
		SerialNumber:          big.NewInt(100),
		Subject:               pkix.Name{Organization: []string{"test"}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
	}
}

func TestCertHashMonitoring(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`)) //nolint
	}))
	defer ts.Close()

	client, err := Connect(ClientConfig{Endpoint: ts.URL, HTTPClient: ts.Client(), EnableTLSHashMonitoring: true})
	if err != nil {
		t.Error("connect should not fail")
	}

	if client.CertHash == "" {
		t.Error("cert hash should be present")
	}

	expectedHash := hex.EncodeToString(ts.Certificate().Signature)
	if expectedHash != client.CertHash {
		t.Error("cert hash mismatch")
	}

	svc, err := ServiceRoot(client)
	if err != nil || svc == nil {
		t.Error("failed to get service root")
	}

	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

	ts.TLS.Certificates = []tls.Certificate{genTLSCert(t)}

	for i := 0; i < 5; i++ {
		_, err = ServiceRoot(client)
		if !errors.Is(err, ErrClientCertChanged) {
			t.Error("expecting client cert changed err")
		}
	}
}

// newServiceRootTLSServer starts an HTTPS server with a self-signed
// certificate that answers every request with an empty service root.
func newServiceRootTLSServer(t *testing.T) *httptest.Server {
	t.Helper()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{}`)) //nolint
	}))
	t.Cleanup(ts.Close)
	return ts
}

// setDefaultTransportTLSConfig puts cfg on http.DefaultTransport until the
// test ends. A process has a non-nil config there after its first HTTPS
// request through the default transport, or when it installs its own. The
// config is process-wide, so tests that call this must not run in parallel.
func setDefaultTransportTLSConfig(t *testing.T, cfg *tls.Config) {
	t.Helper()
	defaultTransport := http.DefaultTransport.(*http.Transport)
	original := defaultTransport.TLSClientConfig
	defaultTransport.TLSClientConfig = cfg
	t.Cleanup(func() {
		defaultTransport.CloseIdleConnections()
		defaultTransport.TLSClientConfig = original
	})
}

// TestConnectInsecureLeavesDefaultTransportTLSConfig checks that an Insecure
// connect skips certificate checks only on the transport gofish builds, and
// that other requests through http.DefaultTransport still verify certificates.
func TestConnectInsecureLeavesDefaultTransportTLSConfig(t *testing.T) {
	ts := newServiceRootTLSServer(t)
	shared := &tls.Config{MinVersion: tls.VersionTLS12}
	setDefaultTransportTLSConfig(t, shared)

	// No HTTPClient, so gofish builds its transport from http.DefaultTransport.
	if _, err := Connect(ClientConfig{Endpoint: ts.URL, Insecure: true}); err != nil {
		t.Fatalf("insecure connect to a self-signed server failed: %v", err)
	}

	if shared.InsecureSkipVerify {
		t.Error("connect set InsecureSkipVerify on http.DefaultTransport's tls.Config")
	}
	if len(shared.NextProtos) != 0 {
		t.Errorf("connect changed NextProtos on http.DefaultTransport's tls.Config: %v", shared.NextProtos)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("http.DefaultClient accepted a self-signed certificate after an insecure connect")
	}
	var certErr *tls.CertificateVerificationError
	if !errors.As(err, &certErr) {
		t.Errorf("expected a certificate verification error, got: %v", err)
	}
}

// TestConnectInsecureConcurrent runs Insecure connects in parallel while
// http.DefaultTransport holds a tls.Config. Each connect needs its own copy of
// that config, or go test -race reports the writes to the shared one.
func TestConnectInsecureConcurrent(t *testing.T) {
	ts := newServiceRootTLSServer(t)
	setDefaultTransportTLSConfig(t, &tls.Config{MinVersion: tls.VersionTLS12})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Connect(ClientConfig{Endpoint: ts.URL, Insecure: true}); err != nil {
				t.Errorf("concurrent insecure connect failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

// TestConnectNegotiatesHTTP2 checks that the transport gofish builds still
// offers HTTP/2. net/http turns HTTP/2 off on a transport with its own
// TLSClientConfig or DialContext unless ForceAttemptHTTP2 is set.
func TestConnectNegotiatesHTTP2(t *testing.T) {
	var proto atomic.Value
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto.Store(r.Proto)
		w.Write([]byte(`{}`)) //nolint
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	t.Cleanup(ts.Close)

	if _, err := Connect(ClientConfig{Endpoint: ts.URL, Insecure: true}); err != nil {
		t.Fatalf("insecure connect failed: %v", err)
	}
	if got := proto.Load(); got != "HTTP/2.0" {
		t.Errorf("expected the connect to negotiate HTTP/2.0, got %v", got)
	}
}
