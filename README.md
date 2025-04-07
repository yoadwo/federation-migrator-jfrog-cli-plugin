# Artifactory federation migration
A [JFrog CLI plugin](https://www.jfrog.com/confluence/display/CLI/JFrog+CLI#JFrogCLI-JFrogCLIPlugins) for migrating to/from Standalone federation service

## About this plugin
Use the JFrog CLI to migrate your Federation repositories from the legacy Federation service that is part of Artifactory to the standalone Federation service.
Also described how to perform the rollback to the legacy service.

### Pre-requirements
Need to have [JFrog CLI](https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli) installed.

### Install the plugin
Run the following command to install the migration plugin from the official registry of [JFrog CLI plugins](https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli#JFrogCLI-JFrogCLIPlugins)

`jf plugin install federation-migrator`

This command installs the plugin and the jar of the migration tool, which performs all required operations with one command
#### Note
If a previous installation of the plugin exists, we recommend that you remove or uninstall it before running the plugin installer. On rare occasions, the previous installation can cause the plugin installer to fail.

### Migration Procedure

#### Migrating to federation service

Run the following command to execute the migration to the standalone Federation service

`jf federation-migrator mi_rtfs <base url without /artifactory> <admin access token>`

#### Migrating to Artifactory legacy (Rollback)

Run the following command to execute the migration from the standalone Federation service

`jf federation-migrator mi_rt <base url without /artifactory> <admin access token>`
