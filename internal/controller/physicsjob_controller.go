package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hepv1alpha1 "github.com/KaranSinghDev/data-gravity-operator/api/v1alpha1"
	"github.com/KaranSinghDev/data-gravity-operator/internal/metrics"
	"github.com/KaranSinghDev/data-gravity-operator/internal/scheduling"
	"github.com/KaranSinghDev/data-gravity-operator/internal/storage"
)

// PhysicsJobReconciler reconciles PhysicsJob objects.
type PhysicsJobReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Storage storage.StorageTopologyClient
}

// +kubebuilder:rbac:groups=hep.cern.local,resources=physicsjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hep.cern.local,resources=physicsjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hep.cern.local,resources=physicsjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

func (r *PhysicsJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()

	result, err := r.reconcile(ctx, req)

	// Always record duration and outcome.
	metrics.ReconcileDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.ReconcileTotal.WithLabelValues("error").Inc()
	} else {
		metrics.ReconcileTotal.WithLabelValues("success").Inc()
	}
	return result, err
}

func (r *PhysicsJobReconciler) reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pj hepv1alpha1.PhysicsJob
	if err := r.Get(ctx, req.NamespacedName, &pj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Terminal states — no further action.
	if pj.Status.Phase == hepv1alpha1.PhaseSucceeded || pj.Status.Phase == hepv1alpha1.PhaseFailed {
		return ctrl.Result{}, nil
	}

	// If the owned Job was already created, sync its state back.
	if pj.Status.JobRef != "" {
		return r.syncFromJob(ctx, &pj)
	}

	// ── Resolving ────────────────────────────────────────────────────────────
	pj.Status.Phase = hepv1alpha1.PhaseResolving
	if err := r.Status().Update(ctx, &pj); err != nil {
		return ctrl.Result{}, fmt.Errorf("set Resolving: %w", err)
	}

	replicas, err := r.Storage.Resolve(ctx, pj.Spec.Dataset)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("resolve %q: %w", pj.Spec.Dataset, err)
	}
	if len(replicas) == 0 {
		metrics.ResolutionFailuresTotal.WithLabelValues("RSENotFound").Inc()
		return r.setFailed(ctx, &pj, "RSENotFound",
			"no RSE holds a replica of dataset "+pj.Spec.Dataset)
	}

	// DataLocal / ClosestSite: first replica = highest-priority site.
	// AnyAvailable: still record the RSE for observability, but skip affinity.
	replica := replicas[0]
	policy := string(pj.Spec.SchedulingPolicy)
	metrics.ResolvedTotal.WithLabelValues(replica.RSE, policy).Inc()

	// ── Scheduling ───────────────────────────────────────────────────────────
	var affinity *corev1.Affinity
	if pj.Spec.SchedulingPolicy != hepv1alpha1.AnyAvailable {
		affinity = scheduling.NodeAffinityForSite(replica.SiteLabel)

		// Record estimated bytes not moved over the WAN.
		if replica.DatasetSizeBytes > 0 {
			metrics.DataTransferAvoidedBytes.WithLabelValues(replica.RSE).Add(
				float64(replica.DatasetSizeBytes),
			)
		}
	}

	job := buildJob(&pj, affinity)
	if err := ctrl.SetControllerReference(&pj, job, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("set owner ref: %w", err)
	}
	if err := r.Create(ctx, job); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("create job: %w", err)
	}

	pj.Status.Phase = hepv1alpha1.PhaseScheduled
	pj.Status.ResolvedRSE = replica.RSE
	pj.Status.JobRef = job.Name
	pj.Status.BytesTransferAvoided = replica.DatasetSizeBytes
	if err := r.Status().Update(ctx, &pj); err != nil {
		return ctrl.Result{}, fmt.Errorf("update Scheduled status: %w", err)
	}

	log.Info("job scheduled", "job", job.Name, "rse", replica.RSE, "site", replica.SiteLabel)
	return ctrl.Result{}, nil
}

// syncFromJob reads the owned batch/v1.Job and mirrors its state onto the PhysicsJob status.
func (r *PhysicsJobReconciler) syncFromJob(ctx context.Context, pj *hepv1alpha1.PhysicsJob) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var job batchv1.Job
	nn := types.NamespacedName{Name: pj.Status.JobRef, Namespace: pj.Namespace}
	if err := r.Get(ctx, nn, &job); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, pj, "JobDeleted",
				"owned Job "+pj.Status.JobRef+" was deleted externally")
		}
		return ctrl.Result{}, err
	}

	newPhase := jobPhase(&job)
	if newPhase == pj.Status.Phase {
		return ctrl.Result{}, nil
	}

	log.Info("syncing phase from job", "job", pj.Status.JobRef, "newPhase", newPhase)
	pj.Status.Phase = newPhase
	if err := r.Status().Update(ctx, pj); err != nil {
		return ctrl.Result{}, fmt.Errorf("sync phase from job: %w", err)
	}
	return ctrl.Result{}, nil
}

// setFailed transitions the PhysicsJob to Failed and records a condition.
func (r *PhysicsJobReconciler) setFailed(ctx context.Context, pj *hepv1alpha1.PhysicsJob, reason, msg string) (ctrl.Result, error) {
	pj.Status.Phase = hepv1alpha1.PhaseFailed
	apimeta.SetStatusCondition(&pj.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            msg,
		ObservedGeneration: pj.Generation,
	})
	if err := r.Status().Update(ctx, pj); err != nil {
		return ctrl.Result{}, fmt.Errorf("set Failed: %w", err)
	}
	return ctrl.Result{}, nil
}

// jobPhase maps batch/v1.Job conditions to a PhysicsJob Phase.
func jobPhase(job *batchv1.Job) hepv1alpha1.Phase {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return hepv1alpha1.PhaseSucceeded
		}
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return hepv1alpha1.PhaseFailed
		}
	}
	if job.Status.Active > 0 {
		return hepv1alpha1.PhaseRunning
	}
	return hepv1alpha1.PhaseScheduled
}

// buildJob constructs the owned batch/v1.Job for a PhysicsJob.
func buildJob(pj *hepv1alpha1.PhysicsJob, affinity *corev1.Affinity) *batchv1.Job {
	backoffLimit := int32(3)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pj.Name,
			Namespace: pj.Namespace,
			Labels:    map[string]string{"app.kubernetes.io/managed-by": "data-gravity-operator"},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Affinity:      affinity,
					Containers: []corev1.Container{
						{
							Name:      "compute",
							Image:     pj.Spec.Image,
							Command:   pj.Spec.Command,
							Resources: pj.Spec.Resources,
						},
					},
				},
			},
		},
	}
}

// SetupWithManager registers the reconciler and watches owned Jobs.
func (r *PhysicsJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hepv1alpha1.PhysicsJob{}).
		Owns(&batchv1.Job{}).
		Named("physicsjob").
		Complete(r)
}
