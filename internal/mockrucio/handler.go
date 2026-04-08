package mockrucio

import (
	"encoding/json"
	"net/http"
	"strings"
)

type replicaResponse struct {
	RSE string `json:"rse"`
}

// NewHandler returns an http.Handler that implements the Rucio replica endpoint:
//
//	GET /dids/{scope}/{name}/replicas → []{"rse": "<RSE>"}
//
// Only DIDs present in SeedData are served; all others return 404.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/dids/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// URL shape: /dids/{scope}/{name}/replicas
		// After stripping the prefix, we expect: {scope}/{name}/replicas
		trimmed := strings.TrimPrefix(r.URL.Path, "/dids/")
		if !strings.HasSuffix(trimmed, "/replicas") {
			http.NotFound(w, r)
			return
		}
		didPath := strings.TrimSuffix(trimmed, "/replicas")

		// Convert first "/" to ":" to form the DID scope:name
		slashIdx := strings.Index(didPath, "/")
		if slashIdx < 0 {
			http.NotFound(w, r)
			return
		}
		did := didPath[:slashIdx] + ":" + didPath[slashIdx+1:]

		rses, ok := SeedData[did]
		if !ok {
			http.NotFound(w, r)
			return
		}

		resp := make([]replicaResponse, len(rses))
		for i, rse := range rses {
			resp[i] = replicaResponse{RSE: rse}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})
	return mux
}
