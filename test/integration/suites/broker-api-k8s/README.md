# SPIFFE Broker API (Kubernetes) Suite

## Description

Spins up a KIND cluster and exercises the SPIFFE Broker API end-to-end against
real workloads. Covers:

* A broker fetches a pod workload's SVID via `WorkloadPIDReference` (legacy
  PID-based path).
* A broker fetches a pod workload's SVID via `KubernetesObjectReference`
  (`pods/core`).
* A broker fetches a **non-pod** object's SVID via `KubernetesObjectReference`
  (`kustomizations.kustomize.toolkit.fluxcd.io`) — exercises the generic
  object-attestation path that resolves the resource via the REST mapper and
  emits the uniform object-meta selector vocabulary.
* A broker whose `allowed_reference_types` is restricted to PID references
  gets `PermissionDenied` at the gRPC layer when it asks for a
  `KubernetesObjectReference`.
* A workload that is not in the agent's broker allowlist can still use the
  Workload API to fetch its own SVID, but is rejected at the mTLS layer when
  it tries to dial the broker endpoint as a broker.
* A broker can also reach the agent over the **TCP** broker listener (via a
  ClusterIP Service in front of the agent daemonset), proving the same mTLS
  and reference-type semantics apply to remote-style brokers.
* The global `allowed_reference_types_over_tcp` setting denies PID
  references over TCP even when the broker's per-broker allowlist would
  otherwise permit them — fail-closed against remote use of local-only
  reference types.

Only the Flux Kustomization CRD is installed (no controllers); the resource
just needs to exist in the API server so the broker can reference it.
