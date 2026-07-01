package biz

// Student mirrors demo1kratos's students table. This is the article service, so
// it does NOT own students — it keeps this duplicate just to confirm (inside a FOR
// SHARE row lock, taken in CreateArticle/UpdateArticle) that an article's
// student_id refers to a genuine student. The two services share one database, see
// configs/config.yaml; the student service (demo1kratos) owns this table.
//
// Student 与 demo1kratos 的 students 表结构一致。这里是文章服务，并不拥有学生表，
// 保留这份镜像仅用于：建/改文章时（在 FOR SHARE 行锁保护下）校验 student_id 对应的
// 学生确实存在。两个服务共用一个库；学生表的真正归属方是学生服务 demo1kratos。
type Student struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:128;not null"`
	Age       int32
	ClassName string `gorm:"size:128"`
}

func (Student) TableName() string { return "students" }
