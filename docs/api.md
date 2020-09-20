# API Docs

This Document documents the types introduced by the Access-Manager to be consumed by users.

## Table of Contents
* [BindingsSpec](#bindingspec)
* [ClusterSpec](#clusterspec)
* [NamespacedSpec](#namespacedspec)
* [NamespaceSpec](#namespacespec)
* [RbacDefinition](#rbacdefinition)
* [RbacDefinitionList](#rbacdefinitionlist)
* [RbacDefinitionSpec](#rbacdefinitionspec)


## BindingsSpec

BindingsSpec defines the name and "body" of a RoleBinding.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the RoleBinding. Optional, if not set `roleName` is used. | string | false |
| roleName | Name of the Role or ClusterRole to reference. | string | true |
| kind | Kind of the `roleName` Either `Role` or `ClusterRole`. | string | true |
| allServiceAccounts | Whether all `ServiceAccount`s of this namespace should be included as subjects. | bool | false |
| subjects | List of RBAC-Subjects. | [][rbacv1.Subject](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#subject-v1-rbac-authorization-k8s-io) | true |

[Back to TOC](#table-of-contents)

## ClusterSpec

ClusterSpec defines the name and "body" of a ClusterRoleBinding.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the ClusterRoleBinding. Optional, if not set `clusterRoleName` is used. | string | false |
| clusterRoleName | Name of the ClusterRole to reference. | string | true |
| subjects | List of RBAC-Subjects. | [][rbacv1.Subject](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#subject-v1-rbac-authorization-k8s-io) | true |

[Back to TOC](#table-of-contents)

## NamespacedSpec

NamespacedSpec describes a set of RoleBindings to create in different namespaces.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | Single namespace name. Optional, but one of `namespace` or `namespaceSelector` is required. | [NamespaceSpec](#namespacespec) | false |
| namespaceSelector | LabelSelector. Optional, but one of `namespace` or `namespaceSelector` is required. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-core) | false |
| bindings | List of RoleBindings to create. | [][BindingsSpec](#bindingspec) | true |


[Back to TOC](#table-of-contents)

## NamespaceSpec

NamespaceSpec defines a name of a single namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of a single namespace. | string | true |


[Back to TOC](#table-of-contents)

## RbacDefinition

RbacDefinition is the definition object itself.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta) | true |
| spec | | [RbacDefinitionSpec](#rbacdefinitionspec) | true |

[Back to TOC](#table-of-contents)

## RbacDefinitionList

RbacDefinitionList is a list of RbacDefinitions.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard list metadata. | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#listmeta-v1-meta) | true |
| items | List of Definitions. | []*[RbacDefinition](#rbacdefinition) | true |

[Back to TOC](#table-of-contents)

## RbacDefinitionSpec

RbacDefinitionSpec defines namespace- and cluster-spec objects.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| paused | Represents whether any actions on the underlaying managed objects are being performed. Only delete actions will be performed. | bool | false |
| namespaced | Optional, but one of `namespaced` or `cluster` is required. | [NamespacedSpec](#namespacedspec) | false |
| cluster | Optional, but one of `namespaced` or `cluster` is required. | [ClusterSpec](#clusterspec) | false |


[Back to TOC](#table-of-contents)
