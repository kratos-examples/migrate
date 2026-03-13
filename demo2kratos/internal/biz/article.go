package biz

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
	"gorm.io/gorm"
)

type Article struct {
	ID        int64
	Title     string
	Content   string
	StudentID int64
}

type ArticleUsecase struct {
	data *data.Data
	log  *log.Helper
}

func NewArticleUsecase(data *data.Data, logger log.Logger) *ArticleUsecase {
	return &ArticleUsecase{data: data, log: log.NewHelper(logger)}
}

func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	db := uc.data.DB()

	// Use GORM transaction to save article
	// 使用 GORM 事务保存文章
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := &models.Record{
			Message: a.Title,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		a.ID = int64(record.ID)
		return nil
	})
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("db: %v", err))
	}

	var res Article
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("fake: %v", err))
	}
	res.ID = a.ID
	res.Title = a.Title
	return &res, nil
}

func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	var res Article
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, id int64) *ebzkratos.Ebz {
	return nil
}

func (uc *ArticleUsecase) GetArticle(ctx context.Context, id int64) (*Article, *ebzkratos.Ebz) {
	db := uc.data.DB()

	var record models.Record
	if err := db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", err))
		}
		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", err))
	}

	return &Article{
		ID:    int64(record.ID),
		Title: record.Message,
	}, nil
}

func (uc *ArticleUsecase) ListArticles(ctx context.Context, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
	var items []*Article
	gofakeit.Slice(&items)
	return items, int32(len(items)), nil
}
