
## Cert-Manager and GCP Private Ceertificate Authority

>> Experimental/WIP, DO NOT USE

[Cert-Manager](https://github.com/jetstack/cert-manager) integration with GCP Private CA

* Requires go 1.14, make

* if you run the CRD locally, your `Application Default Credentials` must have access to privateCA to gen certs.  You must also have setup a privateCA already.
In the example below, the private CA is called `prod-root`

```bash
gcloud config set account ..
gcloud config set project ...
gcloud auth application-default login
(or export `export GOOGLE_APPLICATION_CREDENTIALS=path/to/svc_account.son`)
```

1. Run k8s Minikube:

```bash
minikube start
```


2. Install CertManager

```bash
kubectl create namespace cert-manager
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --version v0.15.2 \
  --set installCRDs=true --set extraArgs[0]="--enable-certificate-owner-ref=true"
kubectl get pods --namespace cert-manager
```

2. Run CRD locally

The following will run the CRDs locally on your laptop but connect to the k8s server wherever it is

```bash
make && make install && make run
```


3. Configure `Issuer`  CRD

The Issuer CRD essentially acts as a configMap currently where an admin would place specifications around the CA, `Issuer` any certs 
issued by it would fall under

In the following config, a `PrivateCAIssuer` called `privateca-issuer` where any Certificate request that is issue references back here will use
the `issuer`, `location` and `project` values 

- To customize, Edit `config/samples/privateca_v1alpha1_issuer.yaml`
```yaml
apiVersion: privateca.google.com/v1alpha1
kind: PrivateCAIssuer
metadata:
  name: privateca-issuer
spec:
  issuer: prod-root
  location: us-central1
  project: mineral-minutia-820
```

As mentioned, the CA itself on GCP is called `prod-root`

and apply

```bash
kubectl apply -f config/samples/privateca_v1alpha1_issuer.yaml

$ kubectl get crd
NAME                                     CREATED AT
certificaterequests.cert-manager.io      2020-08-18T18:08:12Z
certificates.cert-manager.io             2020-08-18T18:08:12Z
challenges.acme.cert-manager.io          2020-08-18T18:08:12Z
clusterissuers.cert-manager.io           2020-08-18T18:08:12Z
issuers.cert-manager.io                  2020-08-18T18:08:12Z
orders.acme.cert-manager.io              2020-08-18T18:08:12Z
privatecaissuers.privateca.google.com    2020-08-20T18:24:12Z
privatecarequests.privateca.google.com   2020-08-20T18:24:12Z

```

Check that the CA issuer is ready

```yaml
$ kubectl describe PrivateCAIssuer privateca-issuer

    NAME                                     CREATED AT
    certificaterequests.cert-manager.io      2020-08-20T18:34:17Z
    certificates.cert-manager.io             2020-08-20T18:34:17Z
    challenges.acme.cert-manager.io          2020-08-20T18:34:17Z
    clusterissuers.cert-manager.io           2020-08-20T18:34:17Z
    issuers.cert-manager.io                  2020-08-20T18:34:17Z
    orders.acme.cert-manager.io              2020-08-20T18:34:17Z
    privatecaissuers.privateca.google.com    2020-08-20T18:35:57Z
    privatecarequests.privateca.google.com   2020-08-20T18:35:57Z

```

5. Create k8s certificate

Edit `config/samples/privateca_v1alpha1_certificaterequest.yaml` and specify a _unique_ certificate name.

The the `secretName: certificate-sample-tls-sscc12` value will become the cert that gets issued in privateCA

```bash
kubectl apply -f config/samples/privateca_v1alpha1_privatecarequest.yaml 

kubectl describe certificate certificate-sample
 
kubectl get secret/certificate-sample-tls-sscc12  -oyaml
```

Then look for the `Events` section in the describe, the final "Updated Secrets" message means the cert got generated

```yaml
$ kubectl describe certificate certificate-sample
      Name:         certificate-sample
      Namespace:    default
      Labels:       app.kubernetes.io/instance=gcp-cert-manager-operator
                    app.kubernetes.io/managed-by=gcp-cert-manager-operator
                    app.kubernetes.io/name=cert-manager
      Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                      {"apiVersion":"cert-manager.io/v1alpha2","kind":"Certificate","metadata":{"annotations":{},"labels":{"app.kubernetes.io/instance":"gcp-cer...
      API Version:  cert-manager.io/v1alpha3
      Kind:         Certificate
      Metadata:
        Creation Timestamp:  2020-08-20T18:37:31Z
        Generation:          1
        Resource Version:    1257
        Self Link:           /apis/cert-manager.io/v1alpha3/namespaces/default/certificates/certificate-sample
        UID:                 1a9ebbea-272a-420d-92b6-50606e0ab78c
      Spec:
        Dns Names:
          test.example.com
        Issuer Ref:
          Group:      privateca.google.com
          Kind:       PrivateCAIssuer
          Name:       privateca-issuer
        Secret Name:  certificate-sample-tls-sscc12
      Status:
        Conditions:
          Last Transition Time:  2020-08-20T18:37:33Z
          Message:               Certificate is up to date and has not expired
          Reason:                Ready
          Status:                True
          Type:                  Ready
        Not After:               2020-08-29T02:37:32Z
      Events:
        Type    Reason        Age   From          Message
        ----    ------        ----  ----          -------
        Normal  GeneratedKey  43s   cert-manager  Generated a new private key
        Normal  Requested     43s   cert-manager  Created new CertificateRequest resource "certificate-sample-1188600488"
        Normal  UpdateMeta    41s   cert-manager  Updated metadata on Secret resource             <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Manual Update of Secret{}
```


Though the certificate is issued

```
$ kubectl get secret/certificate-sample-tls-sscc12  -oyaml
apiVersion: v1
data:
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUIwekNDQVhtZ0F3SUJBZ0lWQUpBUDEyQWltOUg4QkEzZWRLZnFDYm1nZ24yTk1Bb0dDQ3FHU000OUJBTUMKTURVeEVUQVBCZ05WQkFvVENGUmxjM1FnVEV4RE1TQXdIZ1lEVlFRREV4ZFVaWE4wSUZCeWIyUjFZM1JwYjI0ZwpVbTl2ZENCRFFUQWVGdzB5TURBME1qSXdNREU0TlRWYUZ3MHpNREEwTWpJeE1ESTJNelphTURVeEVUQVBCZ05WCkJBb1RDRlJsYzNRZ1RFeERNU0F3SGdZRFZRUURFeGRVWlhOMElGQnliMlIxWTNScGIyNGdVbTl2ZENCRFFUQloKTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCQURrMFdaWTMwWUk5eUhwL0hwVjZIeGh4cHI0TE12MAoyVmk0ZnBDbFh5cURWT0tybnJ4NmxEcW9ZZGpmU2RBL2pWTFpyR1MvUFVOellBUGErZmQ3c0pLalpqQmtNQTRHCkExVWREd0VCL3dRRUF3SUJCakFTQmdOVkhSTUJBZjhFQ0RBR0FRSC9BZ0VDTUIwR0ExVWREZ1FXQkJSZXcvaHcKdHQ2R2VwRDRCWHB3clJWQnlEclhnVEFmQmdOVkhTTUVHREFXZ0JSZXcvaHd0dDZHZXBENEJYcHdyUlZCeURyWApnVEFLQmdncWhrak9QUVFEQWdOSUFEQkZBaUJnSUNVOEFYcmdwY2JHQWsvMXJiZk96N0lDZ29VUVlqN0o2eFFkCi9LWkVpQUloQVB0UWlDQjltODVhRjVCU2xETGlrVUFHSyswOHZBY1VzT3NkSlF2LzcvbDIKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURnakNDQXlpZ0F3SUJBZ0lWQU9xcXZTNHJkWC9YTjBQTTg5eVpZY0hyQlVWd01Bb0dDQ3FHU000OUJBTUMKTURVeEVUQVBCZ05WQkFvVENGUmxjM1FnVEV4RE1TQXdIZ1lEVlFRREV4ZFVaWE4wSUZCeWIyUjFZM1JwYjI0ZwpVbTl2ZENCRFFUQWVGdzB5TURBNE1qQXhPRE0zTXpKYUZ3MHlNREE0TWprd01qTTNNekphTUFBd2dnRWlNQTBHCkNTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFEUzdYMVhuQ2h3UFBJbW9LOEFhcEI1SFdyVUZTYnMKNHg2WEZVYVNvN0RXbGlQaHlpK2dERW12Z3pVZWp5OXAxZGJLNVlPbjFVb3hDYWJxalJ5VEJIVEtwRFRsbm4yaAowMmRGSEg0WUZSZGUzVGdpVTg3b0QrNFgzN3UvenZ5c3F2S0hCd2JHRzFrSmM3ZzVqZC9UOUdtQXZKUVVvQTIxCkk1QXZmY0J1Qy9VY3FZMjY3M05xYzRndS9BemFuMEZpNmhGRTFrbjgvOHpYSjZIT2NaWjh3QThIWkl2YW5DTG8KcHh5ZVdVcFl0aG1FN2xObGVWeW5JZk9TSkFyOGRKNSttekNaZDdiMEFZZjBWQ0Z4K0Urcm1kRWs1WlFzVHd6dgpCb1BRN2VoVW1LZVRUMWZvU3VuK0lwVHZsSjBpb1ZBdm1VeTMvTjhhM3hYc0RpSXN3Q3Q2SjNoYkFnTUJBQUdqCmdnRjlNSUlCZVRBTUJnTlZIUk1CQWY4RUFqQUFNQjBHQTFVZERnUVdCQlR3YjhtYjVDMHJGVjVjSEszN0Zpa2wKU2FSWk5UQWZCZ05WSFNNRUdEQVdnQlJldy9od3R0NkdlcEQ0Qlhwd3JSVkJ5RHJYZ1RDQmlBWUlLd1lCQlFVSApBUUVFZkRCNk1IZ0dDQ3NHQVFVRkJ6QUNobXhvZEhSd09pOHZjM1J2Y21GblpTNW5iMjluYkdWaGNHbHpMbU52CmJTOXdjbWwyWVhSbFkyRmZZMjl1ZEdWdWRGOWpPV0ZsTURjMVlpMWxOREEyTFRRd09HWXRPV1U0TUMxaU1qRTAKT0RoaU1ERmtOV1V2TG5CeWFYWmhkR1ZmWTJGZlpHRjBZUzlqWVM1amNuUXdIZ1lEVlIwUkFRSC9CQlF3RW9JUQpkR1Z6ZEM1bGVHRnRjR3hsTG1OdmJUQitCZ05WSFI4RWR6QjFNSE9nY2FCdmhtMW9kSFJ3T2k4dmMzUnZjbUZuClpTNW5iMjluYkdWaGNHbHpMbU52YlM5d2NtbDJZWFJsWTJGZlkyOXVkR1Z1ZEY5ak9XRmxNRGMxWWkxbE5EQTIKTFRRd09HWXRPV1U0TUMxaU1qRTBPRGhpTURGa05XVXZMbkJ5YVhaaGRHVmZZMkZmWkdGMFlTOWpjbXd1WTNKcwpNQW9HQ0NxR1NNNDlCQU1DQTBnQU1FVUNJSFpvUmVmdnprSWkxQXUraER1alk0WUhYalFqd2V5WGZVeElMZzh0CjBNRytBaUVBKzNPSXp4ckp3Q2NIWThTYjVFN0MxMmhmRE44WTFNT2JoaDJmSHRPSDM2Zz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBMHUxOVY1d29jRHp5SnFDdkFHcVFlUjFxMUJVbTdPTWVseFZHa3FPdzFwWWo0Y292Cm9BeEpyNE0xSG84dmFkWFd5dVdEcDlWS01RbW02bzBja3dSMHlxUTA1WjU5b2ROblJSeCtHQlVYWHQwNElsUE8KNkEvdUY5Kzd2ODc4cktyeWh3Y0d4aHRaQ1hPNE9ZM2YwL1JwZ0x5VUZLQU50U09RTDMzQWJndjFIS21OdXU5egphbk9JTHZ3TTJwOUJZdW9SUk5aSi9QL00xeWVoem5HV2ZNQVBCMlNMMnB3aTZLY2NubGxLV0xZWmhPNVRaWGxjCnB5SHpraVFLL0hTZWZwc3dtWGUyOUFHSDlGUWhjZmhQcTVuUkpPV1VMRThNN3dhRDBPM29WSmluazA5WDZFcnAKL2lLVTc1U2RJcUZRTDVsTXQvemZHdDhWN0E0aUxNQXJlaWQ0V3dJREFRQUJBb0lCQUFNdnJ6c0prdHJQTU9GQQpnQ1JEZDljOGlJYVhvelRrNFd0cTJOd1NPUE9rNVBuZU1nWDY2WW9MTTF3NDBZQ0p5R3JjT0xicVUrcVZ1TDNHClQrNHduUXNLbS9uMEFQWFcyYktEa2F3UGRZRHJXcE82TEYvNURhR3lzWVZlUFJibXBpOC8vZXcwTVk0Zy8yRnYKTVRoa2NzdU5EYmVhQzFyM0pKQnpGOXNSdHl3dHBIdDJDKzlCRjdKVWtjNFhyMlR3NmtpWEF4SWNzT3NzK2s3bgpCMzFtalBvZzJVVmRXVUdtOUJIOXRZc2pMV09DczRRQndydmptREhPbE8vRm50L2NGSzVRYTlYcDgrWVhQR3VtCnBycVg4aUorT1YvYmc3R2dQUkM5YXFRWW4wUWROdWFJd2pQMEY5eE9jcmhvWC8zdHNxUXBTaTJOQ0lYSFIwN00KTDhHUGI5RUNnWUVBOTlVdmpjVXNkTDl5SlRKVlFPcUtHbCtCbkkreUpCTHBIRHFMZ3krTFVzMWs4Ym85dGxXNApnd2R3RHhseTd1b0JLSHNRQ3RrRnRXbTB5YWZEZ1N2TVdDQldlUko2RXV2Q0drK2svWEw3em92Vm5pOGFPSW5qCjlHb3RpVk1LQnBtbHA3c3lCV2FzVlJDbGlzd00yZkxjRE5teHdnQjVvVHBOT2UyanBrQ2dJSlVDZ1lFQTJlRDEKVW02MWdUbXZHelFPcU9nQkFWNERrYm5vQnUwZ3FNYkF2TktnTlpZWm5iRUVSby9HVlJRZnA2SmZlWTM1VUh4dwo0clFxczdYd1d4a1NzVjUwdDhvS2NvdTd2VzBnSGkxdEZOOTlQWEVsdDJQa3lManZMbS92bUFWZUxma0ZlMzFZCjFPUnhwSjJ3dnRvczdKZFRGUHJCZFZKMGIxQVBQSWFjb3NUYlNTOENnWUI0cHFkMDdEV2RUSXBrUTJHdnJiNjMKNGlEMk9CcHdaMmhtM3JXR2t3SFB1TUJZMGVNelBmNEtnL2R3MG1IYW43OGFsdmFUWVYzZkdHdno5Q0ZBWkRNaAovL1E1RXQ2dEVXczRaZWVibjN1bzdQaDgvczlVRVFVUnV4TWFGSHdBQkpMWjJrOGF4QVpIajBnWUR3aCtualcwClo5S2E4S0pGOUYwZVEydDFCMmN0RlFLQmdGWUxpNWJrZGZYMDYveVlVSG5RTmlWdUZZYkZucWF0bTBwTVErM08KV01zUTNranlrYmUwTENXSmJ2N3JGejJRSGpmMURUZmE0MHBadmZTY01FK3Y5L1JsYkQ3VWhHNUkzSGhPaEZmTAo4MUFDa1Z5ZHJNckFqbVVPZTliVHQ5LzhDbmc4aG9wOU5ZeEhZbmZjL2dUcHRqd0EwOG9icURRVnNBNjlNcnJ0CmQ1U3RBb0dCQUpDRVdKSnVNWk1xQTlpbThnbllaZm1sMnd4MngvUDRLbnZEM1p2LzJNdjdXM2ZWRkFZWlJRMTMKR3JIUWV6UWlzV1BkZSthUTlwTkZvaS9NRC80alRQbk9aVFJyZlIwVmtRT0FBY2F3REhIOXp0T1B4VzQ1aDlKWgpqTGZxdXZSSDZoQW9LbGFpZ3pmM29aamZBZjdlTGFWVjdQbFZsSnZOcVlMN1E1cnJSNUVjCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
kind: Secret
metadata:
  annotations:
    cert-manager.io/alt-names: test.example.com
    cert-manager.io/certificate-name: certificate-sample
    cert-manager.io/common-name: ""
    cert-manager.io/ip-sans: ""
    cert-manager.io/issuer-kind: PrivateCAIssuer
    cert-manager.io/issuer-name: privateca-issuer
    cert-manager.io/uri-sans: ""
  creationTimestamp: "2020-08-20T18:37:31Z"
  name: certificate-sample-tls-sscc12
  namespace: default
  ownerReferences:
  - apiVersion: cert-manager.io/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Certificate
    name: certificate-sample
    uid: 1a9ebbea-272a-420d-92b6-50606e0ab78c
  resourceVersion: "1255"
  selfLink: /api/v1/namespaces/default/secrets/certificate-sample-tls-sscc12
  uid: 294a7236-13e7-4ae7-9999-2455a0b49b14
type: kubernetes.io/tls
```

6. Delete Certificate

```bash
kubectl delete -f config/samples/privateca_v1alpha1_privatecarequest.yaml 

kubectl describe certificate certificate-sample
 
kubectl get secret/certificate-sample-tls-bta2 -oyaml
```

#### References:
- [RBAC](https://www.ibm.com/support/knowledgecenter/SSFC4F_1.3.0/cert-manager/3.4.0/cert_manager_rbac.html#operator)
- [stepIssuer](https://github.com/smallstep/step-issuer)



### Appendix

```
go run ./main.go

## Setup
2020-08-20T14:36:03.983-0400	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": ":8080"}
2020-08-20T14:36:03.984-0400	INFO	setup	starting manager
2020-08-20T14:36:03.984-0400	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
2020-08-20T14:36:03.984-0400	INFO	controller-runtime.controller	Starting EventSource	{"controller": "privatecaissuer", "source": "kind source: /, Kind="}
2020-08-20T14:36:03.985-0400	INFO	controller-runtime.controller	Starting EventSource	{"controller": "certificaterequest", "source": "kind source: /, Kind="}
2020-08-20T14:36:04.086-0400	INFO	controller-runtime.controller	Starting Controller	{"controller": "certificaterequest"}
2020-08-20T14:36:04.185-0400	INFO	controller-runtime.controller	Starting Controller	{"controller": "privatecaissuer"}
2020-08-20T14:36:04.185-0400	INFO	controller-runtime.controller	Starting workers	{"controller": "privatecaissuer", "worker count": 1}
2020-08-20T14:36:04.186-0400	INFO	controller-runtime.controller	Starting workers	{"controller": "certificaterequest", "worker count": 1}
2020-08-20T14:36:12.326-0400	DEBUG	controllers.PrivateCAIssuer	PrivateCAIssuerReconciler 	{"privatecaissuer": "default/privateca-issuer", "group": "default/privateca-issuer"}
2020-08-20T14:36:12.326-0400	DEBUG	controllers.PrivateCAIssuer	PrivateCAIssuerReconciler %v	{"privatecaissuer": "default/privateca-issuer", "group": "privateca-issuer"}
2020-08-20T14:36:12.326-0400	DEBUG	controllers.PrivateCAIssuer	Setting Finalizer %v	{"privatecaissuer": "default/privateca-issuer", "privateca.google.com": "privatecaissuer.finalizers.controllers"}
2020-08-20T14:36:12.332-0400	INFO	controllers.PrivateCAIssuer	setting lastTransitionTime for LocalCA condition	{"privatecaissuer": "default/privateca-issuer", "condition": "Ready", "time": "2020-08-20T14:36:12.332-0400"}
2020-08-20T14:36:12.332-0400	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"PrivateCAIssuer","namespace":"default","name":"privateca-issuer","uid":"4b13c2a6-55de-4bca-8595-60f0b185e691","apiVersion":"privateca.google.com/v1alpha1","resourceVersion":"984"}, "reason": "Verified", "message": "Signing CA verified and ready to issue certificates"}
2020-08-20T14:36:12.336-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "privatecaissuer", "name": "privateca-issuer", "namespace": "default"}
2020-08-20T14:36:12.337-0400	DEBUG	controllers.PrivateCAIssuer	PrivateCAIssuerReconciler 	{"privatecaissuer": "default/privateca-issuer", "group": "default/privateca-issuer"}
2020-08-20T14:36:12.337-0400	DEBUG	controllers.PrivateCAIssuer	PrivateCAIssuerReconciler %v	{"privatecaissuer": "default/privateca-issuer", "group": "privateca-issuer"}
2020-08-20T14:36:12.337-0400	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"PrivateCAIssuer","namespace":"default","name":"privateca-issuer","uid":"4b13c2a6-55de-4bca-8595-60f0b185e691","apiVersion":"privateca.google.com/v1alpha1","resourceVersion":"985"}, "reason": "Verified", "message": "Signing CA verified and ready to issue certificates"}
2020-08-20T14:36:12.387-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "privatecaissuer", "name": "privateca-issuer", "namespace": "default"}



## Issue
2020-08-20T14:37:31.773-0400	DEBUG	controllers.PrivateCARequest	 Got Certificate Request %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "default/certificate-sample-1188600488"}
2020-08-20T14:37:31.773-0400	DEBUG	controllers.PrivateCARequest	Setting Finalizer %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "privateca.finalizers.controllers"}
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Signed Cert %s	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "-----BEGIN CERTIFICATE-----\nMIIDgjCCAyigAwIBAgIVAOqqvS4rdX/XN0PM89yZYcHrBUVwMAoGCCqGSM49BAMC\nMDUxETAPBgNVBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24g\nUm9vdCBDQTAeFw0yMDA4MjAxODM3MzJaFw0yMDA4MjkwMjM3MzJaMAAwggEiMA0G\nCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDS7X1XnChwPPImoK8AapB5HWrUFSbs\n4x6XFUaSo7DWliPhyi+gDEmvgzUejy9p1dbK5YOn1UoxCabqjRyTBHTKpDTlnn2h\n02dFHH4YFRde3TgiU87oD+4X37u/zvysqvKHBwbGG1kJc7g5jd/T9GmAvJQUoA21\nI5AvfcBuC/UcqY2673Nqc4gu/Azan0Fi6hFE1kn8/8zXJ6HOcZZ8wA8HZIvanCLo\npxyeWUpYthmE7lNleVynIfOSJAr8dJ5+mzCZd7b0AYf0VCFx+E+rmdEk5ZQsTwzv\nBoPQ7ehUmKeTT1foSun+IpTvlJ0ioVAvmUy3/N8a3xXsDiIswCt6J3hbAgMBAAGj\nggF9MIIBeTAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTwb8mb5C0rFV5cHK37Fikl\nSaRZNTAfBgNVHSMEGDAWgBRew/hwtt6GepD4BXpwrRVByDrXgTCBiAYIKwYBBQUH\nAQEEfDB6MHgGCCsGAQUFBzAChmxodHRwOi8vc3RvcmFnZS5nb29nbGVhcGlzLmNv\nbS9wcml2YXRlY2FfY29udGVudF9jOWFlMDc1Yi1lNDA2LTQwOGYtOWU4MC1iMjE0\nODhiMDFkNWUvLnByaXZhdGVfY2FfZGF0YS9jYS5jcnQwHgYDVR0RAQH/BBQwEoIQ\ndGVzdC5leGFtcGxlLmNvbTB+BgNVHR8EdzB1MHOgcaBvhm1odHRwOi8vc3RvcmFn\nZS5nb29nbGVhcGlzLmNvbS9wcml2YXRlY2FfY29udGVudF9jOWFlMDc1Yi1lNDA2\nLTQwOGYtOWU4MC1iMjE0ODhiMDFkNWUvLnByaXZhdGVfY2FfZGF0YS9jcmwuY3Js\nMAoGCCqGSM49BAMCA0gAMEUCIHZoRefvzkIi1Au+hDujY4YHXjQjweyXfUxILg8t\n0MG+AiEA+3OIzxrJwCcHY8Sb5E7C12hfDN8Y1MObhh2fHtOH36g=\n-----END CERTIFICATE-----\n"}
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Signed CA %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "-----BEGIN CERTIFICATE-----\nMIIB0zCCAXmgAwIBAgIVAJAP12Aim9H8BA3edKfqCbmggn2NMAoGCCqGSM49BAMC\nMDUxETAPBgNVBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24g\nUm9vdCBDQTAeFw0yMDA0MjIwMDE4NTVaFw0zMDA0MjIxMDI2MzZaMDUxETAPBgNV\nBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24gUm9vdCBDQTBZ\nMBMGByqGSM49AgEGCCqGSM49AwEHA0IABADk0WZY30YI9yHp/HpV6Hxhxpr4LMv0\n2Vi4fpClXyqDVOKrnrx6lDqoYdjfSdA/jVLZrGS/PUNzYAPa+fd7sJKjZjBkMA4G\nA1UdDwEB/wQEAwIBBjASBgNVHRMBAf8ECDAGAQH/AgECMB0GA1UdDgQWBBRew/hw\ntt6GepD4BXpwrRVByDrXgTAfBgNVHSMEGDAWgBRew/hwtt6GepD4BXpwrRVByDrX\ngTAKBggqhkjOPQQDAgNIADBFAiBgICU8AXrgpcbGAk/1rbfOz7ICgoUQYj7J6xQd\n/KZEiAIhAPtQiCB9m85aF5BSlDLikUAGK+08vAcUsOsdJQv/7/l2\n-----END CERTIFICATE-----\n"}
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Creating Secret with Name %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "certificate-sample-tls-sscc12"}
I0820 14:37:33.181259  379956 conditions.go:233] Setting lastTransitionTime for CertificateRequest "certificate-sample-1188600488" condition "Ready" to 2020-08-20 14:37:33.181242077 -0400 EDT m=+89.747132348
2020-08-20T14:37:33.181-0400	INFO	controllers.PrivateCARequest	Successfully issued certificate	{"certificaterequest": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.182-0400	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"CertificateRequest","namespace":"default","name":"certificate-sample-1188600488","uid":"d616567c-60f9-4f30-b07b-cccce11b1c7e","apiVersion":"cert-manager.io/v1alpha2","resourceVersion":"1244"}, "reason": "Issued", "message": "Successfully issued certificate"}
2020-08-20T14:37:33.215-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"}
2020-08-20T14:37:33.215-0400	DEBUG	controllers.PrivateCARequest	 Got Certificate Request %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.215-0400	DEBUG	controllers.PrivateCARequest	Setting Finalizer %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "privateca.finalizers.controllers"}
I0820 14:37:33.573152  379956 conditions.go:233] Setting lastTransitionTime for CertificateRequest "certificate-sample-1188600488" condition "Ready" to 2020-08-20 14:37:33.573143269 -0400 EDT m=+90.139033487
2020-08-20T14:37:33.573-0400	INFO	controllers.PrivateCARequest	Unable to create CertifiateRequest ---> Certificate alrady exists: rpc error: code = AlreadyExists desc = Resource 'projects/mineral-minutia-820/locations/us-central1/certificateAuthorities/prod-root/certificates/certificate-sample-tls-sscc12' already exists	{"certificaterequest": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.573-0400	DEBUG	controller-runtime.manager.events	Warning	{"object": {"kind":"CertificateRequest","namespace":"default","name":"certificate-sample-1188600488","uid":"d616567c-60f9-4f30-b07b-cccce11b1c7e","apiVersion":"cert-manager.io/v1alpha2","resourceVersion":"1244"}, "reason": "Failed", "message": "Unable to create CertifiateRequest ---> Certificate alrady exists: rpc error: code = AlreadyExists desc = Resource 'projects/mineral-minutia-820/locations/us-central1/certificateAuthorities/prod-root/certificates/certificate-sample-tls-sscc12' already exists"}
2020-08-20T14:37:33.589-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"


## Revoke
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Signed Cert %s	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "-----BEGIN CERTIFICATE-----\nMIIDgjCCAyigAwIBAgIVAOqqvS4rdX/XN0PM89yZYcHrBUVwMAoGCCqGSM49BAMC\nMDUxETAPBgNVBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24g\nUm9vdCBDQTAeFw0yMDA4MjAxODM3MzJaFw0yMDA4MjkwMjM3MzJaMAAwggEiMA0G\nCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDS7X1XnChwPPImoK8AapB5HWrUFSbs\n4x6XFUaSo7DWliPhyi+gDEmvgzUejy9p1dbK5YOn1UoxCabqjRyTBHTKpDTlnn2h\n02dFHH4YFRde3TgiU87oD+4X37u/zvysqvKHBwbGG1kJc7g5jd/T9GmAvJQUoA21\nI5AvfcBuC/UcqY2673Nqc4gu/Azan0Fi6hFE1kn8/8zXJ6HOcZZ8wA8HZIvanCLo\npxyeWUpYthmE7lNleVynIfOSJAr8dJ5+mzCZd7b0AYf0VCFx+E+rmdEk5ZQsTwzv\nBoPQ7ehUmKeTT1foSun+IpTvlJ0ioVAvmUy3/N8a3xXsDiIswCt6J3hbAgMBAAGj\nggF9MIIBeTAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTwb8mb5C0rFV5cHK37Fikl\nSaRZNTAfBgNVHSMEGDAWgBRew/hwtt6GepD4BXpwrRVByDrXgTCBiAYIKwYBBQUH\nAQEEfDB6MHgGCCsGAQUFBzAChmxodHRwOi8vc3RvcmFnZS5nb29nbGVhcGlzLmNv\nbS9wcml2YXRlY2FfY29udGVudF9jOWFlMDc1Yi1lNDA2LTQwOGYtOWU4MC1iMjE0\nODhiMDFkNWUvLnByaXZhdGVfY2FfZGF0YS9jYS5jcnQwHgYDVR0RAQH/BBQwEoIQ\ndGVzdC5leGFtcGxlLmNvbTB+BgNVHR8EdzB1MHOgcaBvhm1odHRwOi8vc3RvcmFn\nZS5nb29nbGVhcGlzLmNvbS9wcml2YXRlY2FfY29udGVudF9jOWFlMDc1Yi1lNDA2\nLTQwOGYtOWU4MC1iMjE0ODhiMDFkNWUvLnByaXZhdGVfY2FfZGF0YS9jcmwuY3Js\nMAoGCCqGSM49BAMCA0gAMEUCIHZoRefvzkIi1Au+hDujY4YHXjQjweyXfUxILg8t\n0MG+AiEA+3OIzxrJwCcHY8Sb5E7C12hfDN8Y1MObhh2fHtOH36g=\n-----END CERTIFICATE-----\n"}
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Signed CA %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "-----BEGIN CERTIFICATE-----\nMIIB0zCCAXmgAwIBAgIVAJAP12Aim9H8BA3edKfqCbmggn2NMAoGCCqGSM49BAMC\nMDUxETAPBgNVBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24g\nUm9vdCBDQTAeFw0yMDA0MjIwMDE4NTVaFw0zMDA0MjIxMDI2MzZaMDUxETAPBgNV\nBAoTCFRlc3QgTExDMSAwHgYDVQQDExdUZXN0IFByb2R1Y3Rpb24gUm9vdCBDQTBZ\nMBMGByqGSM49AgEGCCqGSM49AwEHA0IABADk0WZY30YI9yHp/HpV6Hxhxpr4LMv0\n2Vi4fpClXyqDVOKrnrx6lDqoYdjfSdA/jVLZrGS/PUNzYAPa+fd7sJKjZjBkMA4G\nA1UdDwEB/wQEAwIBBjASBgNVHRMBAf8ECDAGAQH/AgECMB0GA1UdDgQWBBRew/hw\ntt6GepD4BXpwrRVByDrXgTAfBgNVHSMEGDAWgBRew/hwtt6GepD4BXpwrRVByDrX\ngTAKBggqhkjOPQQDAgNIADBFAiBgICU8AXrgpcbGAk/1rbfOz7ICgoUQYj7J6xQd\n/KZEiAIhAPtQiCB9m85aF5BSlDLikUAGK+08vAcUsOsdJQv/7/l2\n-----END CERTIFICATE-----\n"}
2020-08-20T14:37:33.062-0400	DEBUG	controllers.PrivateCARequest	Creating Secret with Name %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "certificate-sample-tls-sscc12"}
I0820 14:37:33.181259  379956 conditions.go:233] Setting lastTransitionTime for CertificateRequest "certificate-sample-1188600488" condition "Ready" to 2020-08-20 14:37:33.181242077 -0400 EDT m=+89.747132348
2020-08-20T14:37:33.181-0400	INFO	controllers.PrivateCARequest	Successfully issued certificate	{"certificaterequest": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.182-0400	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"CertificateRequest","namespace":"default","name":"certificate-sample-1188600488","uid":"d616567c-60f9-4f30-b07b-cccce11b1c7e","apiVersion":"cert-manager.io/v1alpha2","resourceVersion":"1244"}, "reason": "Issued", "message": "Successfully issued certificate"}
2020-08-20T14:37:33.215-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"}
2020-08-20T14:37:33.215-0400	DEBUG	controllers.PrivateCARequest	 Got Certificate Request %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.215-0400	DEBUG	controllers.PrivateCARequest	Setting Finalizer %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "privateca.finalizers.controllers"}
I0820 14:37:33.573152  379956 conditions.go:233] Setting lastTransitionTime for CertificateRequest "certificate-sample-1188600488" condition "Ready" to 2020-08-20 14:37:33.573143269 -0400 EDT m=+90.139033487
2020-08-20T14:37:33.573-0400	INFO	controllers.PrivateCARequest	Unable to create CertifiateRequest ---> Certificate alrady exists: rpc error: code = AlreadyExists desc = Resource 'projects/mineral-minutia-820/locations/us-central1/certificateAuthorities/prod-root/certificates/certificate-sample-tls-sscc12' already exists	{"certificaterequest": "default/certificate-sample-1188600488"}
2020-08-20T14:37:33.573-0400	DEBUG	controller-runtime.manager.events	Warning	{"object": {"kind":"CertificateRequest","namespace":"default","name":"certificate-sample-1188600488","uid":"d616567c-60f9-4f30-b07b-cccce11b1c7e","apiVersion":"cert-manager.io/v1alpha2","resourceVersion":"1244"}, "reason": "Failed", "message": "Unable to create CertifiateRequest ---> Certificate alrady exists: rpc error: code = AlreadyExists desc = Resource 'projects/mineral-minutia-820/locations/us-central1/certificateAuthorities/prod-root/certificates/certificate-sample-tls-sscc12' already exists"}
2020-08-20T14:37:33.589-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"}
2020-08-20T14:39:27.467-0400	DEBUG	controllers.PrivateCARequest	 Got Certificate Request %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "default/certificate-sample-1188600488"}
2020-08-20T14:39:27.467-0400	DEBUG	controllers.PrivateCARequest	Finalizer called to Delete %s	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "certificate-sample-tls-sscc12"}
2020-08-20T14:39:28.175-0400	DEBUG	controllers.PrivateCARequest	Certificate Revoked: %s	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "CESSATION_OF_OPERATION"}
2020-08-20T14:39:28.206-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"}
2020-08-20T14:39:28.207-0400	DEBUG	controllers.PrivateCARequest	 Got Certificate Request %v	{"certificaterequest": "default/certificate-sample-1188600488", "privateca.google.com": "default/certificate-sample-1188600488"}
2020-08-20T14:39:28.207-0400	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "certificaterequest", "name": "certificate-sample-1188600488", "namespace": "default"}

```