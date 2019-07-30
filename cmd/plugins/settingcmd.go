package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/koinotice/vite/cmd/utils"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/cmd/one"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/hdwallet"
	"github.com/pmker/onegw/pmkoo"
	"github.com/pmker/onegw/pmkoo/types"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"path/filepath"
)

var defaultNodeConfigFileName = "init.json"

var (
	settingFlags = utils.MergeFlags(configFlags, generalFlags)

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
	settingCommand = cli.Command{
		Action:      utils.MigrateFlags(settingAction),
		Name:        "setting",
		Usage:       "load config in db",
		Flags:       walletFlags,
		Category:    "WALLET COMMANDS",
		Description: `wallet create`,
	}
)

// localConsole starts chain new gvite node, attaching chain JavaScript console to it at the same time.
func settingAction(ctx *cli.Context) error {
	cfg := &types.Config{}
	if jsonConf, err := ioutil.ReadFile(defaultNodeConfigFileName); err == nil {

		err = json.Unmarshal(jsonConf, &cfg)
		if err != nil {
			log.Error("Cannot unmarshal the default config file content", "error", err)

		}

	}
	if cfg.Mnemonic == "" {
		mnemonic, err := hdwallet.NewMnemonic(256)
		if err != nil {
			log.Error("create mnemo", "error", err)

		}
		cfg.Mnemonic = mnemonic

		fmt.Printf("me:%s \n", mnemonic)

	}
	ec := Nats()

	fmt.Printf("me:%v \n", &cfg)

	if err := ec.Publish("vgw.onegw.init", &cfg); err != nil {
		fmt.Printf("vgw.onegw.init err %s", err.Error())
	}

	bowPath := filepath.Join(one.DataDir(), "bgdb")

	opt := bgdb.DefaultOptions
	opt.Dir = bowPath
	opt.ValueDir = opt.Dir
	//opt.Logger = emptyLogger{}
	bowdb, err := bgdb.Open(opt)

	if err != nil {
		log.Error(err.Error())

	}
	bowdb.Cache().Set("config", "adfs", 0)
	ss, _ := bowdb.Cache().Get("config")
	fmt.Printf("ss %s", ss)

	v := &pmkoo.Vgw{

		Bowdb: bowdb,

		Nat:    ec,
		Config: cfg,
	}
	data, _ := json.MarshalIndent(cfg, "", " ")

	msg := &nats.Msg{
		Data: []byte(data),
	}
	v.SysInit(msg)

	err = ioutil.WriteFile(defaultNodeConfigFileName, data, 0644)
	if err != nil {
		log.Error("save config", "error", err)
		return err
	}
	select {}

	return nil
}
