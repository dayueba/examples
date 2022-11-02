package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmcli/dtmimp"
	"github.com/dtm-labs/client/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func WrapHandler(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		ret := fn(c)
		status, res := dtmcli.Result2HttpJSON(ret)

		b, _ := json.Marshal(res)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		}
		c.JSON(status, res)
	}
}

func MustBarrierFromGin(c *gin.Context) *dtmcli.BranchBarrier {
	ti, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
	logger.FatalIfError(err)
	return ti
}

func pdbGet() *sql.DB {
	db, err := dtmimp.PooledDB(BusiConf)
	logger.FatalIfError(err)
	return db
}

type ReqHTTP struct {
	Amount         int    `json:"amount"`
	TransInResult  string `json:"trans_in_result"`
	TransOutResult string `json:"trans_out_Result"`
	Store          string `json:"store"` // default mysql, value can be mysql|redis
}

func reqFrom(c *gin.Context) *ReqHTTP {
	v, ok := c.Get("trans_req")
	if !ok {
		req := ReqHTTP{}
		err := c.BindJSON(&req)
		logger.FatalIfError(err)
		c.Set("trans_req", &req)
		v = &req
	}
	return v.(*ReqHTTP)
}

func (t *ReqHTTP) String() string {
	return fmt.Sprintf("amount: %d transIn: %s transOut: %s", t.Amount, t.TransInResult, t.TransOutResult)
}

func string2DtmError(str string) error {
	return map[string]error{
		dtmcli.ResultFailure: dtmcli.ErrFailure,
		dtmcli.ResultOngoing: dtmcli.ErrOngoing,
		dtmcli.ResultSuccess: nil,
		"":                   nil,
	}[str]
}

