package data

import (
	"comment-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCommentRepo)

// Data .
type Data struct {
	commentDB *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	commentDB, err := gorm.Open(mysql.Open(c.Database.Dsn), &gorm.Config{
		//Logger: glooger.Default.LogMode(glooger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		commentDB: commentDB,
	}, cleanup, nil
}
