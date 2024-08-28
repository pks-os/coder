# How to use the Workspace Activity Refactor Safely

<!-- TODO
- [ ] Link release
- [ ] Link autostop
-->

### Preface

Three Coder enterprise features are influenced by user activity in their
workspaces:

- Autostop: Shuts off workspaces after a period of inactivity.
- Template Insights: Reports usage statistics on templates back to
  administrators.
- Dormancy: Automatically deletes idle workspaces to reduce costs.

In response to some reports of inconsistent behavior across these features, we shipped a refactor in `v2.15.0` to consolidate activity reporting and improve predictability in the
dashboard.

This refactor is a breaking change, and incompatible with outdated Coder clients (VSCode, Jetbrains, CLI) or agents. As a blanket rule, all versions above 2.13.0 are compatible with the refactor. When a client or workspace agent is too far behind, the afformentioned features are suceptible to data loss. Use the following guides to make sure your users and their workspaces are up-to-date to avoid this.


### Upgrading to the required versions

Product | Version | Where to upgrade
------- | ------- | -----------------
CLI     | >= [v2.13.0](https://github.com/coder/coder/releases/tag/v2.13.0) | Use our [installation steps](../install/index.md)
VSCode Extension | TODO | Install from the [VSCode marketplace](https://marketplace.visualstudio.com/items?itemName=coder.coder-remote&ssr=false#review-details)
Jetbrains Gateway Extension | >= [v2.12.0](https://github.com/coder/jetbrains-coder/releases/tag/v2.12.0) | Install from the [Gateway Plugin marketplace](https://plugins.jetbrains.com/plugin/19620-coder/)  


> Note: While we advise having all users on your deployment update their local tools before upgrading to v2.15.0+, this feature can be reverted by enabling the `--legacy-activity-usage` flag on your deployment.


