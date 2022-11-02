package main

import (
	"time"
	"fmt"

	"tccdemo/server"

	"github.com/dtm-labs/client/dtmcli"
	// "github.com/dtm-labs/client/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lithammer/shortuuid/v3"
)

var Busi = "http://localhost:8080/api/busi"

var BusiAPI = "/api/busi"

const DefaultHTTPServer = "http://localhost:36789/api/dtmsvr"

func main() {
	app := gin.Default()
	app.POST(BusiAPI+"/TccBTransInTry", server.TccBTransInTryHandler())
	app.POST(BusiAPI+"/TccBTransInConfirm", server.TccBTransInConfirmHandler())
	app.POST(BusiAPI+"/TccBTransInCancel", server.TccBTransInCancelHandler())
	app.POST(BusiAPI+"/TccBTransOutTry", server.TccBTransOutTryhandler())
	app.POST(BusiAPI+"/TccBTransOutConfirm", server.TccBTransOutConfirmHandler())
	app.POST(BusiAPI+"/TccBTransOutCancel", server.TccBTransOutCancelHandler())
	
	go func() {
		time.Sleep(time.Second * 5)
		gid := shortuuid.New()
		err := dtmcli.TccGlobalTransaction(DefaultHTTPServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			resp, err := tcc.CallBranch(
				&server.ReqHTTP{Amount: 30}, 
				Busi+"/TccBTransOutTry",
				Busi+"/TccBTransOutConfirm", 
				Busi+"/TccBTransOutCancel",
			)
			if err != nil {
				return resp, err
			}
			return tcc.CallBranch(&server.ReqHTTP{Amount: 30}, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
		})
		// logger.FatalIfError(err)
		if err != nil {
			fmt.Println("转账失败", err)
		}
	}()
	app.Run(":8080")
}
