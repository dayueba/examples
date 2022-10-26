package main

import (
	"net/http"
	"fmt"
	"time"

	"github.com/dtm-labs/client/dtmcli"
	"github.com/lithammer/shortuuid/v3"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db1 *sqlx.DB
	db2 *sqlx.DB
	dtmServer = "http://localhost:36789/api/dtmsvr"
)

func init() {
	dsn1 := "root:123456@tcp(127.0.0.1:3306)/db1?charset=utf8mb4&parseTime=True"
	db1 = sqlx.MustConnect("mysql", dsn1)
	dsn2 := "root:123456@tcp(127.0.0.1:3306)/db2?charset=utf8mb4&parseTime=True"
	db2 = sqlx.MustConnect("mysql", dsn2)
}

type Req struct {
	Amount         int    `json:"amount"`
	Username string `json:"username"`
}

// todo

func main() {
	addr := "127.0.0.1:8080"
	r := gin.Default()
	// 扣减余额
	r.POST("/SagaBTransOut", func(c *gin.Context) {
		req := Req{}
		_ = c.BindJSON(&req) // 好坑，参数要这样获取
		username := req.Username
		amount := req.Amount

		result, err := db1.Exec("update user_account set balance = balance - ? where username = ? and balance >= ?", amount, username, amount)
		raws, _ := result.RowsAffected()
		// 未转账成功也算错
		if err != nil || raws == 0{
			c.String(http.StatusConflict, "failed")
			return
		}
    c.String(http.StatusOK, "Ok!")
  })
	// 有bug，不是幂等
	// 扣减余额的补偿，也就是增加回余额
	r.POST("/SagaBTransOutCom", func(c *gin.Context) {
		req := Req{}
		_ = c.BindJSON(&req) // 好坑，参数要这样获取
		username := req.Username
		amount := req.Amount

		_, err := db1.Exec("update user_account set balance = balance + ? where username = ?", amount, username)
		if err != nil {
			c.String(http.StatusConflict, err.Error())
			return
		}
    c.String(http.StatusOK, "Ok!")
  })
	// 增加余额
	r.POST("/SagaBTransIn", func(c *gin.Context) {
		// 模拟出错
		c.String(http.StatusConflict, "error")
		return

		req := Req{}
		_ = c.BindJSON(&req) // 好坑，参数要这样获取
		username := req.Username
		amount := req.Amount

		_, err := db2.Exec("update user_account set balance = balance + ? where username = ?", amount, username)
		if err != nil {
			c.String(http.StatusConflict, err.Error())
			return
		}
    c.String(http.StatusOK, "Ok!")
  })
	// 扣减余额的补偿，也就是增加回余额
	r.POST("/SagaBTransInCom", func(c *gin.Context) {
		req := Req{}
		_ = c.BindJSON(&req) // 好坑，参数要这样获取
		username := req.Username
		amount := req.Amount

		_, err := db2.Exec("update user_account set balance = balance - ? where username = ?", amount, username)
		if err != nil {
			c.String(http.StatusConflict, err.Error())
			return
		}
    c.String(http.StatusOK, "Ok!")
  })
	
	newaddr := "http://" + addr
	go func() {
		for i := 0; i < 1; i++ {
			time.Sleep(time.Second * 5) // 等待程序启动
			amount := 100
			req1 := gin.H{"username": "A", "amount": amount}
			req2 := gin.H{"username": "B", "amount": amount}
			saga := dtmcli.NewSaga(dtmServer, shortuuid.New()).
				Add(newaddr+"/SagaBTransOut", newaddr+"/SagaBTransOutCom", req1).
				Add(newaddr+"/SagaBTransIn", newaddr+"/SagaBTransInCom", req2)
			err := saga.Submit()
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	r.Run(addr)
}