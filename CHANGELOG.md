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
