package plugins

import (
	"fmt"

	"github.com/koinotice/vite/cmd/nodemanager"
	"github.com/koinotice/vite/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	demoFlags = utils.MergeFlags(configFlags, generalFlags )

	//demo,please add this `demoCommand` to one.go
	/**
	app.Commands = []cli.Command{
		versionCommand,
		licenseCommand,
		consoleCommand,
		attachCommand,
		demoCommand,
	}
	*/
	demoCommand = cli.Command{
		Action:      utils.MigrateFlags(demoAction),
		Name:        "demo",
		Usage:       "demo",
		Flags:       demoFlags,
		Category:    "DEMO COMMANDS",
		Description: `demo`,
	}
)

// localConsole starts chain new gvite node, attaching chain JavaScript console to it at the same time.
func demoAction(ctx *cli.Context) error {

	// Create and start the node based on the CLI flags
	nodeManager, err := nodemanager.NewSubCmdNodeManager(ctx, nodemanager.FullNodeMaker{})
	if err != nil {
		fmt.Println("demo error", err)
		return err
	}
	nodeManager.Start()
	defer nodeManager.Stop()

	//Tips: add your code here
	fmt.Println("demo print")
	return nil
}
