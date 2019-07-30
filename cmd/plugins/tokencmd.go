package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/koinotice/vite/cmd/utils"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/pmkoo"
	"github.com/pmker/onegw/pmkoo/models"
	"github.com/pmker/onegw/pmkoo/types"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"time"
)

func Nats() *nats.EncodedConn {
	config := &types.Config{}
	if jsonConf, err := ioutil.ReadFile(defaultNodeConfigFileName); err == nil {

		err = json.Unmarshal(jsonConf, &config)
		if err != nil {
			log.Error("Cannot unmarshal the default config file content", "error", err)

		}

	}
	nc, err := nats.Connect(config.NatsConnectUrl)
	if err != nil {

	}
	//defer nc.Close()

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {

	}
	return ec
}

var (
	tokenCommand = cli.Command{
		Action: utils.MigrateFlags(tokenAction),
		Name:   "token",
		Usage:  "token init",
		//Flags:       walletFlags,
		Category:    "WALLET COMMANDS",
		Description: `wallet create`,
	}
)

// localConsole starts chain new gvite node, attaching chain JavaScript console to it at the same time.
func tokenAction(ctx *cli.Context) error {
	tokens := []*pmkoo.Token{}
	//t1 := &pmkoo.Token{
	//	Address:          "0x514910771af9ca656af840dff83e8264ecf986ca",
	//	Symbol:           "LINK",
	//	Name:             "ChainLink Token",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_e794a04389cc28463ec35ffd",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t1)
	//
	//t2 := &pmkoo.Token{
	//	Address:          "0xa15c7ebe1f07caf6bff097d8a589fb8ac49ae5b3",
	//	Symbol:           "NPXS",
	//	Name:             "Pundi X Token",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_4748b91eb2b5c1cdfd0486bc",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t2)
	//t3 := &pmkoo.Token{
	//	Address:          "0xe41d2489571d322189246dafa5ebde1f4699f498",
	//	Symbol:           "ZRX",
	//	Name:             "ZRX",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_0a81c02b5b94653081baed20",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t3)
	//
	//t4 := &pmkoo.Token{
	//	Address:          "0xbbbbca6a901c926f240b89eacb641d8aec7aeafd",
	//	Symbol:           "LRC",
	//	Name:             "LoopringCoin",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_25e5f191cbb00a88a6267e0f",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t4)
	//
	//t5 := &pmkoo.Token{
	//	Address:          "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
	//	Symbol:           "BRC",
	//	Name:             "Baer Chain",
	//	Decimals:         8,
	//	ViteTokenAddress: "tti_a45e13f43bcddb34209918a3",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t5)
	//
	//t6 := &pmkoo.Token{
	//	Address:          "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07",
	//	Symbol:           "OMG",
	//	Name:             "OmiseGO",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_288b0ca5799e3e094ee9a4fa",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t6)
	//
	//t7 := &pmkoo.Token{
	//	Address:          "0x8971f9fd7196e5cee2c1032b50f656855af7dd26",
	//	Symbol:           "LAMB",
	//	Name:             "Lambda",
	//	Decimals:         18,
	//	ViteTokenAddress: "tti_c0bd4e1dc82e7b24429d289f",
	//	Created:          time.Now(),
	//}
	//tokens = append(tokens, t7)


	test := &pmkoo.Token{
		Address:          "0x64b140435482dc1cad7632937f14363d19850f83",
		Symbol:           "NICE",
		Name:             "NICE",
		Decimals:         18,
		ViteTokenAddress: "tti_25e5f191cbb00a88a6267e0f",
		Created:          time.Now(),
	}
	tokens = append(tokens, test)

	models.Connect("postgres://pmker:Zheli123@127.0.0.1/pmker?sslmode=disable")


	nats := Nats()
	for index, token := range tokens {
		fmt.Printf("index:%d,token:%+v \n", index, token)

		//Nats.publish("vgw.token.create",t2)
		//tokenData := &models.Token{
		//	TokenId: token.ViteTokenAddress,
		//	TokenSymbol:token.Symbol,
		//	Type:1,
		//	//UpdatedAt:  time.Now().UTC(),
		//	//CreatedAt:  time.Now().UTC(),
		//}
		//err2 := models.TokenDaoPG.InsertToken(tokenData)

		//if err2 != nil {
		//	fmt.Printf("insert wallet err %s", err2.Error())
		//}
		if err := nats.Publish("vgw.token.create", token); err != nil {
			fmt.Printf("asdfasdf err %s", err.Error())
		}
	}

	select {

	}

	return nil
}
