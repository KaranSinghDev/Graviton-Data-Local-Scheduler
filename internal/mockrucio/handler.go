package mockrucio

import (
	"encoding/json"
	"net/http"
	"strings"
)

type replicaResponse struct {
	RSE              string `json:"rse"`
	DatasetSizeBytes int64  `json:"bytes,omitempty"`
}

// NewHandler returns an http.Handler that implements the Rucio replica endpoint:
//
//	GET /dids/{scope}/{name}/replicas → []{"rse": "<RSE>", "bytes": <size>}
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
		trimmed := strings.TrimPrefix(r.URL.Path, "/dids/")
		if !strings.HasSuffix(trimmed, "/replicas") {
			http.NotFound(w, r)
			return
		}
		didPath := strings.TrimSuffix(trimmed, "/replicas")

		slashIdx := strings.Index(didPath, "/")
		if slashIdx < 0 {
			http.NotFound(w, r)
			return
		}
		did := didPath[:slashIdx] + ":" + didPath[slashIdx+1:]

		entry, ok := SeedData[did]
		if !ok {
			http.NotFound(w, r)
			return
		}

		resp := make([]replicaResponse, len(entry.RSEs))
		for i, rse := range entry.RSEs {
			resp[i] = replicaResponse{RSE: rse, DatasetSizeBytes: entry.DatasetSizeBytes}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})
	return mux
}
