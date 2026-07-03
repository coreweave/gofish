package gofish

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/coreweave/gofish/common"
	"github.com/coreweave/gofish/redfish"
)

type APIClientSerialized struct {
	Endpoint           string
	Service            *Service
	Auth               *redfish.AuthToken
	KeepAlive          bool
	Settings           common.ClientSettings
	CertHashMonitoring bool
	CertHash           string
	SemLen             int
}

func (c *APIClient) Serialize() ([]byte, error) {
	out := APIClientSerialized{
		Endpoint:           c.endpoint,
		Service:            c.Service,
		Auth:               c.auth,
		KeepAlive:          c.keepAlive,
		Settings:           c.Settings,
		CertHashMonitoring: c.certHashMonitoring,
		CertHash:           c.CertHash,
		SemLen:             len(c.sem),
	}
	return json.Marshal(out)
}

func DeserializeClient(ctx context.Context, dumpWriter io.Writer, httpClient *http.Client, cfg []byte) (*APIClient, error) {
	clientData := &APIClientSerialized{}

	err := json.Unmarshal(cfg, clientData)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		ctx:                ctx,
		endpoint:           clientData.Endpoint,
		HTTPClient:         httpClient,
		Service:            clientData.Service,
		auth:               clientData.Auth,
		sem:                make(chan bool, clientData.SemLen),
		dumpWriter:         dumpWriter,
		keepAlive:          clientData.KeepAlive,
		Settings:           clientData.Settings,
		CertHash:           clientData.CertHash,
		certHashMonitoring: clientData.CertHashMonitoring,
	}, nil
}
