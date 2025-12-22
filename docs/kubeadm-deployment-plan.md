# AMSS Deployment Plan (Kubeadm, Single VM, External Postgres/Redis)

This plan mirrors production patterns: Kubernetes runs stateless services, while Postgres/Redis live outside the cluster (managed or on the VM). It is written so you can reuse it for other projects with minimal changes.

## Goals
- Production-like Kubernetes setup on a single VM using kubeadm.
- Postgres/Redis external to the cluster (Docker or systemd).
- Ingress for HTTP traffic and optional TLS using a domain or nip.io.
- Repeatable steps for app build, deploy, and updates.

## Assumptions
- VM: Ubuntu 25.04, 4 vCPU / 8 GB RAM / 75 GB disk.
- Public IP: 51.79.85.92.
- No domain initially (optional nip.io).
- AMSS app builds into two binaries: server and worker.

## High-level Architecture
- Kubernetes (single-node kubeadm):
  - amss-server Deployment + Service
  - amss-worker Deployment
  - Ingress for REST + OpenAPI + Swagger UI
- External services on the VM:
  - Postgres (Docker or systemd)
  - Redis (Docker or systemd)
- Optional:
  - cert-manager + Let's Encrypt using nip.io for TLS
  - Prometheus for metrics

## Step 1: VM Base Prep
1. Update system packages:
   - `sudo apt update && sudo apt -y upgrade`
2. Disable swap (required by kubeadm):
   - `sudo swapoff -a`
   - remove swap entry from `/etc/fstab`
3. Ensure required kernel modules and sysctl:
   - `overlay`, `br_netfilter`
   - `net.bridge.bridge-nf-call-iptables=1`
   - `net.ipv4.ip_forward=1`

## Step 2: Install containerd (runtime)
1. Install containerd and configure systemd cgroups:
   - `/etc/containerd/config.toml` -> `SystemdCgroup = true`
2. Restart containerd:
   - `sudo systemctl restart containerd`

## Step 3: Install Kubernetes (kubeadm / kubelet / kubectl)
1. Install kubeadm, kubelet, kubectl via official repo.
2. Pin package versions (avoid accidental upgrades).

## Step 4: Initialize the Cluster
1. Initialize:
   - `sudo kubeadm init --pod-network-cidr=192.168.0.0/16`
2. Configure kubeconfig:
   - `mkdir -p $HOME/.kube`
   - `sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config`
   - `sudo chown $(id -u):$(id -g) $HOME/.kube/config`
3. Allow workloads on control-plane (single node):
   - `kubectl taint nodes --all node-role.kubernetes.io/control-plane-`

## Step 5: Install CNI (Pod Networking)
Recommended: Calico (stable, production-like).
- Apply Calico manifest for the chosen pod CIDR (192.168.0.0/16).

## Step 6: Install Ingress Controller
Recommended: ingress-nginx (common in production).
- Apply ingress-nginx manifest.
- Verify Service type (LoadBalancer not needed on single node; NodePort is fine).

## Step 7: (Optional) TLS with cert-manager + nip.io
If you have no domain, use:
- `amss.51.79.85.92.nip.io` (resolves to your IP automatically)
Steps:
1. Install cert-manager.
2. Create ClusterIssuer (Let's Encrypt).
3. Configure Ingress with TLS for `amss.51.79.85.92.nip.io`.

## Step 8: External Postgres/Redis (Docker)
Using Docker on the VM is acceptable and production-like for learning.

1. Run Postgres (example):
   - Postgres port: 5432
   - Bind on host IP (not only 127.0.0.1).
2. Run Redis (example):
   - Redis port: 6379
3. Firewall:
   - Allow 5432/6379 only from the node’s internal IP or from the cluster node itself.

Environment variables for AMSS:
- `DB_URL=postgres://amss:amss@<VM_IP>:5432/amss?sslmode=disable`
- `REDIS_ADDR=<VM_IP>:6379`

## Step 9: Container Registry Strategy
You need a registry that the cluster can pull from.

Options:
1. Docker Hub / GHCR (most production-like).
2. Local registry container on the VM (simplest for dev).

Recommended: push images to GHCR or Docker Hub.

## Step 10: Build and Push Images
Build and tag:
- `amss-server:latest`
- `amss-worker:latest`

Push to your registry:
- `docker push <registry>/amss-server:latest`
- `docker push <registry>/amss-worker:latest`

## Step 11: Kubernetes Manifests (AMSS)
Create a `deploy/k8s/` directory with:
- `namespace.yaml`
- `configmap.yaml` (non-secret env)
- `secret.yaml` (JWT keys, DB creds)
- `server-deployment.yaml`
- `worker-deployment.yaml`
- `service.yaml`
- `ingress.yaml`
- `migrate-job.yaml` (one-off goose migrations)

Key settings:
- Readiness: `/ready`
- Liveness: `/health`
- Expose `/openapi.yaml` and `/docs`

## Step 12: Migrations
Run a Kubernetes Job that executes goose migrations against Postgres:
- Job should run before deploying app (or as a separate step).
- Ensure `DB_URL` is set in Job env.

## Step 13: Rollouts and Updates
- Update image tags, apply manifests.
- Monitor rollout:
  - `kubectl rollout status deployment/amss-server -n amss`

## Step 14: Observability
Built-in:
- `/metrics` for Prometheus
- OTLP tracing if configured

Optional:
- Deploy Prometheus + Grafana in-cluster.

## Step 15: Security and Ops
- Use SSH keys only, disable root login.
- Firewall open ports:
  - 22 (SSH), 80/443 (Ingress), 6443 (K8s API if needed)
- Store JWT keys and DB credentials in Secrets.
- Backups:
  - Postgres backups via `pg_dump` or volume snapshots.

## Suggested Next Actions
- Confirm OS package compatibility for kubeadm on Ubuntu 25.04.
- Decide if you want TLS now (nip.io) or later.
- Create `deploy/k8s/` manifests for AMSS and apply them.

---
If you want, I can generate the exact manifests for AMSS (Namespace, Deployments, Services, Ingress, Job) based on this plan.*** End Patch"}]});`
