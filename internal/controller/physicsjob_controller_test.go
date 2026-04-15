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

var _ = Describe("PhysicsJob Controller", func() {
	const (
		resourceName = "test-physicsjob"
		testDID      = "data23_13p6TeV:DAOD_PHYS.123456"
		testImage    = "gitlab-registry.cern.ch/atlas/athena:latest"
	)

	ctx := context.Background()
	nn := types.NamespacedName{Name: resourceName, Namespace: "default"}

	AfterEach(func() {
		pj := &hepv1alpha1.PhysicsJob{}
		err := k8sClient.Get(ctx, nn, pj)
		if err == nil {
			Expect(k8sClient.Delete(ctx, pj)).To(Succeed())
		}

		// Clean up the owned Job if it exists.
		job := &batchv1.Job{}
		err = k8sClient.Get(ctx, nn, job)
		if err == nil {
			Expect(k8sClient.Delete(ctx, job)).To(Succeed())
		}
	})

	Context("DataLocal scheduling — known DID", func() {
		BeforeEach(func() {
			By("creating the PhysicsJob")
			pj := &hepv1alpha1.PhysicsJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: hepv1alpha1.PhysicsJobSpec{
					Dataset:          testDID,
					Image:            testImage,
					SchedulingPolicy: hepv1alpha1.DataLocal,
				},
			}
			err := k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})
			if apierrors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, pj)).To(Succeed())
			}
		})

		It("transitions to Scheduled and creates an owned Job", func() {
			mock := &mockStorage{
				replicas: map[string][]storage.ReplicaInfo{
					testDID: {
						{RSE: "CERN-PROD_DATADISK", SiteLabel: "cern-prod"},
					},
				},
			}
			r := &PhysicsJobReconciler{
				Client:  k8sClient,
				Scheme:  k8sClient.Scheme(),
				Storage: mock,
			}

			By("first reconcile — sets Resolving, creates Job, sets Scheduled")
			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			By("checking PhysicsJob phase is Scheduled")
			pj := &hepv1alpha1.PhysicsJob{}
			Expect(k8sClient.Get(ctx, nn, pj)).To(Succeed())
			Expect(pj.Status.Phase).To(Equal(hepv1alpha1.PhaseScheduled))
			Expect(pj.Status.ResolvedRSE).To(Equal("CERN-PROD_DATADISK"))
			Expect(pj.Status.JobRef).To(Equal(resourceName))

			By("checking the owned batch/v1.Job was created")
			job := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, nn, job)).To(Succeed())
			Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(job.Spec.Template.Spec.Containers[0].Image).To(Equal(testImage))

			By("checking NodeAffinity was injected")
			aff := job.Spec.Template.Spec.Affinity
			Expect(aff).NotTo(BeNil())
			terms := aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions[0].Values).To(ContainElement("cern-prod"))
		})
	})

	Context("AnyAvailable scheduling — affinity must be absent", func() {
		BeforeEach(func() {
			pj := &hepv1alpha1.PhysicsJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: hepv1alpha1.PhysicsJobSpec{
					Dataset:          testDID,
					Image:            testImage,
					SchedulingPolicy: hepv1alpha1.AnyAvailable,
				},
			}
			err := k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})
			if apierrors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, pj)).To(Succeed())
			}
		})

		It("creates a Job with no NodeAffinity", func() {
			mock := &mockStorage{
				replicas: map[string][]storage.ReplicaInfo{
					testDID: {{RSE: "CERN-PROD_DATADISK", SiteLabel: "cern-prod"}},
				},
			}
			r := &PhysicsJobReconciler{
				Client:  k8sClient,
				Scheme:  k8sClient.Scheme(),
				Storage: mock,
			}
			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			job := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, nn, job)).To(Succeed())
			Expect(job.Spec.Template.Spec.Affinity).To(BeNil())
		})
	})

	Context("unknown DID — should fail", func() {
		BeforeEach(func() {
			pj := &hepv1alpha1.PhysicsJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: hepv1alpha1.PhysicsJobSpec{
					Dataset: "unknown:dataset.000",
					Image:   testImage,
				},
			}
			err := k8sClient.Get(ctx, nn, &hepv1alpha1.PhysicsJob{})
			if apierrors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, pj)).To(Succeed())
			}
		})

		It("transitions to Failed with RSENotFound condition", func() {
			mock := &mockStorage{replicas: map[string][]storage.ReplicaInfo{}}
			r := &PhysicsJobReconciler{
				Client:  k8sClient,
				Scheme:  k8sClient.Scheme(),
				Storage: mock,
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
