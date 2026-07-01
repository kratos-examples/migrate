package service

import (
	"context"

	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
)

type StudentService struct {
	pb.UnimplementedStudentServiceServer

	uc *biz.StudentUsecase
}

func NewStudentService(uc *biz.StudentUsecase) *StudentService {
	return &StudentService{uc: uc}
}

func (s *StudentService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentReply, error) {
	if req.Name == "" {
		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
	}
	v, ebz := s.uc.CreateStudent(ctx, &biz.Student{
		Name:      req.Name,
		Age:       req.Age,
		ClassName: req.ClassName,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
}

func (s *StudentService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if req.Name == "" {
		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
	}
	v, ebz := s.uc.UpdateStudent(ctx, &biz.Student{
		ID:        req.Id,
		Name:      req.Name,
		Age:       req.Age,
		ClassName: req.ClassName,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.UpdateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
}

func (s *StudentService) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if ebz := s.uc.DeleteStudent(ctx, req.Id); ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.DeleteStudentReply{Success: true}, nil
}

func (s *StudentService) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	v, ebz := s.uc.GetStudent(ctx, req.Id)
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.GetStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
}

func (s *StudentService) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsReply, error) {
	students, count, ebz := s.uc.ListStudents(ctx, req.Page, req.PageSize)
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.StudentInfo, 0, len(students))
	for _, v := range students {
		items = append(items, &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName})
	}
	return &pb.ListStudentsReply{Students: items, Count: count}, nil
}
