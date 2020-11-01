# access-manager

![build](https://github.com/ckotzbauer/access-manager/workflows/build/badge.svg)

The Access-Manager is a Kubernetes-Operator using the [Operator-SDK](https://github.com/operator-framework/operator-sdk) to simplify complex [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) configurations in your cluster.

## Motivation

The idea for this came up, when managing many different RBAC-Roles on namespace-basis. This was getting more complex over time, and the administrator always has to ensure that the correct roles are applied for different people or ServiceAccounts in multiple namespaces. The scope of the operator is limited to the creation and removal of `RoleBinding`s and `ClusterRoleBinding`s. So all referenced `Role`s and `ClusterRole`s have to exist. Let's automate it.

## Kubernetes & OS Compatibility

The image contains versions of `k8s.io/client-go`. Kubernetes aims to provide forwards & backwards compatibility of one minor version between client and server:

| access-manager  | k8s.io/client-go | k8s.io/apimachinery | expected kubernetes compatibility |
|-----------------|------------------|---------------------|-----------------------------------|
| master          | v0.19.3          | v0.19.3             | 1.18.x, 1.19.x, 1.20.x            |
| 0.4.0           | v0.19.2          | v0.19.2             | 1.18.x, 1.19.x, 1.20.x            |
| 0.3.0           | v0.18.8          | v0.18.8             | 1.17.x, 1.18.x, 1.19.x            |
| 0.2.0           | v12.0.0          | v0.18.5             | 1.17.x, 1.18.x, 1.19.x            |
| 0.1.0           | v12.0.0          | v0.18.3             | 1.17.x, 1.18.x, 1.19.x            |

See the [release notes](https://github.com/ckotzbauer/access-manager/releases) for specific version compatibility information, including which
combination have been formally tested.

## Installation

**Note:** The `ServiceAccount` must have at least the permissions that it should grant. The `cluster-admin` `ClusterRole` is assigned to the `ServiceAccount` by default.

#### Manifests

```
kubectl apply -f config/crd/access-manager.io_rbacdefinitions.yaml
kubectl apply -f config/rbac
kubectl apply -f config/manager
```

#### Helm-Chart

```
helm repo add ckotzbauer https://ckotzbauer.github.io/helm-charts
helm install ckotzbauer/access-manager
```

## Example Definition

The `RbacDefinition` itself is cluster-scoped.

```yaml
apiVersion: access-manager.io/v1beta1
kind: RbacDefinition
metadata:
  name: example-definition
spec:
  namespaced:
  - namespace:
      name: my-product
    bindings:
    - roleName: my-product-management
      kind: Role
      subjects:
      - name: my-product-team
        kind: Group
      - name: devops-team
        kind: Group
  - namespaceSelector:
      matchLabels:
        ci: "true"
    bindings:
    - roleName: ci-deploy
      kind: ClusterRole
      subjects:
      - name: ci
        namespace: ci-service
        kind: ServiceAccount
  cluster:
  - name: john-view-binding
    clusterRoleName: view
    subjects:
    - name: john
      kind: User
```

This would create the following objects:
- A `RoleBinding` named `my-product-management` in the namespace `my-product` assigning the `my-product-management` `Role` to the `Group`s `my-product-team` and `devops-team`.
- A `RoleBinding` named `ci-deploy` in each namespace labeled with `ci: true` assigning the `ci-deploy` `ClusterRole` to the `ServiceAccount` `ci` in the `ci-service` namespace.
- A `ClusterRoleBinding` named `john-view-binding` assigning the `view` `ClusterRole` to the `User` `john`.

For more details, please read the [api-docs](https://github.com/ckotzbauer/access-manager/blob/master/docs/api.md) and view YAMLs in the `examples` directory.


## Behaviors

- A `RbacDefinition` can be marked as "paused" (set `spec.paused` to `true`), so that the operator will not interfere you.
- The `RoleBinding`s and `ClusterRoleBinding`s are named the same as the given `Role` or `ClusterRole` unless the name is explicitly specified.
- If there is a existing binding with the same name that is not owned by the `RbacDefinition` it is not touched.
- The operator detects changes to all `RbacDefinition`s, `Namespace`s and `ServiceAccount`s automatically.


## Roadmap

- Expose Prometheus metrics about created bindings and reconcile errors.
- Manage secrets (e.g. imagePullSecrets) to be available in certain namespaces automatically.


#### Credits

This projects was inspired by the [RBACManager](https://github.com/FairwindsOps/rbac-manager).


[Contributing](https://github.com/ckotzbauer/access-manager/blob/master/CONTRIBUTING.md)
--------
[License](https://github.com/ckotzbauer/access-manager/blob/master/LICENSE)
--------
[Changelog](https://github.com/ckotzbauer/access-manager/blob/master/CHANGELOG.md)
--------
