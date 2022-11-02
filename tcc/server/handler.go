package server

import (
	"database/sql"
	"errors"
	"fmt"
	// "time"

	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmcli/dtmimp"
	"github.com/gin-gonic/gin"
)

const (
	userA = "A" // 转出的username
	userB = "B" // 转入的username
)

var BusiConf = dtmcli.DBConf{
	Driver:   "mysql",
	Host:     "127.0.0.1",
	Port:     3306,
	User:     "root",
	Password: "123456",
}

// 冻结资金
func tccAdjustTrading(db dtmcli.DB, uname string, amount int) error {
	fmt.Println("冻结资金", uname, amount)
	affected, err := dtmimp.DBExec(BusiConf.Driver, db, `update db1.user_account
		set trading_balance=trading_balance+?
		where username=? and trading_balance + ? + balance >= 0`, amount, uname, amount)
	if err == nil && affected == 0 {
		return errors.New("update error, maybe balance not enough")
	}
	return err
}

// 解冻资金
func tccAdjustBalance(db dtmcli.DB, uname string, amount int) error {
	fmt.Println("解冻资金", uname, amount)
	affected, err := dtmimp.DBExec(BusiConf.Driver, db, `update db1.user_account
		set trading_balance=trading_balance-?,
		balance=balance+? where username=?`, amount, amount, uname)
	if err == nil && affected == 0 {
		return errors.New("update user_account 0 rows")
	}
	return err
}

func TccBTransInTryHandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		// req := reqFrom(c)
		// if req.TransInResult != "" {
		// 	return string2DtmError(req.TransInResult)
		// }
		// MustBarrierFromGin(c).CallWithDB(pdbGet(), func(tx *sql.Tx) error {
		// 	return tccAdjustTrading(tx, userB, req.Amount)
		// })
		return dtmcli.ErrFailure
	})
}

func TccBTransInConfirmHandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		return MustBarrierFromGin(c).CallWithDB(pdbGet(), func(tx *sql.Tx) error {
			return tccAdjustBalance(tx, userB, reqFrom(c).Amount)
		})
	})
}

func TccBTransInCancelHandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		return MustBarrierFromGin(c).CallWithDB(pdbGet(), func(tx *sql.Tx) error {
			fmt.Println(reqFrom(c).Amount)
			fmt.Println("调用接口啦！！！")
			return tccAdjustTrading(tx, userB, -reqFrom(c).Amount)
		})
	})
}

func TccBTransOutTryhandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		req := reqFrom(c)
		if req.TransOutResult != "" {
			return string2DtmError(req.TransOutResult)
		}
		bb := MustBarrierFromGin(c)
		return bb.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
			return tccAdjustTrading(tx, userA, -req.Amount)
		})
		// return dtmcli.ErrFailure
	})
}

func TccBTransOutCancelHandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		// fmt.Println("-------------------", time.Now())
		// time.Sleep(5 * time.Second)
		// fmt.Println("---------------------", time.Now())
		req := reqFrom(c)
		bb := MustBarrierFromGin(c)
		return bb.CallWithDB(pdbGet(), func(tx *sql.Tx) error {
			return tccAdjustTrading(tx, userA, req.Amount)
		})
	})
}

func TccBTransOutConfirmHandler() gin.HandlerFunc {
	return WrapHandler(func(c *gin.Context) interface{} {
		return MustBarrierFromGin(c).CallWithDB(pdbGet(), func(tx *sql.Tx) error {
			return tccAdjustBalance(tx, userA, -reqFrom(c).Amount)
		})
	})
}
