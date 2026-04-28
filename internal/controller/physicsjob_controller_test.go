package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hepv1alpha1 "github.com/KaranSinghDev/data-gravity-operator/api/v1alpha1"
	"github.com/KaranSinghDev/data-gravity-operator/internal/storage"
)

// mockStorage is a test double for storage.StorageTopologyClient.
type mockStorage struct {
	replicas map[string][]storage.ReplicaInfo
	err      error
}

func (m *mockStorage) Resolve(_ context.Context, did string) ([]storage.ReplicaInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.replicas[did], nil
}

// cleanupResources does best-effort deletion; it strips Job finalizers so
// envtest's API server can remove them (envtest has no job controller).
func cleanupResources(ctx context.Context, nn types.NamespacedName) {
	job := &batchv1.Job{}
	if err := k8sClient.Get(ctx, nn, job); err == nil {
		job.Finalizers = nil
		_ = k8sClient.Update(ctx, job)
		_ = k8sClient.Delete(ctx, job)
	}
	pj := &hepv1alpha1.PhysicsJob{}
	if err := k8sClient.Get(ctx, nn, pj); err == nil {
		_ = k8sClient.Delete(ctx, pj)
	}
}

const testDID = "data23_13p6TeV:DAOD_PHYS.123456"
const testImage = "gitlab-registry.cern.ch/atlas/athena:latest"

var _ = Describe("PhysicsJob Controller", func() {
	ctx := context.Background()

	// ── DataLocal ────────────────────────────────────────────────────────────
	Context("DataLocal scheduling — known DID", func() {
		nn := types.NamespacedName{Name: "pj-datalocal", Namespace: "default"}

		BeforeEach(func() {
			if apierrors.IsNotFound(k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})) {
				Expect(k8sClient.Create(ctx, &hepv1alpha1.PhysicsJob{
					ObjectMeta: metav1.ObjectMeta{Name: nn.Name, Namespace: nn.Namespace},
					Spec: hepv1alpha1.PhysicsJobSpec{
						Dataset:          testDID,
						Image:            testImage,
						SchedulingPolicy: hepv1alpha1.DataLocal,
					},
				})).To(Succeed())
			}
		})
		AfterEach(func() { cleanupResources(ctx, nn) })

		It("transitions to Scheduled and creates an owned Job with NodeAffinity", func() {
			r := &PhysicsJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Storage: &mockStorage{replicas: map[string][]storage.ReplicaInfo{
					testDID: {{RSE: "CERN-PROD_DATADISK", SiteLabel: "cern-prod", DatasetSizeBytes: 2_500_000_000_000}},
				}},
			}

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			By("PhysicsJob should be Scheduled")
			pj := &hepv1alpha1.PhysicsJob{}
			Expect(k8sClient.Get(ctx, nn, pj)).To(Succeed())
			Expect(pj.Status.Phase).To(Equal(hepv1alpha1.PhaseScheduled))
			Expect(pj.Status.ResolvedRSE).To(Equal("CERN-PROD_DATADISK"))
			Expect(pj.Status.JobRef).To(Equal(nn.Name))
			Expect(pj.Status.BytesTransferAvoided).To(Equal(int64(2_500_000_000_000)))

			By("owned Job should carry the right image")
			job := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, nn, job)).To(Succeed())
			Expect(job.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))

			By("NodeAffinity should pin to cern-prod")
			aff := job.Spec.Template.Spec.Affinity
			Expect(aff).NotTo(BeNil())
			terms := aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms[0].MatchExpressions[0].Values).To(ContainElement("cern-prod"))
		})
	})

	// ── AnyAvailable ─────────────────────────────────────────────────────────
	Context("AnyAvailable scheduling — affinity must be absent", func() {
		nn := types.NamespacedName{Name: "pj-anyavail", Namespace: "default"}

		BeforeEach(func() {
			if apierrors.IsNotFound(k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})) {
				Expect(k8sClient.Create(ctx, &hepv1alpha1.PhysicsJob{
					ObjectMeta: metav1.ObjectMeta{Name: nn.Name, Namespace: nn.Namespace},
					Spec: hepv1alpha1.PhysicsJobSpec{
						Dataset:          testDID,
						Image:            testImage,
						SchedulingPolicy: hepv1alpha1.AnyAvailable,
					},
				})).To(Succeed())
			}
		})
		AfterEach(func() { cleanupResources(ctx, nn) })

		It("creates a Job with no NodeAffinity", func() {
			r := &PhysicsJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Storage: &mockStorage{replicas: map[string][]storage.ReplicaInfo{
					testDID: {{RSE: "CERN-PROD_DATADISK", SiteLabel: "cern-prod"}},
				}},
			}

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			job := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, nn, job)).To(Succeed())
			Expect(job.Spec.Template.Spec.Affinity).To(BeNil())
		})
	})

	// ── Unknown DID ──────────────────────────────────────────────────────────
	Context("unknown DID — should fail", func() {
		nn := types.NamespacedName{Name: "pj-unknown", Namespace: "default"}

		BeforeEach(func() {
			if apierrors.IsNotFound(k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})) {
				Expect(k8sClient.Create(ctx, &hepv1alpha1.PhysicsJob{
					ObjectMeta: metav1.ObjectMeta{Name: nn.Name, Namespace: nn.Namespace},
					Spec: hepv1alpha1.PhysicsJobSpec{
						Dataset: "unknown:dataset.000",
						Image:   testImage,
					},
				})).To(Succeed())
			}
		})
		AfterEach(func() { cleanupResources(ctx, nn) })

		It("transitions to Failed with RSENotFound condition", func() {
			r := &PhysicsJobReconciler{
				Client:  k8sClient,
				Scheme:  k8sClient.Scheme(),
				Storage: &mockStorage{replicas: map[string][]storage.ReplicaInfo{}},
			}

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			pj := &hepv1alpha1.PhysicsJob{}
			Expect(k8sClient.Get(ctx, nn, pj)).To(Succeed())
			Expect(pj.Status.Phase).To(Equal(hepv1alpha1.PhaseFailed))

			var found bool
			for _, c := range pj.Status.Conditions {
				if c.Reason == "RSENotFound" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "expected RSENotFound condition")
		})
	})
})
