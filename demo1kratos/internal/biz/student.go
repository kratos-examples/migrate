package biz

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yylego/go-migrate/checkmigration"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
	"github.com/yylego/must"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Student is the GORM type mapped to the "students" table.
//
// Student 是映射到 students 表的 GORM 模型
type Student struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"size:128;not null"`
	Age       int32
	ClassName string `gorm:"size:128"`
}

func (Student) TableName() string { return "students" }

// The mirrored Article type behind cascade-delete lives in article.go.
// 用于级联删除的 Article 镜像模型定义在 article.go 中。

type StudentUsecase struct {
	data *data.Data
	slog *slog.Logger
}

func NewStudentUsecase(data *data.Data, logger *slog.Logger) (*StudentUsecase, error) {
	// AutoMigrate is disabled on purpose: the schema comes from the sibling migrate-kit
	// project via hand-written go-migrate scripts. Here we just check that the live
	// schema matches the models and crash fast when a migration is pending, since the
	// sibling runs the migration.
	//
	// 有意停用 AutoMigrate：表结构改由旁挂的 migrate-kit 子项目用 go-migrate 脚本化迁移管理。
	// 这里只检查实时表结构是否还与模型匹配，有待迁移则直接断言崩溃（真正的迁移请运行 migrate-kit）。
	//
	// if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
	// 	return nil, err
	// }
	must.Length(checkmigration.CheckMigrate(data.DB(), []any{&Student{}, &Article{}}), 0)
	return &StudentUsecase{data: data, slog: logger}, nil
}

func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	must.Nice(s.Name)

	res := &Student{Name: s.Name, Age: s.Age, ClassName: s.ClassName}
	if err := uc.data.DB().WithContext(ctx).Create(res).Error; err != nil {
		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("create student: %v", err))
	}
	uc.slog.InfoContext(ctx, "created student", "id", res.ID, "name", res.Name)
	return res, nil
}

func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	must.True(s.ID > 0)
	must.Nice(s.Name)

	res := &Student{ID: s.ID}
	upd := uc.data.DB().WithContext(ctx).Model(res).Updates(map[string]any{
		"name":       s.Name,
		"age":        s.Age,
		"class_name": s.ClassName,
	})
	if upd.Error != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", upd.Error))
	}
	if upd.RowsAffected == 0 {
		return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", s.ID))
	}
	if err := uc.data.DB().WithContext(ctx).First(res, s.ID).Error; err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("reload student: %v", err))
	}
	return res, nil
}

func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	// Atomic, race-safe cascade delete, in one transaction:
	//   1. lock the student row (FOR UPDATE) so no article can target
	//      this student meanwhile — CreateArticle takes a conflicting FOR SHARE
	//      lock on the same row, so the two operations serialize;
	//   2. delete the student's articles (children first);
	//   3. delete the student (parent last).
	// 原子且并发安全的级联删除，全部在一个事务里完成：
	//   ① 用 FOR UPDATE 锁住学生行，删除期间不允许给该学生并发新建文章——CreateArticle
	//      会对同一行加互斥的 FOR SHARE 锁，二者因此串行化；
	//   ② 先删该学生名下的文章（子表在前）；
	//   ③ 再删学生本身（父表在后）。
	var notFound bool
	var removedArticles int64
	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
		var s Student
		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).First(&s, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				notFound = true
				return nil
			}
			return err
		}
		del := db.Where("student_id = ?", id).Delete(&Article{})
		if del.Error != nil {
			return del.Error
		}
		removedArticles = del.RowsAffected
		return db.Delete(&Student{}, id).Error
	})
	if err != nil {
		return ebzkratos.New(pb.ErrorTxError("delete student with articles: %v", err))
	}
	if notFound {
		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
	}
	uc.slog.InfoContext(ctx, "deleted student and cascaded articles", "student_id", id, "articles_removed", removedArticles)
	return nil
}

func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
	must.True(id > 0)

	res := &Student{}
	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
		}
		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", err))
	}
	return res, nil
}

func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
	must.True(page >= 1)
	must.True(pageSize >= 1)

	db := uc.data.DB().WithContext(ctx)

	var count int64
	if err := db.Model(&Student{}).Count(&count).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("count students: %v", err))
	}

	var items []*Student
	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
	}
	return items, int32(count), nil
}
