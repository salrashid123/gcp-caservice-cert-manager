package controllers

import (
	"context"

	"fmt"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	caapi "github.com/salrashid123/cert-manager-gcp-privateca/api/v1alpha1"
)

type PrivateCAIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=privateca.google.com,resources=privatecaissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=privateca.google.com,resources=privatecaissuers/status,verbs=get;update;patch

func (r *PrivateCAIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("privatecaissuer", req.NamespacedName)

	iss := caapi.PrivateCAIssuer{}
	log.V(1).Info("PrivateCAIssuerReconciler ", "group", req)

	if err := r.Get(ctx, req.NamespacedName, &iss); err != nil {
		log.Error(err, "failed to retrieve PrivateCAIssuer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(1).Info("PrivateCAIssuerReconciler %v", "group", iss.Name)

	myFinalizerName := "privatecaissuer.finalizers.controllers"

	if iss.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(iss.ObjectMeta.Finalizers, myFinalizerName) {
			log.V(1).Info("Setting Finalizer %v", group, myFinalizerName)
			iss.ObjectMeta.Finalizers = append(iss.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, &iss); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(iss.ObjectMeta.Finalizers, myFinalizerName) {
			iss.ObjectMeta.Finalizers = removeString(iss.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(ctx, &iss); err != nil {
				return ctrl.Result{}, err
			}
			log.V(1).Info("PrivateCAIssuerReconciler Cleanup Completed  %s", group, iss.Name)
		}
		return ctrl.Result{}, nil
	}

	if err := validateStepIssuerSpec(iss.Spec); err != nil {
		err := r.setPrivateCAIssuerStatus(ctx, log, &iss, caapi.PrivateCAIssuerConditionFalse, "BadConfig", "Issuer Configuration invalid: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.setPrivateCAIssuerStatus(ctx, log, &iss, caapi.PrivateCAIssuerConditionTrue, "Verified", "Signing CA verified and ready to issue certificates")
}

func (r *PrivateCAIssuerReconciler) setPrivateCAIssuerStatus(ctx context.Context, log logr.Logger, localCA *caapi.PrivateCAIssuer, status caapi.PrivateCAIssuerConditionStatus, reason, message string, args ...interface{}) error {
	// Format the message and update the localCA variable with the new Condition
	completeMessage := fmt.Sprintf(message, args...)
	r.setLocalCACondition(log, localCA, caapi.PrivateCAIssuerConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == caapi.PrivateCAIssuerConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(localCA, eventType, reason, completeMessage)

	// Actually update the LocalCA resource
	return r.Client.Status().Update(ctx, localCA)
}

func (r *PrivateCAIssuerReconciler) setLocalCACondition(log logr.Logger, localCA *caapi.PrivateCAIssuer, conditionType caapi.PrivateCAIssuerConditionType, status caapi.PrivateCAIssuerConditionStatus, reason, message string) {
	newCondition := caapi.PrivateCAIssuerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	nowTime := metav1.NewTime(r.Clock.Now())
	newCondition.LastTransitionTime = &nowTime

	for idx, cond := range localCA.Status.Conditions {
		if cond.Type != conditionType {
			continue
		}

		if cond.Status == status {
			newCondition.LastTransitionTime = cond.LastTransitionTime
		} else {
			log.Info("found status change for LocalCA condition; setting lastTransitionTime", "condition", conditionType, "old_status", cond.Status, "new_status", status, "time", nowTime.Time)
		}

		localCA.Status.Conditions[idx] = newCondition
		return
	}

	localCA.Status.Conditions = append(localCA.Status.Conditions, newCondition)
	log.Info("setting lastTransitionTime for LocalCA condition", "condition", conditionType, "time", nowTime.Time)
}

func validateStepIssuerSpec(s caapi.PrivateCAIssuerSpec) error {
	switch {
	case s.Project == "":
		return fmt.Errorf("spec.project cannot be empty")
	case s.Location == "":
		return fmt.Errorf("spec.location cannot be empty")
	case s.Issuer == "":
		return fmt.Errorf("spec.issuer cannot be empty")
	default:
		return nil
	}
}

func (r *PrivateCAIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&caapi.PrivateCAIssuer{}).
		Complete(r)
}
