package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "gitlab.jinvovo.vom/mondo/btcli/docs"
)

// @title BT CLI
// @version 1.0.0
// @description BT资产化相关API

// @host 183.66.226.110:10001
// @BasePath /
func main() {
	cfg := NewConfig()
	if err := cfg.Init("config.toml"); err != nil {
		panic("On init toml:" + err.Error())
	}
	router := gin.Default()
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("v1")
	h := NewHandler(cfg)
	{
		v1.Handle("GET", "/genkey", h.GenKey)
		v1.Handle("GET", "/validaddress", h.ValidAddress)
		v1.Handle("GET", "/olobalance", h.OLOBalance)
		v1.Handle("GET", "/tokenbalance", h.TokenBalance)
		v1.Handle("POST", "/buildolotx", h.BuildOLOTx)
		v1.Handle("POST", "/buildtokentx", h.BuildTokenTx)
		v1.Handle("POST", "/buildtokenissuetx", h.BuildTokenIssueTx)
		v1.Handle("POST", "/buildtokenredeemtx", h.BuildTokenRedeemTx)
		v1.Handle("POST", "/buildtokenbatchtx", h.BuildTokenBatchTransferTx)
		v1.Handle("POST", "/buildtokenbatchtxs", h.BuildTokenBatchTransfersTx)
		v1.Handle("POST", "/sendtx", h.SendTx)
		v1.Handle("GET", "/checktx", h.CheckTx)
	}
	if err := router.Run(cfg.Bind); err != nil {
		panic(err)
	}
}
