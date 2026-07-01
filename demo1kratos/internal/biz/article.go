package biz

// Article mirrors demo2kratos's articles table. This is the student service, so
// it does NOT own articles — it keeps this duplicate just to cascade-delete a
// student's articles when the student is removed (the two services share one
// database, see configs/config.yaml). The article service (demo2kratos) owns
// this table.
//
// Article 与 demo2kratos 的 articles 表结构一致。这里是学生服务，并不拥有文章表，
// 保留这份镜像仅用于：删除学生时顺带删掉他名下的文章（两个服务共用一个库）。
// 文章表的真正归属方是文章服务 demo2kratos。
type Article struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"size:256;not null"`
	Content   string `gorm:"type:text"`
	StudentID int64  `gorm:"index"`
}

func (Article) TableName() string { return "articles" }
