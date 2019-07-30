package plugins

import (
	"fmt"
	//"github.com/koinotice/vite/cmd/console"

	//"github.com/koinotice/vite/cmd/console"
	"github.com/pmker/onegw/cmd/one"
	"github.com/pmker/onegw/cmd/params"
	"github.com/pmker/onegw/cmd/utils"
	"github.com/pmker/onegw/log15"
	"gopkg.in/urfave/cli.v1"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// onegw is the official command-line client for Vite

var (
	log = log15.New("module", "onegw/main")

	app = cli.NewApp()

	//config
	configFlags = []cli.Flag{
		utils.ConfigFileFlag,
	}
	//general
	generalFlags = []cli.Flag{
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
	}
)

func init() {

	//TODO: Whether the command name is fixed ？
	app.Name = filepath.Base(os.Args[0])
	app.HideVersion = false
	app.Version = params.Version
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "one gate way",
			Email: "leven.zsh@gmail.com",
		},
	}
	app.Copyright = "Copyright 2018-2024 The onegw Authors"
	app.Usage = "the onegw cli application"

	//Import: Please add the New command here
	app.Commands = []cli.Command{
		versionCommand,
		licenseCommand,
		walletCommand,
		settingCommand,
		tokenCommand,

	}
	sort.Sort(cli.CommandsByName(app.Commands))

	//Import: Please add the New Flags here
	app.Flags = utils.MergeFlags(configFlags, generalFlags)

	app.Before = beforeAction
	app.Action = action
	app.After = afterAction
}

func Loading() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func beforeAction(ctx *cli.Context) error {
	max := runtime.NumCPU() + 1
	log.Info("runtime num", "max", max)
	runtime.GOMAXPROCS(max)

	//TODO: we can add dashboard here
	if ctx.GlobalIsSet(utils.PProfEnabledFlag.Name) {
		pprofPort := ctx.GlobalUint(utils.PProfPortFlag.Name)
		var listenAddress string
		if pprofPort == 0 {
			pprofPort = 8080
		}
		listenAddress = fmt.Sprintf("%s:%d", "0.0.0.0", pprofPort)
		var visitAddress = fmt.Sprintf("http://localhost:%d/debug/pprof", pprofPort)

		go func() {
			log.Info("Enable chain performance analysis tool, you can visit the address of `" + visitAddress + "`")
			http.ListenAndServe(listenAddress, nil)
		}()
	}

	return nil
}

func action(ctx *cli.Context) error {

	//Make sure No subCommands were entered,Only the flags
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	err := one.NewOneApp()

	// nodeManager, err := nodemanager.NewDefaultNodeManager(ctx, nodemanager.FullNodeMaker{})
	if err != nil {
		return fmt.Errorf("new one app start error, %+v", err)
	}
	return nil

	//return nodeManager.Start()
}

func afterAction(ctx *cli.Context) error {

	// Resets terminal mode.
	//console.Stdin.Close()

	return nil
}
