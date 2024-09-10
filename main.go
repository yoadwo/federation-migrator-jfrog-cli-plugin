package main

import (
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"migration-plugin/commands"
)

func main() {
	plugins.PluginMain(getApp())
}

func getApp() components.App {
	app := components.App{}
	app.Name = "federation-migrator"
	app.Description = "Migrate from the legacy federation to the new service (or rollback)."
	app.Version = "v1.0.0"
	app.Commands = getCommands()
	return app
}

func getCommands() []components.Command {
	return []components.Command{
		commands.GetMigrateToRtfsCommand(),
		commands.GetMigrateToRTCommand(),
	}
}
