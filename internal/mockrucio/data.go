package mockrucio

// SeedData maps Rucio DIDs (scope:name) to an ordered list of RSE names that
// hold replicas. Entries mirror realistic ATLAS and CMS dataset placements
// across WLCG Tier-1 and Tier-2 sites.
var SeedData = map[string][]string{
	// ATLAS Run-3 collision data — primary at CERN, replica at BNL
	"data23_13p6TeV:DAOD_PHYS.123456": {
		"CERN-PROD_DATADISK",
		"BNL-OSG2_DATADISK",
	},
	// ATLAS Monte Carlo — replica chain across three Tier-1 sites
	"mc23_13p6TeV:DAOD_PHYS.654321": {
		"IN2P3-CC_DATADISK",
		"CERN-PROD_DATADISK",
		"SARA_DATADISK",
	},
	// ATLAS PHYSLITE (slim) — Canada and US Tier-2
	"data23_13p6TeV:DAOD_PHYSLITE.789012": {
		"TRIUMF-LCG2_DATADISK",
		"AGLT2_DATADISK",
	},
	// ATLAS high-pileup special run — CERN only
	"data23_13p6TeV:DAOD_HIGG4D2.999001": {
		"CERN-PROD_DATADISK",
	},
	// CMS Run-3 AOD — Fermilab primary, CERN mirror
	"cms_run3:AOD.CMS2023A.001": {
		"FNAL_DISK",
		"CERN-PROD_DATADISK",
	},
	// CMS NanoAOD — spread across three continents
	"cms_run3:NANOAOD.CMS2023B.042": {
		"FNAL_DISK",
		"IN2P3-CC_DATADISK",
		"INFN-T1_DATADISK",
	},
	// CMS MiniAOD for b-physics — DESY and GridKA
	"cms_run3:MINIAOD.BPH2023.007": {
		"DESY-HH_DATADISK",
		"GridKA_DATADISK",
	},
	// LHCb simulation sample — UK Tier-1 only
	"lhcb_sim:DST.BuToKmumu.2023": {
		"RAL-LCG2_DATADISK",
	},
	// LHCb real data — CERN and CNAF
	"lhcb_collision23:FULL.TURBO.000123": {
		"CERN-PROD_DATADISK",
		"INFN-T1_DATADISK",
	},
}
