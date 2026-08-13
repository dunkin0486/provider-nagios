#!/usr/bin/env sh
set -eu

if [ -z "${NAGIOS_API_TOKEN:-}" ]; then
  echo "NAGIOS_API_TOKEN is required"
  exit 1
fi

if [ -z "${NAGIOS_URL:-}" ]; then
  echo "NAGIOS_URL is required"
  exit 1
fi

echo "Waiting for Kubernetes API readiness..."
for i in $(seq 1 180); do
  if kubectl get --raw=/readyz >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [ "$i" = "180" ]; then
    echo "Kubernetes API did not become ready"
    exit 1
  fi
done

echo "Applying provider CRDs..."
kubectl apply -f package/crds

echo "Applying acceptance provider config and host manifest..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  namespace: default
  name: acceptance-provider-secret
type: Opaque
stringData:
  credentials: |
    {"url": "${NAGIOS_URL}", "token": "${NAGIOS_API_TOKEN}"}
---
apiVersion: nagios.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: acceptance
  namespace: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: default
      name: acceptance-provider-secret
      key: credentials
---
apiVersion: monitoring.nagios.crossplane.io/v1alpha1
kind: Host
metadata:
  name: acceptance-host
  namespace: default
  annotations:
    crossplane.io/external-name: acceptance-host.example.com
spec:
  forProvider:
    address: 127.0.0.1
    maxCheckAttempts: "3"
    checkPeriod: 24x7
    notificationInterval: "30"
    notificationPeriod: 24x7
    contacts:
      - nagiosadmin
    templates:
      - generic-host
  providerConfigRef:
    name: acceptance
    kind: ProviderConfig
EOF

echo "Waiting for Host to become Ready..."
kubectl wait --for=condition=Ready --timeout="${ACCEPTANCE_WAIT_TIMEOUT:-900s}" host.monitoring.nagios.crossplane.io/acceptance-host -n default

echo "Host is Ready. Validating observed state..."
observed_address="$(kubectl get host.monitoring.nagios.crossplane.io acceptance-host -n default -o jsonpath='{.status.atProvider.address}')"
if [ "${observed_address}" != "127.0.0.1" ]; then
  echo "Unexpected observed address: ${observed_address}"
  exit 1
fi

echo "Acceptance test passed"
