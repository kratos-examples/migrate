package biz

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yylego/go-migrate/checkmigration"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
	"github.com/yylego/must"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Article is the GORM type mapped to the "articles" table. This service owns
// the table; demo1kratos keeps a duplicate of it just to cascade-delete a
// student's articles (the two services share one database).
//
// Article 是映射到 articles 表的 GORM 模型，本服务是这张表的归属方；
// demo1kratos 里有一份镜像，仅用于删学生时顺带删文章（两服务共用一个库）
type Article struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"size:256;not null"`
	Content   string `gorm:"type:text"`
	StudentID int64  `gorm:"index"`
}

func (Article) TableName() string { return "articles" }

type ArticleUsecase struct {
	data *data.Data
	slog *slog.Logger
}

func NewArticleUsecase(data *data.Data, logger *slog.Logger) (*ArticleUsecase, error) {
	// AutoMigrate is disabled on purpose: the schema comes from the sibling migrate-kit
	// project via hand-written go-migrate scripts. Here we just check that the live
	// schema matches the models and crash fast when a migration is pending, since the
	// sibling runs the migration.
	//
	// 有意停用 AutoMigrate：表结构改由旁挂的 migrate-kit 子项目用 go-migrate 脚本化迁移管理。
	// 这里只检查实时表结构是否还与模型匹配，有待迁移则直接断言崩溃（真正的迁移请运行 migrate-kit）。
	//
	// if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
	// 	return nil, err
	// }
	must.Length(checkmigration.CheckMigrate(data.DB(), []any{&Article{}, &Student{}}), 0)
	return &ArticleUsecase{data: data, slog: logger}, nil
}

func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	must.Nice(a.Title)
	must.True(a.StudentID > 0)

	// Lock the student row and insert the article in one transaction: the FOR
	// SHARE lock blocks a concurrent DeleteStudent (which takes FOR UPDATE) from
	// removing this student before we commit, so we cannot end up with an article
	// pointing at a student that's being deleted.
	// 在一个事务里锁住学生行再插入文章：FOR SHARE 锁会挡住并发的 DeleteStudent
	// （它持 FOR UPDATE）在本事务提交前删除该学生，从而绝不会创建出指向
	// "正在被删除的学生"的文章
	res := &Article{Title: a.Title, Content: a.Content, StudentID: a.StudentID}
	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
		var student Student
		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).First(&student, a.StudentID).Error; err != nil {
			return err
		}
		return db.Create(res).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ebzkratos.New(pb.ErrorBadParam("student %d does not exist", a.StudentID))
		}
		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("create article: %v", err))
	}
	uc.slog.InfoContext(ctx, "created article", "id", res.ID, "student_id", res.StudentID)
	return res, nil
}

func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	must.True(a.ID > 0)
	must.Nice(a.Title)
	must.True(a.StudentID > 0)

	// Same transaction + FOR SHARE lock as CreateArticle: the (new) owning
	// student cannot be deleted while we re-point the article.
	// 与 CreateArticle 相同的事务 + FOR SHARE 锁：改文章归属期间，新归属的学生不会被并发删除
	res := &Article{ID: a.ID}
	var studentMissing, articleMissing bool
	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
		var student Student
		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).First(&student, a.StudentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				studentMissing = true
				return nil
			}
			return err
		}
		upd := db.Model(res).Updates(map[string]any{
			"title":      a.Title,
			"content":    a.Content,
			"student_id": a.StudentID,
		})
		if upd.Error != nil {
			return upd.Error
		}
		if upd.RowsAffected == 0 {
			articleMissing = true
			return nil
		}
		return db.First(res, a.ID).Error
	})
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("update article: %v", err))
	}
	if studentMissing {
		return nil, ebzkratos.New(pb.ErrorBadParam("student %d does not exist", a.StudentID))
	}
	if articleMissing {
		return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", a.ID))
	}
	return res, nil
}

func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	del := uc.data.DB().WithContext(ctx).Delete(&Article{}, id)
	if del.Error != nil {
		return ebzkratos.New(pb.ErrorDbError("delete article: %v", del.Error))
	}
	if del.RowsAffected == 0 {
		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
	}
	uc.slog.InfoContext(ctx, "deleted article", "id", id)
	return nil
}

func (uc *ArticleUsecase) GetArticle(ctx context.Context, id int64) (*Article, *ebzkratos.Ebz) {
	must.True(id > 0)

	res := &Article{}
	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
		}
		return nil, ebzkratos.New(pb.ErrorDbError("get article: %v", err))
	}
	return res, nil
}

func (uc *ArticleUsecase) ListArticles(ctx context.Context, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	db := uc.data.DB().WithContext(ctx)

	var total int64
	if err := db.Model(&Article{}).Count(&total).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("count articles: %v", err))
	}

	var items []*Article
	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("list articles: %v", err))
	}
	return items, int32(total), nil
}

// ListStudentArticles returns one student's articles, one page at a time. The
// student↔article relationship gets its own endpoint instead of overloading
// ListArticles with an extra flag.
//
// ListStudentArticles 分页返回某个学生的文章。学生↔文章这层关系单独开一个接口，
// 而不是往 ListArticles 上塞过滤参数。
func (uc *ArticleUsecase) ListStudentArticles(ctx context.Context, studentID int64, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
	must.True(studentID > 0)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	db := uc.data.DB().WithContext(ctx)

	var total int64
	if err := db.Model(&Article{}).Where("student_id = ?", studentID).Count(&total).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("count student articles: %v", err))
	}

	var items []*Article
	if err := db.Where("student_id = ?", studentID).Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("list student articles: %v", err))
	}
	return items, int32(total), nil
}
