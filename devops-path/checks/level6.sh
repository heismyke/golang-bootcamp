#!/usr/bin/env bash
# Level 6: Helm. Checks chart/visits, then runs the Level 5 checks against
# the Helm release.
source "$(dirname "$0")/lib.sh"

need helm "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
need jq "sudo apt install jq"

CHART="$ROOT/chart/visits"

section "Chart"
require_file chart/visits/Chart.yaml "helm create chart/visits, then replace the templates with your Level 5 manifests"
require_file chart/visits/values-prod.yaml "production overrides: more replicas, bigger resources"

if helm lint "$CHART" -f "$CHART/values-prod.yaml" >/tmp/visits-helm.log 2>&1; then
	pass "helm lint passes"
else
	fail "helm lint failed"
	grep -E 'ERROR|WARNING' /tmp/visits-helm.log | head -10 | sed 's/^/        /'
fi

replicas() {
	helm template visits "$CHART" "$@" 2>/dev/null |
		awk '/^kind: Deployment/{d=1} d && /^  replicas:/{print $2; exit}'
}
dev=$(replicas)
prod=$(replicas -f "$CHART/values-prod.yaml")
if [ -n "$dev" ] && [ -n "$prod" ] && [ "$dev" != "$prod" ]; then
	pass "replicas come from values (dev: $dev, prod: $prod)"
else
	fail "replicas are the same for dev and prod ('$dev' / '$prod')"
	hint "replicas: {{ .Values.replicaCount }}, overridden in values-prod.yaml"
fi
if helm template visits "$CHART" --set image.tag=9.9.9 2>/dev/null | grep -q 'image: .*:9.9.9'; then
	pass "image tag can be set with --set image.tag=..."
else
	fail "--set image.tag=9.9.9 doesn't change the image"
fi
if helm template visits "$CHART" 2>/dev/null | grep -qE 'checksum/(config|secret)'; then
	pass "pods restart when config changes (checksum annotation)"
else
	fail "no checksum/config annotation on the pod template"
	hint "annotations: checksum/config: {{ include (print \$.Template.BasePath \"/configmap.yaml\") . | sha256sum }}"
fi

section "Release"
status=$(helm status visits -n visits --kube-context kind-devops -o json 2>/dev/null | jq -r '.info.status // empty')
if [ "$status" == "deployed" ]; then
	pass "release 'visits' is deployed in namespace 'visits'"
else
	fail "release 'visits' status: '${status:-not found}'"
	hint "kubectl delete -f k8s/ ; helm upgrade --install visits chart/visits -n visits --create-namespace"
	summary
fi
revs=$(helm history visits -n visits --kube-context kind-devops -o json 2>/dev/null | jq length)
((${revs:-0} >= 2)) && pass "release has $revs revisions (you've upgraded it)" ||
	fail "only $revs revision; do a 'helm upgrade' (and try 'helm rollback')"

echo
info "Running the Level 5 checks against the Helm release..."
bash "$ROOT/checks/level5.sh" || FAIL=$((FAIL + 1))

summary
