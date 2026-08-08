# Tools — what, why, key command

## Jenkins Kubernetes plugin
**What:** provisions a fresh, ephemeral agent pod for every pipeline run, defined inline in the `Jenkinsfile`.
**Why here:** no static build servers to patch or secure — each run starts from a known-clean pod and leaves nothing behind.
**Key idea:** `agent { kubernetes { yaml "..." } }`, then `container('name') { sh '...' }` per stage.

## Checkov
**What:** static analyzer for IaC (Helm charts, Kubernetes YAML, Dockerfiles).
**Why here:** catches misconfigurations (missing `runAsNonRoot`, missing resource limits, mutable image tags) before anything is even built.
**Command:** `checkov -d k8s/ --compact`

## Kaniko
**What:** builds container images without a Docker daemon.
**Why here:** Jenkins agent pods shouldn't run privileged — Kaniko builds in userspace instead, which keeps CI itself compliant with the same no-privileged policy Kyverno enforces on deploys.
**Command:** `/kaniko/executor --context=. --dockerfile=Dockerfile --destination=<image> --digest-file=digest.txt`

## Syft
**What:** generates a Software Bill of Materials (every package/library inside an image).
**Why here:** you can't secure what you can't inventory — and it's the input to the signed attestation step.
**Command:** `syft <image> -o cyclonedx-json=sbom.json`

## Trivy
**What:** vulnerability scanner for images (and filesystems, IaC, SBOMs).
**Why here:** the actual CVE gate — fails the build on CRITICAL/HIGH findings.
**Command:** `trivy image --timeout 15m --severity CRITICAL,HIGH --exit-code 1 <image>`

## Cosign (Sigstore)
**What:** signs and verifies container images and arbitrary artifacts (like SBOMs).
**Why here:** proves an image came from this pipeline and hasn't been tampered with; key-based signing needs no external OIDC provider.
**Commands:**
`cosign sign --key cosign.key --tlog-upload=false --use-signing-config=false --new-bundle-format=false <image>@<digest>`
`cosign attest --key cosign.key --tlog-upload=false --use-signing-config=false --new-bundle-format=false --predicate sbom.json --type cyclonedx <image>@<digest>`

## Helm
**What:** templating and release-management tool for Kubernetes manifests.
**Why here:** installs Jenkins itself, and deploys the app — the image is pinned by **digest**, not tag, so a deploy is always traceable to one exact signed build.
**Command:** `helm upgrade --install app ./k8s/app-chart -n app --set image.digest=<digest>`

## Kyverno
**What:** Kubernetes-native policy engine (admission controller).
**Why here:** closes the loop — the cluster refuses to run any image that isn't cosign-signed, no matter how it got there. Also enforces non-root, non-privileged, no-privilege-escalation, and resource limits on every pod.
**Command to check what it caught:** `kubectl get events -n app --field-selector reason=PolicyViolation`

## Falco
**What:** runtime security — watches syscalls at the kernel/eBPF level.
**Why here:** everything above happens before deploy; Falco is the only tool here watching *after* deploy, for things like a shell being spawned in a running container.
**Command to tail alerts:** `kubectl logs -n falco -l app.kubernetes.io/name=falco -f`

## Velero
**What:** backs up and restores Kubernetes resources + persistent volumes.
**Why here:** DR practice — if the `app` namespace gets nuked, you can prove you can bring it back. Backed by a local MinIO instance (S3-compatible), no cloud account needed.
**Commands:** `velero backup create app-manual --include-namespaces app` / `velero restore create --from-backup app-manual`