module github.com/salrashid123/cert-manager-gcp-privateca

go 1.13

require (
	cloud.google.com/go/security/privateca/apiv1alpha1 v0.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/jetstack/cert-manager v0.15.2
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/smallstep/step-issuer v0.2.0
	google.golang.org/api v0.29.0 // indirect
	google.golang.org/genproto/googleapis/cloud/security/privateca/v1alpha1 v0.0.0
	google.golang.org/grpc v1.30.0
	istio.io/pkg v0.0.0-20200717003958-7c88638e196b
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v0.18.4
	k8s.io/utils v0.0.0-20200720150651-0bdb4ca86cbc
	sigs.k8s.io/controller-runtime v0.6.1
)

replace cloud.google.com/go/security/privateca/apiv1alpha1 => ./src/cloud.google.com/go/security/privateca/apiv1alpha1

replace google.golang.org/genproto/googleapis/cloud/security/privateca/v1alpha1 => ./src/google.golang.org/genproto/googleapis/cloud/security/privateca/v1alpha1
