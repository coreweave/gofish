//
// SPDX-License-Identifier: BSD-3-Clause
//

package common

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// orderedGetClient serves canned GET bodies keyed by URL and delays the gated
// URL until the last collection member has been served, forcing fetch
// completion order to differ from the collection's document order.
type orderedGetClient struct {
	*TestClient
	mu      sync.Mutex
	bodies  map[string]string
	served  []string
	gate    chan struct{}
	gateURL string
	lastURL string
}

func (c *orderedGetClient) GetWithContext(_ context.Context, url string) (*http.Response, error) {
	if url == c.gateURL {
		select {
		case <-c.gate:
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("gate for %s never opened: member fetches are no longer concurrent", url)
		}
	}

	c.mu.Lock()
	body, ok := c.bodies[url]
	c.served = append(c.served, url)
	c.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unexpected GET %s", url)
	}

	if url == c.lastURL {
		close(c.gate)
	}
	return jsonResponse(body), nil
}

func (c *orderedGetClient) Get(url string) (*http.Response, error) {
	return c.GetWithContext(context.Background(), url)
}

// TestGetCollectionObjectsKeepsDocumentOrder pins the returned member order to
// the collection document. The first-listed member completes last, so a result
// in completion order would come back rotated.
func TestGetCollectionObjectsKeepsDocumentOrder(t *testing.T) {
	const colURI = "/redfish/v1/PSUs"
	ids := []string{"2", "10", "0", "1"}

	bodies := map[string]string{}
	members := ""
	for i, id := range ids {
		if i > 0 {
			members += ","
		}
		uri := colURI + "/" + id
		members += fmt.Sprintf(`{"@odata.id":%q}`, uri)
		bodies[uri] = fmt.Sprintf(`{"@odata.id":%q,"Id":%q,"Name":"PSU %s"}`, uri, id, id)
	}
	bodies[colURI] = fmt.Sprintf(`{"Members":[%s],"Members@odata.count":%d}`, members, len(ids))

	client := &orderedGetClient{
		TestClient: &TestClient{},
		bodies:     bodies,
		gate:       make(chan struct{}),
		gateURL:    colURI + "/" + ids[0],
		lastURL:    colURI + "/" + ids[len(ids)-1],
	}

	resources, err := GetCollectionObjectsWithContext[Resource](context.Background(), client, colURI)
	if err != nil {
		t.Fatalf("GetCollectionObjectsWithContext: %v", err)
	}

	var got []string
	for _, r := range resources {
		got = append(got, r.ID)
	}
	if fmt.Sprint(got) != fmt.Sprint(ids) {
		t.Fatalf("member order = %v, want document order %v", got, ids)
	}

	// the gate must have made the first-listed member finish last, or this
	// test could pass in completion order too
	last := client.served[len(client.served)-1]
	if last != client.gateURL {
		t.Fatalf("first-listed member was served %q-last; the gate did not stagger completion", last)
	}
}
