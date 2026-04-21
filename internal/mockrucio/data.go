package mockrucio

// DatasetEntry holds the replica locations and estimated size for one dataset.
type DatasetEntry struct {
	RSEs             []string
	DatasetSizeBytes int64
}

// SeedData maps Rucio DIDs (scope:name) to replica lists with size estimates.
// Sizes are representative of real ATLAS/CMS/LHCb dataset classes (DAOD ≈ 1–5 TB,
// NanoAOD ≈ 50–200 GB, PHYSLITE ≈ 200–800 GB).
var SeedData = map[string]DatasetEntry{
	// ATLAS Run-3 collision data — primary at CERN, replica at BNL
	"data23_13p6TeV:DAOD_PHYS.123456": {
		RSEs:             []string{"CERN-PROD_DATADISK", "BNL-OSG2_DATADISK"},
		DatasetSizeBytes: 2_500_000_000_000, // 2.5 TB
	},
	// ATLAS Monte Carlo — replica chain across three Tier-1 sites
	"mc23_13p6TeV:DAOD_PHYS.654321": {
		RSEs:             []string{"IN2P3-CC_DATADISK", "CERN-PROD_DATADISK", "SARA_DATADISK"},
		DatasetSizeBytes: 1_800_000_000_000, // 1.8 TB
	},
	// ATLAS PHYSLITE (slim format) — Canada and US Tier-2
	"data23_13p6TeV:DAOD_PHYSLITE.789012": {
		RSEs:             []string{"TRIUMF-LCG2_DATADISK", "AGLT2_DATADISK"},
		DatasetSizeBytes: 400_000_000_000, // 400 GB
	},
	// ATLAS high-pileup special run — CERN only
	"data23_13p6TeV:DAOD_HIGG4D2.999001": {
		RSEs:             []string{"CERN-PROD_DATADISK"},
		DatasetSizeBytes: 900_000_000_000, // 900 GB
	},
	// CMS Run-3 AOD — Fermilab primary, CERN mirror
	"cms_run3:AOD.CMS2023A.001": {
		RSEs:             []string{"FNAL_DISK", "CERN-PROD_DATADISK"},
		DatasetSizeBytes: 3_200_000_000_000, // 3.2 TB
	},
	// CMS NanoAOD — spread across three continents
	"cms_run3:NANOAOD.CMS2023B.042": {
		RSEs:             []string{"FNAL_DISK", "IN2P3-CC_DATADISK", "INFN-T1_DATADISK"},
		DatasetSizeBytes: 120_000_000_000, // 120 GB
	},
	// CMS MiniAOD for b-physics — DESY and GridKA
	"cms_run3:MINIAOD.BPH2023.007": {
		RSEs:             []string{"DESY-HH_DATADISK", "GridKA_DATADISK"},
		DatasetSizeBytes: 600_000_000_000, // 600 GB
	},
	// LHCb simulation sample — UK Tier-1 only
	"lhcb_sim:DST.BuToKmumu.2023": {
		RSEs:             []string{"RAL-LCG2_DATADISK"},
		DatasetSizeBytes: 80_000_000_000, // 80 GB
	},
	// LHCb real data — CERN and CNAF
	"lhcb_collision23:FULL.TURBO.000123": {
		RSEs:             []string{"CERN-PROD_DATADISK", "INFN-T1_DATADISK"},
		DatasetSizeBytes: 1_100_000_000_000, // 1.1 TB
	},
}
