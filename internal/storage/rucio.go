package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// rseToSiteLabel derives the topology.cern.io/site label value from an RSE name.
// Convention: the site name is the lowercase prefix before the first underscore.
// e.g. "CERN-PROD_DATADISK" → "cern-prod", "BNL-OSG2_DATADISK" → "bnl-osg2".
func rseToSiteLabel(rse string) string {
	idx := strings.Index(rse, "_")
	if idx < 0 {
		return strings.ToLower(rse)
	}
	return strings.ToLower(rse[:idx])
}

// RucioClient calls a Rucio-compatible REST endpoint to resolve DIDs to replica lists.
type RucioClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRucioClient constructs a RucioClient targeting baseURL.
func NewRucioClient(baseURL string) *RucioClient {
	return &RucioClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type rucioReplica struct {
	RSE string `json:"rse"`
}

// Resolve implements StorageTopologyClient by calling GET /dids/{scope}/{name}/replicas.
func (c *RucioClient) Resolve(ctx context.Context, did string) ([]ReplicaInfo, error) {
	parts := strings.SplitN(did, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid DID %q: expected scope:name", did)
	}
	scope, name := parts[0], parts[1]

	url := fmt.Sprintf("%s/dids/%s/%s/replicas", c.baseURL, scope, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rucio request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rucio returned HTTP %d", resp.StatusCode)
	}

	var raw []rucioReplica
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode replicas: %w", err)
	}

	out := make([]ReplicaInfo, len(raw))
	for i, r := range raw {
		out[i] = ReplicaInfo{
			RSE:       r.RSE,
			SiteLabel: rseToSiteLabel(r.RSE),
		}
	}
	return out, nil
}
