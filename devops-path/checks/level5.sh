#!/usr/bin/env bash
# Level 5: Kubernetes. Checks what's running in namespace "visits" on the
# kind cluster "devops", then does a rolling restart under load.
source "$(dirname "$0")/lib.sh"

need kubectl "see Level 5 in README.md for install steps"
need kind "see Level 5 in README.md for install steps"
need jq "sudo apt install jq"

NS=visits
URL="http://localhost:8091"
k() { kubectl --context kind-devops -n "$NS" "$@"; }

section "Cluster"
if kind get clusters 2>/dev/null | grep -qx devops; then
	pass "kind cluster 'devops' exists"
else
	fail "no kind cluster named 'devops'"
	hint "kind create cluster --name devops --config setup/kind-config.yaml"
	summary
fi
if kubectl --context kind-devops get ns "$NS" >/dev/null 2>&1; then
	pass "namespace '$NS' exists"
else
	fail "namespace '$NS' doesn't exist"
	hint "kubectl apply -f k8s/"
	summary
fi

section "App deployment"
dep=$(k get deployment visits -o json 2>/dev/null) || {
	fail "no deployment named 'visits'"
	summary
}
ready=$(jq '.status.readyReplicas // 0' <<<"$dep")
((ready >= 3)) && pass "$ready replicas ready" || fail "$ready replicas ready, want at least 3"

c=$(jq '.spec.template.spec.containers[0]' <<<"$dep")
pod=$(jq '.spec.template.spec' <<<"$dep")
image=$(jq -r '.image' <<<"$c")
if [[ "$image" == *:* ]] && [[ "$image" != *:latest ]]; then
	pass "image has a pinned tag ($image)"
else
	fail "image '$image' uses latest or no tag"
	hint "tag images with a version, e.g. visits:1.0.0, and kind load docker-image it"
fi
[ "$(jq '.readinessProbe != null' <<<"$c")" == true ] && pass "has a readiness probe" || fail "no readiness probe (use /readyz)"
[ "$(jq '.livenessProbe != null' <<<"$c")" == true ] && pass "has a liveness probe" || fail "no liveness probe (use /healthz)"
if [ "$(jq '.resources.requests.cpu and .resources.requests.memory and .resources.limits.memory' <<<"$c")" == true ]; then
	pass "sets CPU/memory requests and a memory limit"
else
	fail "missing resources.requests.cpu, requests.memory or limits.memory"
fi
if [ "$(jq --argjson p "$pod" '(.securityContext.runAsNonRoot // $p.securityContext.runAsNonRoot // false)' <<<"$c")" == true ]; then
	pass "runAsNonRoot is set"
else
	fail "securityContext.runAsNonRoot is not true"
fi
[ "$(jq '.securityContext.readOnlyRootFilesystem // false' <<<"$c")" == true ] && pass "read-only root filesystem" ||
	fail "securityContext.readOnlyRootFilesystem is not true"
grep -q 'configMap' <<<"$c" && pass "reads config from a ConfigMap" || fail "no ConfigMap used (envFrom.configMapRef or valueFrom.configMapKeyRef)"
grep -q 'secretKeyRef\|secretRef' <<<"$c" && pass "reads credentials from a Secret" || fail "no Secret used for DATABASE_URL or the password"
[ "$(jq '.spec.strategy.rollingUpdate.maxUnavailable' <<<"$dep")" == '0' ] && pass "rolling updates never drop below desired capacity (maxUnavailable: 0)" ||
	fail "set strategy.rollingUpdate.maxUnavailable to 0"

section "Database"
sts_ready=$(k get statefulset postgres -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
[ "${sts_ready:-0}" -ge 1 ] && pass "postgres StatefulSet is ready" || fail "no ready StatefulSet named 'postgres'"
bound=$(k get pvc -o json 2>/dev/null | jq '[.items[] | select(.status.phase=="Bound")] | length')
((${bound:-0} >= 1)) && pass "postgres data is on a bound PersistentVolumeClaim" || fail "no bound PVC (use volumeClaimTemplates)"

section "Networking and scaling"
k get service visits >/dev/null 2>&1 && pass "service 'visits' exists" || fail "no service named 'visits'"
k get ingress -o name 2>/dev/null | grep -q . && pass "an Ingress exists" || fail "no Ingress"
k get hpa -o json 2>/dev/null | jq -e '.items[] | select(.spec.scaleTargetRef.name=="visits")' >/dev/null &&
	pass "a HorizontalPodAutoscaler targets the deployment" || fail "no HPA for deployment 'visits'"
k get pdb -o name 2>/dev/null | grep -q . && pass "a PodDisruptionBudget exists" || fail "no PodDisruptionBudget"

if wait_http "$URL/" 15; then
	pass "the app answers through the ingress at $URL"
else
	fail "nothing answers at $URL"
	hint "install ingress-nginx for kind and create the cluster with setup/kind-config.yaml"
	summary
fi
curl -fsS "$URL/readyz" | grep -q '"storage":"postgres"' && pass "the app is connected to Postgres" ||
	fail "/readyz doesn't report postgres storage"
hosts=$(for _ in $(seq 1 30); do curl -fsS "$URL/" 2>/dev/null; echo; done | grep -o '"hostname":"[^"]*"' | sort -u | wc -l)
((hosts >= 2)) && pass "requests are spread across $hosts pods" || fail "all requests went to $hosts pod(s)"

section "Zero-downtime rolling restart"
info "restarting the deployment while sending requests (about 1 minute)..."
results=$(mktemp)
(
	end=$((SECONDS + 300))
	while ((SECONDS < end)) && [ ! -f "$results.stop" ]; do
		curl -s -o /dev/null -w '%{http_code}\n' --max-time 2 "$URL/visits" >>"$results"
		sleep 0.05
	done
) &
loader=$!
sleep 2
k rollout restart deployment visits >/dev/null
if k rollout status deployment visits --timeout 180s >/dev/null 2>&1; then
	pass "rollout completed"
else
	fail "rollout did not finish within 3 minutes"
fi
sleep 3
touch "$results.stop"
wait "$loader"
total=$(wc -l <"$results")
bad=$(grep -cv '^200$' "$results")
rm -f "$results" "$results.stop"
if ((bad == 0 && total > 50)); then
	pass "$total requests during the restart, 0 failed"
else
	fail "$bad of $total requests failed during the rolling restart"
	hint "readiness probe on /readyz + SHUTDOWN_DELAY (or a preStop sleep) + maxUnavailable: 0"
fi

summary
