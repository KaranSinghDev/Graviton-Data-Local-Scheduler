package scheduling

import corev1 "k8s.io/api/core/v1"

const SiteLabelKey = "topology.cern.io/site"

// NodeAffinityForSite returns an Affinity that hard-pins pods to nodes
// carrying the label topology.cern.io/site=siteLabel.
func NodeAffinityForSite(siteLabel string) *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      SiteLabelKey,
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{siteLabel},
							},
						},
					},
				},
			},
		},
	}
}
