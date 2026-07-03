package data

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData)

type Data struct {
	db *gorm.DB
}

// DB exposes the underlying gorm handle so the biz code can run true queries.
//
// DB 暴露底层 gorm 句柄，供 biz 层执行真实的数据库读写
func (d *Data) DB() *gorm.DB {
	return d.db
}

func NewData(c *conf.Data, logger *slog.Logger) (*Data, func(), error) {
	must.Same(c.Database.Driver, "postgres")
	db := rese.P1(gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{}))
	cleanup := func() {
		logger.Info("closing the data resources")
		_ = rese.P1(db.DB()).Close()
	}
	return &Data{db: db}, cleanup, nil
}
