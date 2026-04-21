package mockrucio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_KnownDID(t *testing.T) {
	h := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dids/data23_13p6TeV/DAOD_PHYS.123456/replicas", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	var resp []struct{ RSE string `json:"rse"` }
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one RSE")
	}
	if resp[0].RSE != "CERN-PROD_DATADISK" {
		t.Errorf("want CERN-PROD_DATADISK as first RSE, got %q", resp[0].RSE)
	}
}

func TestHandler_UnknownDID(t *testing.T) {
	h := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dids/unknown/does.not.exist/replicas", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := NewHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dids/data23_13p6TeV/DAOD_PHYS.123456/replicas", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestHandler_AllSeedDIDs(t *testing.T) {
	h := NewHandler()
	for did, rses := range SeedData {
		// convert scope:name → scope/name for URL
		var scope, name string
		for i, c := range did {
			if c == ':' {
				scope = did[:i]
				name = did[i+1:]
				break
			}
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dids/"+scope+"/"+name+"/replicas", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("DID %q: want 200, got %d", did, rec.Code)
			continue
		}
		var resp []struct{ RSE string `json:"rse"` }
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Errorf("DID %q: decode: %v", did, err)
			continue
		}
		if len(resp) != len(rses.RSEs) {
			t.Errorf("DID %q: want %d RSEs, got %d", did, len(rses.RSEs), len(resp))
		}
	}
}
