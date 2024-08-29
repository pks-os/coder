# Workspace Activity Refactor Requirements

### Preface

Three Coder enterprise features are influenced by user activity in their
workspaces:

- [Autostop](../workspaces.md#autostart-and-autostop): Shuts off workspaces after a period of inactivity.
- Template Insights: Reports usage statistics on templates back to
  administrators.
- [Dormancy](../templates/schedule.md#dormancy-threshold-enterprise): Automatically deletes idle workspaces to reduce costs.

In response to some reports of inconsistent behavior across these features, we shipped a refactor in [`v2.15.0`](https://github.com/coder/coder/releases/tag/v2.15.0) to consolidate activity reporting and improve consistency in the
dashboard.

These improvements will lead to greater consistency today, as well as in future activity-related features; however, the refactor is not compatible with all previous versions of the Coder client. 

### Risks

When a client (CLI, Jetbrains extension) or workspace agent is too far behind, the afformentioned features may exhibit incorrect behavior. 

- Autostop: Workspaces may shut off early, even if the user is active in their workspace
- Dormancy: In rare cases, workspaces may be incorrectly marked as dormant
- Template Insights: User activity may be dropped, or reported from the wrong source


### Upgrading to the required versions

The Coder Agents living on existing workspaces and newly created workspaces will automatically update on start. User action is only required to update the following:

Product | Version | Where to upgrade
------- | ------- | -----------------
CLI     | >= [v2.13.0](https://github.com/coder/coder/releases/tag/v2.13.0) | Use our [installation steps](../install/index.md)
Jetbrains Gateway Extension | >= [v2.12.0](https://github.com/coder/jetbrains-coder/releases/tag/v2.12.0) | Install from the [Gateway Plugin marketplace](https://plugins.jetbrains.com/plugin/19620-coder/)  


> Note: While we advise having all users on your deployment update their local tools before upgrading to v2.15.0+, this feature can be reverted by enabling the `--legacy-activity-usage` flag on your deployment.

Once these updates are applied, you can expect to see greater consistency and accuracy in activity related features. If you encounter issues, please [file an issue in our GitHub](https://github.com/coder/coder/issues).
