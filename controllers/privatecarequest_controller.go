package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	privatecav1alpha1 "github.com/salrashid123/cert-manager-gcp-privateca/api/v1alpha1"

	"time"

	privateca "cloud.google.com/go/security/privateca/apiv1alpha1"
	ptypes "github.com/golang/protobuf/ptypes"
	privatecapb "google.golang.org/genproto/googleapis/cloud/security/privateca/v1alpha1"
)

const (
	group = "privateca.google.com"
)

// PrivateCARequestReconciler reconciles a PrivateCARequest object
type PrivateCARequestReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=privateca.google.com,resources=privatecarequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=privateca.google.com,resources=privatecarequests/status,verbs=get;update;patch

func (r *PrivateCARequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certificaterequest", req.NamespacedName)

	log.V(1).Info(" Got Certificate Request %v", group, req)

	// Fetch the CertificateRequest resource being reconciled
	cr := cmapi.CertificateRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, &cr); err != nil {
		//log.Error(err, "failed to retrieve CertificateRequest resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the CertificateRequest's issuerRef and if it does not match the
	// exampleapi group name, log a message at a debug level and stop processing.
	if cr.Spec.IssuerRef.Group != privatecav1alpha1.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", group, cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}
	// also check for Kind, etc
	// if cr.Spec.IssuerRef.Kind != "PrivateCAIssuer" {
	// 	log.V(4).Info("resource does not specify an issuerRef Kind that we are responsible for", group, cr.Spec.IssuerRef.Group)
	// 	return ctrl.Result{}, nil
	// }

	iss := privatecav1alpha1.PrivateCAIssuer{}

	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: cr.Spec.IssuerRef.Name}, &iss); err != nil {
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending,
			"Certificate Issuer not defined: %v", err)
		return ctrl.Result{}, err
	}

	projectID := iss.Spec.Project
	location := iss.Spec.Location
	caName := iss.Spec.Issuer

	c, err := privateca.NewCertificateAuthorityClient(ctx)
	if err != nil {
		log.Error(err, "Unable to create PrivateCA Client")
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Unable to create PrivateCA Client: %v", err)
		return ctrl.Result{}, err
	}

	// set a finalizername to clean up the priaveCA on GCP end before deleting the CRD instance.
	myFinalizerName := "privateca.finalizers.controllers"

	// examine DeletionTimestamp to determine if object is under deletion
	if cr.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		log.V(1).Info("Setting Finalizer %v", group, myFinalizerName)
		if !containsString(cr.ObjectMeta.Finalizers, myFinalizerName) {
			cr.ObjectMeta.Finalizers = append(cr.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &cr); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(cr.ObjectMeta.Finalizers, myFinalizerName) {
			certName := cr.Annotations["cert-manager.io/private-key-secret-name"]
			log.V(1).Info("Finalizer called to Delete %s", group, certName)
			name := fmt.Sprintf("projects/%s/locations/%s/certificateAuthorities/%s/certificates/%s", projectID, location, caName, certName)
			revreq := &privatecapb.RevokeCertificateRequest{
				Name:   name,
				Reason: privatecapb.RevocationReason_CESSATION_OF_OPERATION,
			}
			op, err := c.RevokeCertificate(ctx, revreq)
			if err != nil {
				// TODO catch and handle error
				// "rpc error: code = InvalidArgument desc = Certificate is already revoked"
				// ...however, this error seems to only surface durign LRO...
				return ctrl.Result{}, err
			}

			revresp, err := op.Wait(ctx)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					if e.Message() == "Certificate is already revoked" {
						log.V(1).Info("Certificate Status: %s", group, e.Message())
						return ctrl.Result{}, nil
					}
					// e.Code()   // InvalidArgument
					// e.Message() // Certificate is already revoked
				}
			}
			log.V(1).Info("Certificate Revoked: %s", group, revresp.RevocationDetails.RevocationState)

			cr.ObjectMeta.Finalizers = removeString(cr.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &cr); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	// If the certificate data is already set then we skip this request as it
	// has already been completed in the past.
	if len(cr.Status.Certificate) > 0 {
		log.V(4).Info("existing certificate data found in status, skipping already completed CertificateRequest")
		return ctrl.Result{}, nil
	}

	certName := cr.Annotations["cert-manager.io/private-key-secret-name"]

	parent := fmt.Sprintf("projects/%s/locations/%s/certificateAuthorities/%s", projectID, location, caName)
	creq := &privatecapb.CreateCertificateRequest{
		Parent:        parent,
		CertificateId: certName, //uu.String(),
		Certificate: &privatecapb.Certificate{
			Lifetime: ptypes.DurationProto(200 * time.Hour),
			CertificateConfig: &privatecapb.Certificate_PemCsr{
				PemCsr: string(cr.Spec.CSRPEM),
			},
		},
	}
	op, err := c.CreateCertificate(ctx, creq)
	if err != nil {

		// TODO: catch and handle error:
		// rpc error: code = AlreadyExists desc = Resource 'projects/.../locations/us-central1/certificateAuthorities...certificates/...' already exists	{"certificaterequest": "..."}
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Unable to create CertifiateRequest ---> Certificate alrady exists: %v", err)
				return ctrl.Result{}, err
			}
		}
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Unable to create CertifiateRequest : %v", err)
		return ctrl.Result{}, err
	}

	cresp, err := op.Wait(ctx)
	if err != nil {
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonFailed, "Unable to create CertifiateRequest : %v", err)
		return ctrl.Result{}, err
	}

	// fmt.Printf("%s, %s", resp.GetPemCertificate(), string(privPEM))

	signedPEM := []byte(cresp.GetPemCertificate())
	tlsCertCA := []byte(cresp.GetPemCertificateChain()[0])
	log.V(1).Info("Signed Cert %s", group, cresp.GetPemCertificate())
	log.V(1).Info("Signed CA %v", group, cresp.GetPemCertificateChain()[0])

	// Store the signed certificate data in the status
	cr.Status.Certificate = signedPEM
	// copy the CA data from the CA secret
	cr.Status.CA = tlsCertCA

	// Finally, update the status

	// Fetch the Secret resource containing the CA keypair used for signing
	sec := core.Secret{}

	log.V(1).Info("Creating Secret with Name %v", group, certName)
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: certName}, &sec); err != nil {
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending,
			"Failed to fetch certificate-sample-tls secret resource: %v", err)
		return ctrl.Result{}, err
	}
	sec.Data["tls.crt"] = signedPEM
	sec.Data["ca.crt"] = tlsCertCA

	if err := r.Client.Update(ctx, &sec); err != nil {
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending,
			"Failed to Upddate Secret certificate-sample-tls secret resource: %v", err)
		return ctrl.Result{}, err
	}

	// Finally, update the status
	return ctrl.Result{}, r.setStatus(ctx, log, &cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Successfully issued certificate")
}

func (r *PrivateCARequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

func (r *PrivateCARequestReconciler) setStatus(ctx context.Context, log logr.Logger, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	// Format the message and update the localCA variable with the new Condition
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)
	log.Info(completeMessage)

	// Actually update the LocalCA resource
	// TODO: use the 'status' subresource
	return r.Client.Update(ctx, cr)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
