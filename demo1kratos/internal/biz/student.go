package biz

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
	"gorm.io/gorm"
)

type Student struct {
	ID        int64
	Name      string
	Age       int32
	ClassName string
}

type StudentUsecase struct {
	data *data.Data
	log  *log.Helper
}

func NewStudentUsecase(data *data.Data, logger log.Logger) *StudentUsecase {
	return &StudentUsecase{data: data, log: log.NewHelper(logger)}
}

func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	db := uc.data.DB()

	// Use GORM transaction to save student
	// 使用 GORM 事务保存学生
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		student := &models.Student{
			Name: s.Name,
		}
		if err := tx.Create(student).Error; err != nil {
			return err
		}
		s.ID = int64(student.ID)
		return nil
	})
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("db: %v", err))
	}

	var res Student
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("fake: %v", err))
	}
	res.ID = s.ID
	res.Name = s.Name
	return &res, nil
}

func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	var res Student
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
	return nil
}

func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
	db := uc.data.DB()

	var student models.Student
	if err := db.WithContext(ctx).First(&student, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", err))
		}
		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", err))
	}

	return &Student{
		ID:   int64(student.ID),
		Name: student.Name,
	}, nil
}

func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
	var items []*Student
	gofakeit.Slice(&items)
	return items, int32(len(items)), nil
}
