#!/usr/bin/env bash
# Level 7: Observability. Checks Prometheus, Grafana and alerting started
# from monitoring/compose.yaml alongside the Level 3 stack.
source "$(dirname "$0")/lib.sh"

need jq "sudo apt install jq"

PROM="http://127.0.0.1:9090"
GRAFANA="http://127.0.0.1:3300"

section "Files"
require_file monitoring/prometheus.yml "scrape config for the app's /metrics"
require_file monitoring/alerts.yml "alerting rules"
if ls "$ROOT"/monitoring/grafana/provisioning/datasources/*.y*ml >/dev/null 2>&1; then
	pass "Grafana datasource is provisioned from a file"
else
	fail "no file in monitoring/grafana/provisioning/datasources/"
fi
if ls "$ROOT"/monitoring/grafana/dashboards/*.json >/dev/null 2>&1; then
	pass "a Grafana dashboard is saved as JSON"
else
	fail "no dashboard JSON in monitoring/grafana/dashboards/"
	hint "build it in the UI, then Share → Export → Save to file"
fi

section "Prometheus"
if ! wait_http "$PROM/-/ready" 5; then
	fail "Prometheus isn't running at $PROM"
	hint "docker compose -f compose.yaml -f monitoring/compose.yaml up -d"
	summary
fi
pass "Prometheus is up"
targets=$(curl -fsS "$PROM/api/v1/targets")
up=$(jq '[.data.activeTargets[] | select(.health=="up")] | length' <<<"$targets")
((up >= 1)) && pass "$up scrape target(s) up" || fail "no healthy scrape targets"
q() { curl -fsS "$PROM/api/v1/query" --data-urlencode "query=$1" | jq -r '.data.result | length'; }
(($(q 'http_requests_total') > 0)) && pass "app metrics are being scraped (http_requests_total)" ||
	fail "http_requests_total not found in Prometheus"
(($(q 'app_info') >= 2)) && pass "metrics from several replicas (app_info has $(q 'app_info') series)" ||
	fail "only one replica scraped; use dns_sd_configs or docker_sd_configs to find every app replica"
rules=$(curl -fsS "$PROM/api/v1/rules" | jq '[.data.groups[].rules[] | select(.type=="alerting")] | length')
((rules >= 2)) && pass "$rules alerting rules loaded" || fail "$rules alerting rules loaded, want at least 2 (e.g. app down, high error rate)"

section "Grafana"
if wait_http "$GRAFANA/api/health" 5; then
	pass "Grafana is up"
	ds=$(curl -fsS -u admin:"${GRAFANA_PASSWORD:-admin}" "$GRAFANA/api/datasources" 2>/dev/null | jq '[.[] | select(.type=="prometheus")] | length' 2>/dev/null)
	((${ds:-0} >= 1)) && pass "Prometheus datasource configured" ||
		fail "no Prometheus datasource (or wrong password; set GRAFANA_PASSWORD)"
	db=$(curl -fsS -u admin:"${GRAFANA_PASSWORD:-admin}" "$GRAFANA/api/search?type=dash-db" 2>/dev/null | jq length 2>/dev/null)
	((${db:-0} >= 1)) && pass "$db dashboard(s) loaded" || fail "no dashboards loaded"
else
	fail "Grafana isn't running at $GRAFANA"
fi

summary
