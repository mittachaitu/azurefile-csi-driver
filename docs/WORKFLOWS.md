# Azure File CSI Driver — Workflow Docs

These documents trace the end-to-end code paths of the main CSI RPCs handled by
the Azure File CSI driver, with per-step flows, the Azure/Kubernetes APIs used,
parameter behavior, and Mermaid sequence diagrams.

## Volume lifecycle (controller side)

| RPC | Document | Summary |
|-----|----------|---------|
| `CreateVolume` | [CreateVolumeWorkflow.md](./CreateVolumeWorkflow.md) | Provisions the storage account (find/create) and file share; covers StorageClass parameters (`storageAccount`, `resourceGroup`, `shareName`) and the private endpoint scenario. |
| `DeleteVolume` | [DeleteVolumeWorkflow.md](./DeleteVolumeWorkflow.md) | Deletes only the file share (not the account or secret); covers volumeID parsing, data-plane vs management-plane deletion, and idempotency. |

## Volume lifecycle (node side)

| RPC | Document | Summary |
|-----|----------|---------|
| `NodeStageVolume` / `NodePublishVolume` | [NodeStageAndPublishVolumeWorkflow.md](./NodeStageAndPublishVolumeWorkflow.md) | Mounts the share onto the staging path and bind-mounts it into pods; covers auth modes (account key, MI, WI token, OAuth), SMB/NFS, subdirs, encrypt-in-transit, VHD disk, and kata-cc. |
| `NodeUnpublishVolume` / `NodeUnstageVolume` | [NodeUnpublishAndUnstageVolumeWorkflow.md](./NodeUnpublishAndUnstageVolumeWorkflow.md) | Tears down the per-pod bind mount and the per-node share mount; covers idempotent unmount and proxy/VHD/kata cleanup. |

## Related references

- [driver-parameters.md](./driver-parameters.md) — full StorageClass / volume
  attribute parameter reference.
- [design.md](./design.md) — high-level driver design.
- [csi-debug.md](./csi-debug.md) — debugging guide.
