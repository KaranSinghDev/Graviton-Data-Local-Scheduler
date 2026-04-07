package storage

import "testing"

func TestRseToSiteLabel(t *testing.T) {
	cases := []struct {
		rse  string
		want string
	}{
		{"CERN-PROD_DATADISK", "cern-prod"},
		{"BNL-OSG2_DATADISK", "bnl-osg2"},
		{"IN2P3-CC_DATADISK", "in2p3-cc"},
		{"TRIUMF-LCG2_DATADISK", "triumf-lcg2"},
		{"FNAL_DISK", "fnal"},
		{"NOUNDERSCORERSE", "nounderscorerse"},
	}
	for _, c := range cases {
		got := rseToSiteLabel(c.rse)
		if got != c.want {
			t.Errorf("rseToSiteLabel(%q) = %q, want %q", c.rse, got, c.want)
		}
	}
}
