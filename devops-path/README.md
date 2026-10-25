# DevOps path

You take a small Go web service (`app/`) from "it runs on my laptop" to
"it's built, tested, shipped, scaled, monitored and provisioned
automatically." You write all the infrastructure; the app is already done.

Most levels have a checker that verifies your work, like tests for code:

```bash
./check.sh 0      # Level 0
./check.sh 2      # Level 2, and so on
```

## The app

`app/` is a small HTTP service:

| Endpoint | What it does |
|---|---|
| `GET /` | `{"message","version","hostname"}`; hostname shows which replica answered |
| `POST /visits`, `GET /visits` | increment and read a counter (Postgres, or memory if no DB) |
| `GET /healthz` | liveness: the process is alive |
| `GET /readyz` | readiness: can serve traffic (DB reachable, not shutting down) |
| `GET /metrics` | Prometheus metrics |

Config: `PORT` (8080), `DATABASE_URL` (empty = memory), `SHUTDOWN_DELAY`
(0s), `LOG_LEVEL` (info). Build flag: `-ldflags "-X main.version=1.2.3"`.
The binary has a `-healthcheck` flag that exits 0 if `/healthz` is OK.

Try it now: `cd app && go run .`, then `curl localhost:8080/` in another terminal.

## Before you start

```bash
cd ~/devops-path
git init && git add . && git commit -m "Start DevOps path"
```

Commit after every level. In Level 4 you'll push this repo to GitHub.

---

## Level 0: Linux and shell (1–2 weeks) · `./check.sh 0`

**Learn:** the filesystem, permissions (`chmod`, `chown`), processes (`ps`,
`top`, `kill`, signals), `systemd` and `journalctl`, networking (`ss`, `ip`,
`curl`, `dig`), pipes, and `grep`, `awk`, `sed`, `sort`, `uniq`, `cut`, `xargs`.
**Resources:** [The Linux Command Line](https://linuxcommand.org/tlcl.php)
(free book), [Bash Guide](https://mywiki.wooledge.org/BashGuide),
[ShellCheck](https://www.shellcheck.net).

Write these scripts in `level0/`. Each starts with `#!/usr/bin/env bash`
and `set -euo pipefail`, and must be executable. The log scripts take the
log file as `$1`. Practise on `data/access.log`.

| Script | Output |
|---|---|
| `top_ips.sh LOG [N]` | top N (default 5) client IPs: `COUNT IP`, most requests first, ties sorted by IP |
| `status_counts.sh LOG` | `CODE COUNT` for each status code, sorted by code |
| `errors_by_hour.sh LOG` | number of 5xx responses per hour: `HH COUNT`, only hours that have any, sorted by hour |
| `slow_requests.sh LOG [SECONDS]` | the 10 slowest requests slower than SECONDS (default 1.0): `TIME PATH`, slowest first, ties by path. Response time is the last field |
| `suspicious.sh LOG` | IPs with 3 or more requests to `/admin` or `/wp-login.php`: `COUNT IP`, most first, ties by IP |
| `sysinfo.sh` | five lines: `hostname: …`, `kernel: …`, `cpus: …`, `memory_total_mb: …`, `disk_free_gb: …` (free space on `/`) |
| `backup.sh SRC DEST` | creates `DEST/backup-YYYYmmdd-HHMMSS.tar.gz` of SRC and keeps only the newest 3. With wrong arguments it prints usage to stderr and exits 1 |

**Bonus (no checker):** write a systemd *user* unit
(`~/.config/systemd/user/visits.service`) that runs the app, restarts it on
crash, and logs to the journal. Kill it with `kill -9` and watch systemd
bring it back (`journalctl --user -u visits -f`).

## Level 1: Git and GitHub (3–5 days) · checklist

**Learn:** commits, branches, merge vs rebase, resolving conflicts,
`git log --graph`, `git bisect`, `git revert` vs `reset`, pull requests,
code review, conventional commits, semantic versioning and tags.
**Resource:** [Pro Git](https://git-scm.com/book) chapters 1–3 and 7.

- [ ] Create a branch, make two commits, and open a PR to yourself with `gh pr create`
- [ ] Create a merge conflict on purpose and resolve it by hand
- [ ] Squash commits with `git rebase -i` and explain when not to rebase
- [ ] Break something, then find the bad commit with `git bisect`
- [ ] Tag a release `v0.1.0` and push the tag
- [ ] Add branch protection on `main`: require a PR and passing CI (after Level 4)

## Level 2: Docker (1 week) · `./check.sh 2`

**Learn:** images vs containers, layers and caching, multi-stage builds,
`ENTRYPOINT` vs `CMD` (exec vs shell form, and why it matters for
signals), users, healthchecks, volumes, networks, registries.
**Resources:** [Docker docs: Get started](https://docs.docker.com/get-started/),
[Dockerfile best practices](https://docs.docker.com/build/building/best-practices/).

Write `app/Dockerfile` and `app/.dockerignore`. The checker requires:
- a multi-stage build; final image **30 MB or less** (distroless or scratch)
- runs as a **non-root** user
- a `HEALTHCHECK` using `/visits -healthcheck` (short interval, e.g. 5s)
- a `VERSION` build arg that sets both the app version (`-ldflags`) and the
  label `org.opencontainers.image.version`
- stops in under 8 seconds on `docker stop` with exit code 0 (graceful shutdown)

Then explain in `NOTES.md`: why does copying `go.mod` before the source
code speed up rebuilds? Prove it by timing two builds.

## Level 3: Docker Compose (1 week) · `./check.sh 3`

**Learn:** services, networks, named volumes, healthchecks and
`depends_on: condition`, env files, scaling, reverse proxies.

Write `compose.yaml`, `nginx/nginx.conf`, `.env.example`, and a `.gitignore`
that ignores `.env`:
- services named **`app`**, **`postgres`** and **`nginx`**
- nginx is the only thing exposed, on **`localhost:8090`**, proxying to the app
- the app and Postgres are not published to the host
- app and Postgres have healthchecks; the app waits for a healthy database
- Postgres data lives in a named volume (survives `docker compose down`)
- no passwords in `compose.yaml`; they come from `.env`
- `docker compose up -d --scale app=3` works, and nginx spreads requests over all 3

## Level 4: CI/CD with GitHub Actions (1 week) · `./check.sh 4`

**Learn:** workflows, jobs, steps, `needs`, matrix builds, caching,
secrets, `GITHUB_TOKEN` permissions, environments, container registries.
**Resource:** [GitHub Actions docs](https://docs.github.com/en/actions).

Push this repo to GitHub (`gh repo create devops-path --private --source . --push`)
and write `.github/workflows/ci.yml`:
1. **test** job: `gofmt -l` (fail if output isn't empty), `go vet`, `go test` in `app/`
2. **build** job (`needs: test`): build the image and push it to
   `ghcr.io/<you>/visits` **only on main or a tag**, tagged with the git SHA
   (and the version on tags)
3. caching for Go modules and Docker layers
4. a `permissions:` block granting only what's needed (`contents: read`, `packages: write`)

Then break a test on a branch, open a PR, and watch CI block it.

## Level 5: Kubernetes (2–3 weeks) · `./check.sh 5`

**Learn:** pods, deployments, ReplicaSets, services, ingress, ConfigMaps,
Secrets, probes, resources, StatefulSets and PVCs, HPA, PDB, rolling
updates, `kubectl` debugging (`describe`, `logs`, `exec`, `port-forward`, `get events`).
**Resources:** [Kubernetes basics tutorial](https://kubernetes.io/docs/tutorials/kubernetes-basics/),
*Kubernetes Up & Running* (book).

**Install** (no sudo needed):
```bash
mkdir -p ~/.local/bin
curl -Lo ~/.local/bin/kubectl "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
curl -Lo ~/.local/bin/kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ~/.local/bin/kubectl ~/.local/bin/kind

kind create cluster --name devops --config setup/kind-config.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```
Build your image and load it into the cluster (kind can't see your local images otherwise):
```bash
docker build -t visits:1.0.0 --build-arg VERSION=1.0.0 app
kind load docker-image visits:1.0.0 --name devops
```

Write manifests in `k8s/` for namespace **`visits`**:
- Deployment **`visits`**: 3+ replicas, a pinned image tag, readiness (`/readyz`)
  and liveness (`/healthz`) probes, CPU/memory requests and a memory limit,
  `runAsNonRoot`, `readOnlyRootFilesystem`, `maxUnavailable: 0`
- ConfigMap for settings and a Secret for database credentials
- StatefulSet **`postgres`** with a `volumeClaimTemplates` PVC, plus its Service
- Service **`visits`**, an Ingress (reachable at `localhost:8091`), an HPA and a PDB

The checker ends with a **zero-downtime test**: it restarts the deployment
while sending requests and fails if even one errors. Getting that to pass
teaches you how readiness, `SHUTDOWN_DELAY` and rolling updates fit together.

## Level 6: Helm (1 week) · `./check.sh 6`

**Learn:** charts, templates, values, `_helpers.tpl`, `helm upgrade --install`,
`helm rollback`, release history.

Turn `k8s/` into a chart at `chart/visits` with `values.yaml` (dev) and
`values-prod.yaml` (more replicas, bigger resources). Replace the raw
manifests with a release named `visits` in namespace `visits`, upgrade it at
least once, and try a rollback. Add a `checksum/config` annotation so pods
restart when the ConfigMap changes.

## Level 7: Observability (1–2 weeks) · `./check.sh 7`

**Learn:** metrics vs logs vs traces, Prometheus (scraping, PromQL, alert
rules), Grafana (datasources, dashboards, provisioning), the RED method
(Rate, Errors, Duration), SLOs.

Create `monitoring/compose.yaml` that adds Prometheus (`localhost:9090`) and
Grafana (`localhost:3300`) to the Level 3 stack:
`docker compose -f compose.yaml -f monitoring/compose.yaml up -d --scale app=3`
- `monitoring/prometheus.yml` scrapes **every** app replica (`dns_sd_configs`)
- `monitoring/alerts.yml` has at least 2 alerts (instance down, error rate over 5%)
- Grafana's datasource is provisioned from `monitoring/grafana/provisioning/`
- a dashboard JSON in `monitoring/grafana/dashboards/` showing request rate,
  error rate and visits per replica

**Bonus:** do the same in Kubernetes with the `kube-prometheus-stack` Helm chart.

## Level 8: Infrastructure as Code with Terraform (1–2 weeks) · `./check.sh 8`

**Learn:** providers, resources, variables, outputs, state, `plan`/`apply`,
`count`/`for_each`, modules, drift.
**Install:** [Terraform for Linux](https://developer.hashicorp.com/terraform/install#linux).

In `terraform/`, use the `kreuzwerker/docker` provider (no cloud account
needed) to create a network, a Postgres container with a volume, and 2+ app
containers using `count` or `for_each`, with output `app_url`. Requirements:
pinned provider version, `variables.tf` and `outputs.tf`, password marked
`sensitive`, `.terraform.lock.hcl` committed, state and `.terraform/`
gitignored, and **no drift** (a second `plan` shows no changes).

**Bonus:** write an Ansible playbook that installs Docker and deploys the
Compose stack to a fresh VM (`multipass launch` gives you one locally).

## Level 9: Security and GitOps (1–2 weeks) · checklist

- [ ] Add a [Trivy](https://trivy.dev) image scan to CI that fails on HIGH/CRITICAL vulnerabilities
- [ ] Add `govulncheck` and Dependabot (or Renovate) to the repo
- [ ] Generate an SBOM for the image and sign it with `cosign`
- [ ] Add Kubernetes `NetworkPolicy`s: only the ingress controller can reach the app, only the app can reach Postgres. Prove it with `kubectl exec ... curl`
- [ ] Stop storing plain Secrets in git: use Sealed Secrets or SOPS
- [ ] Install **Argo CD** on kind and point it at `chart/visits` in your GitHub repo. From then on, deploy only by merging to main (GitOps), and make CI bump the image tag automatically

## Level 10: Capstone: a real deployment (2+ weeks) · checklist

Deploy **FraudGuard** (`~/fraudguard`) for real, end to end:
- [ ] A small cloud server (a VPS, or the AWS/GCP free tier). **Set a billing alert first**
- [ ] Provisioned with Terraform; configured with Ansible or cloud-init
- [ ] k3s (lightweight Kubernetes) or Docker Compose on the server
- [ ] A domain with HTTPS (cert-manager + Let's Encrypt, or Caddy)
- [ ] CI builds, scans and pushes; Argo CD (or a deploy job) rolls it out
- [ ] Prometheus, Grafana and alerts sent to your email, Slack or Discord
- [ ] Automated, tested database backups (restore one to prove it works)
- [ ] A `RUNBOOK.md`: how to deploy, roll back, restore, and what each alert means
- [ ] An architecture diagram and a write-up in the README. This is your portfolio piece

---

## Interview checklist

Be able to explain, with your own project as the example:
- [ ] What happens from `git push` to the new version serving traffic
- [ ] Container vs VM; what namespaces and cgroups do
- [ ] Liveness vs readiness probes, and what goes wrong if you mix them up
- [ ] How a rolling update achieves zero downtime
- [ ] Requests vs limits; what OOMKilled and CPU throttling mean
- [ ] Deployment vs StatefulSet; how Kubernetes Services find pods
- [ ] How you'd debug a pod stuck in `CrashLoopBackOff` or `Pending`
- [ ] Terraform state: what it is, why it's remote and locked in teams
- [ ] Blue/green vs canary deployments
- [ ] What you'd alert on for a web service, and how you'd avoid alert fatigue
