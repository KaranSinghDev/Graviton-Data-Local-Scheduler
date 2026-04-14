package scheduling

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNodeAffinityForSite_Structure(t *testing.T) {
	aff := NodeAffinityForSite("cern-prod")
	if aff == nil {
		t.Fatal("expected non-nil Affinity")
	}
	na := aff.NodeAffinity
	if na == nil {
		t.Fatal("expected non-nil NodeAffinity")
	}
	ns := na.RequiredDuringSchedulingIgnoredDuringExecution
	if ns == nil {
		t.Fatal("expected required node selector to be set")
	}
	if len(ns.NodeSelectorTerms) != 1 {
		t.Fatalf("want 1 NodeSelectorTerm, got %d", len(ns.NodeSelectorTerms))
	}
	exprs := ns.NodeSelectorTerms[0].MatchExpressions
	if len(exprs) != 1 {
		t.Fatalf("want 1 MatchExpression, got %d", len(exprs))
	}
	expr := exprs[0]
	if expr.Key != SiteLabelKey {
		t.Errorf("key: want %q, got %q", SiteLabelKey, expr.Key)
	}
	if expr.Operator != corev1.NodeSelectorOpIn {
		t.Errorf("operator: want In, got %v", expr.Operator)
	}
	if len(expr.Values) != 1 || expr.Values[0] != "cern-prod" {
		t.Errorf("values: want [cern-prod], got %v", expr.Values)
	}
}

func TestNodeAffinityForSite_DifferentLabels(t *testing.T) {
	sites := []string{"bnl-osg2", "in2p3-cc", "triumf-lcg2", "fnal"}
	for _, site := range sites {
		aff := NodeAffinityForSite(site)
		terms := aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		got := terms[0].MatchExpressions[0].Values[0]
		if got != site {
			t.Errorf("site %q: affinity value = %q, want %q", site, got, site)
		}
	}
}
