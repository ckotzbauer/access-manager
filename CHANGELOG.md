# Changelog

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
