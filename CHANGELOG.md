## Version 0.5.0 (2021-01-10)

* [[`3d43b144`](https://github.com/ckotzbauer&#x2F;access-manager/commit/3d43b144)] - **build**: fix release-process
* [[`5b98e10b`](https://github.com/ckotzbauer&#x2F;access-manager/commit/5b98e10b)] - **build**: update to go 1.15.6
* [[`873c9d19`](https://github.com/ckotzbauer&#x2F;access-manager/commit/873c9d19)] - **chore**: several automation improvements
* [[`c748a4d3`](https://github.com/ckotzbauer&#x2F;access-manager/commit/c748a4d3)] - **chore**: change create-default-labels action
* [[`3fa90e56`](https://github.com/ckotzbauer&#x2F;access-manager/commit/3fa90e56)] - **chore**: use renovate-preset
* [[`4592a21b`](https://github.com/ckotzbauer&#x2F;access-manager/commit/4592a21b)] - **chore**: Bump pascalgn&#x2F;automerge-action (#70)
* [[`39a35341`](https://github.com/ckotzbauer&#x2F;access-manager/commit/39a35341)] - **chore**: update module k8s.io&#x2F;client-go to v0.20.1 (#69)
* [[`a93da23c`](https://github.com/ckotzbauer&#x2F;access-manager/commit/a93da23c)] - **chore**: Bump actions&#x2F;setup-node from v2.1.3 to v2.1.4 (#66)
* [[`a1fc77d6`](https://github.com/ckotzbauer&#x2F;access-manager/commit/a1fc77d6)] - **chore**: update module sigs.k8s.io&#x2F;controller-runtime to v0.7.0 (#65)
* [[`22bb25b4`](https://github.com/ckotzbauer&#x2F;access-manager/commit/22bb25b4)] - **chore**: update to k8s@1.20
* [[`c825a185`](https://github.com/ckotzbauer&#x2F;access-manager/commit/c825a185)] - **chore**: update module onsi&#x2F;gomega to v1.10.4 (#63)
* [[`d0b780d6`](https://github.com/ckotzbauer&#x2F;access-manager/commit/d0b780d6)] - **chore**: change automerge policy
* [[`907b0341`](https://github.com/ckotzbauer&#x2F;access-manager/commit/907b0341)] - **chore**: Bump actions&#x2F;setup-node from v2.1.2 to v2.1.3
* [[`b226aca3`](https://github.com/ckotzbauer&#x2F;access-manager/commit/b226aca3)] - **chore**: update module k8s.io&#x2F;client-go to v0.20.0
* [[`3b5cc7d9`](https://github.com/ckotzbauer&#x2F;access-manager/commit/3b5cc7d9)] - **chore**: enable automerge
* [[`5f9c6e82`](https://github.com/ckotzbauer&#x2F;access-manager/commit/5f9c6e82)] - **chore**: convert project to multi-group
* [[`ddbcf203`](https://github.com/ckotzbauer&#x2F;access-manager/commit/ddbcf203)] - **chore**: update golang docker tag to v1.15.6
* [[`a1745e4d`](https://github.com/ckotzbauer&#x2F;access-manager/commit/a1745e4d)] - **chore**: release 0.4.2
* [[`a623347e`](https://github.com/ckotzbauer&#x2F;access-manager/commit/a623347e)] - **doc**: fix typo
* [[`81d111f8`](https://github.com/ckotzbauer&#x2F;access-manager/commit/81d111f8)] - **feat**: added sync-secret feature
* [[`00aaa566`](https://github.com/ckotzbauer&#x2F;access-manager/commit/00aaa566)] - **security**: add snyk job
* [[`68dc3249`](https://github.com/ckotzbauer&#x2F;access-manager/commit/68dc3249)] - **security**: add docker-image scan

## Version 0.4.2 (2020-11-22)

* [[`1dfbfcdf`](https://github.com/ckotzbauer&#x2F;access-manager/commit/1dfbfcdf)] - **chore**: fix release-pipeline
* [[`b3a6c6c1`](https://github.com/ckotzbauer&#x2F;access-manager/commit/b3a6c6c1)] - **chore**: rename workflow file
* [[`825d7839`](https://github.com/ckotzbauer&#x2F;access-manager/commit/825d7839)] - **chore**: rename action
* [[`ab918ecc`](https://github.com/ckotzbauer&#x2F;access-manager/commit/ab918ecc)] - **chore**: add release workflow
* [[`addfbfb5`](https://github.com/ckotzbauer&#x2F;access-manager/commit/addfbfb5)] - **chore**: update kubebuilder-action
* [[`2dc9e167`](https://github.com/ckotzbauer&#x2F;access-manager/commit/2dc9e167)] - **chore**: fix test pipeline
* [[`caaa775d`](https://github.com/ckotzbauer&#x2F;access-manager/commit/caaa775d)] - **chore**: refactor docker-build and test-pipeline
* [[`a488520b`](https://github.com/ckotzbauer&#x2F;access-manager/commit/a488520b)] - **chore**: Update module sigs.k8s.io&#x2F;controller-runtime to v0.6.4
* [[`3383b989`](https://github.com/ckotzbauer&#x2F;access-manager/commit/3383b989)] - **chore**: Update module k8s.io&#x2F;client-go to v0.19.4
* [[`7802a89d`](https://github.com/ckotzbauer&#x2F;access-manager/commit/7802a89d)] - **chore**: Update module go-logr&#x2F;logr to v0.3.0
# Changelog

## Version 0.4.1

Released on November 1, 2020.

- Updated k8s.io/api to v0.19.3 #51
- Updated k8s.io/apimachinery to v0.19.3 #51
- Updated k8s.io/client-go to v0.19.3 #51
- Updated onsi/ginko to v1.14.2 #47
- Updated onsi/gomega to v1.10.3 #48
- Updated go to v1.15.3
- FIX: Modify (Cluster-)RoleBindings more selective and avoid unneeded changes.
- FIX: Make "name" and "subjects" of BindingSpec optional in CRD.

**Note:** This release was tested on Kubernetes 1.17.11, 1.18.8 and 1.19.1


## Version 0.4.0

Released on September 22, 2020.

- Updated sigs.k8s.io/controller-runtime to v0.6.3 #37
- Updated k8s.io/api to v0.19.2 #40
- Updated k8s.io/apimachinery to v0.19.2 #40
- Updated k8s.io/client-go to v0.19.2 #40
- Updated onsi/ginko to v1.14.1 #27
- Updated onsi/gomega to v1.10.2 #28
- Updated go-logr/logr to v0.2.1 #29
- Updated go-logr/zapr to 49ca6b4dc551f8fdf9fe385fbd7a60ee3b846a21 #29
- Updated go to v1.15.2 #42
- Updated sigs.k8s.io/kind to v0.9.0 #42
- Updated operator-framework/operator-sdk to v1.0.1 #42
- Added automations with Github Actions.
- FEATURE: Added optional `allServiceAccounts` field
- FEATURE: Added controller for ServiceAccounts to detect changes
- ENHANCEMENT: Avoid replacement of untouched `RoleBinding`s and `ClusterRoleBinding`s.

**Note:** This release was tested on Kubernetes 1.17.11, 1.18.8 and 1.19.1


## Version 0.3.0

Released on August 30, 2020.

- Updated sigs.k8s.io/controller-runtime to v0.6.2 #19
- Updated k8s.io/api to v0.18.8 #21
- Updated k8s.io/apimachinery to v0.18.8 #22
- Updated operator-framework/operator-sdk to v1.0.0 #20
- Updated sigs.k8s.io/controller-tools/cmd/controller-gen to v0.4.0
- Major project overhaul with operator-sdk 1.0.0
- Replaced Travis CI with Github Actions.
- Added matrix tests for different K8s versions.
- Release docker images on each push.

**Note:** This release was tested on Kubernetes 1.17.5, 1.18.8 and 1.19.0


## Version 0.2.0

Released on July 5, 2020.

- Removed default operator metrics.
- Make error-handling for resilient.


## Version 0.1.0

Released on June 11, 2020.

- Initial release.
