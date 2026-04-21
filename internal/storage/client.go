package storage

import "context"

// ReplicaInfo describes one RSE holding a dataset replica and its Kubernetes
// node topology label value.
type ReplicaInfo struct {
	// RSE is the Rucio Storage Element name, e.g. "CERN-PROD_DATADISK".
	RSE string
	// SiteLabel is the value for the topology.cern.io/site node label,
	// derived from the RSE name by taking the lowercase site prefix.
	SiteLabel string
	// DatasetSizeBytes is the estimated size of the dataset replica in bytes.
	// Used to compute the physjob_data_transfer_avoided_bytes metric.
	DatasetSizeBytes int64
}

// StorageTopologyClient resolves a Rucio dataset DID to the ordered list of
// RSEs that hold replicas. An empty slice (no error) means the DID is unknown.
type StorageTopologyClient interface {
	Resolve(ctx context.Context, did string) ([]ReplicaInfo, error)
}
