#!/usr/bin/env bash
# Level 8: Infrastructure as Code with Terraform (Docker provider, so no
# cloud account is needed).
source "$(dirname "$0")/lib.sh"

need terraform "https://developer.hashicorp.com/terraform/install#linux"
need jq "sudo apt install jq"

TF="$ROOT/terraform"
tf() { terraform -chdir="$TF" "$@"; }

section "Code"
require_file terraform/main.tf "provision the stack with the kreuzwerker/docker provider"
for f in variables.tf outputs.tf; do
	[ -f "$TF/$f" ] && pass "terraform/$f exists" || fail "terraform/$f is missing (split variables and outputs into their own files)"
done
grep -q 'required_providers' "$TF"/*.tf && grep -qE 'version *= *"' "$TF"/*.tf &&
	pass "provider versions are pinned" || fail "pin the provider version in required_providers"
[ -f "$TF/.terraform.lock.hcl" ] && pass ".terraform.lock.hcl exists (commit it)" || fail "run terraform init; commit .terraform.lock.hcl"
if grep -qE '^\*?\.?terraform/?$|^\.terraform/?$|tfstate' "$ROOT/.gitignore" 2>/dev/null; then
	pass ".terraform/ and state files are gitignored"
else
	fail "add .terraform/ and *.tfstate* to .gitignore (state can contain secrets)"
fi
tf fmt -check -recursive >/dev/null 2>&1 && pass "code is formatted" || fail "run: terraform -chdir=terraform fmt"
tf init -input=false >/dev/null 2>&1
if tf validate >/tmp/visits-tf.log 2>&1; then
	pass "terraform validate passes"
else
	fail "terraform validate failed"
	sed 's/^/        /' /tmp/visits-tf.log | head -10
	summary
fi
grep -qE 'sensitive *= *true' "$TF"/*.tf && pass "the database password is marked sensitive" ||
	fail "mark the password variable or output as sensitive = true"

section "State"
resources=$(tf state list 2>/dev/null)
if [ -z "$resources" ]; then
	fail "nothing in the state yet"
	hint "terraform -chdir=terraform apply"
	summary
fi
grep -q 'docker_network' <<<"$resources" && pass "manages a docker_network" || fail "no docker_network resource"
grep -q 'docker_volume' <<<"$resources" && pass "manages a docker_volume for Postgres" || fail "no docker_volume resource"
apps=$(grep -c 'docker_container\..*app' <<<"$resources")
((apps >= 2)) && pass "$apps app containers (count or for_each)" || fail "want at least 2 app containers created with count or for_each"

tf plan -input=false -detailed-exitcode >/dev/null 2>&1
case $? in
0) pass "no drift: plan shows no changes after apply" ;;
2) fail "plan shows changes; infrastructure doesn't match the code (run apply, or fix a resource that always changes)" ;;
*) fail "terraform plan failed" ;;
esac

url=$(tf output -raw app_url 2>/dev/null)
if [ -n "$url" ] && wait_http "$url/readyz" 10; then
	pass "output app_url ($url) serves the app"
	curl -fsS "$url/readyz" | grep -q postgres && pass "the app is connected to Postgres" || fail "app isn't using Postgres"
else
	fail "output 'app_url' is missing or doesn't respond"
fi

summary
