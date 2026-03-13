package models

import "gorm.io/gorm"

// Student represents a student database model
// 学生数据库模型
type Student struct {
	gorm.Model
	Name string `gorm:"type:varchar(255)"`
}

// TableName returns the table name
// 返回表名
func (*Student) TableName() string {
	return "students"
}
