package repository

import (
	"gin-demo/internal/model"
)

func FindStudent(student_id string) (*model.Student, error) {
	var k model.Student
	err := db.Where("student_id=?", student_id).First(&k).Error
	if err != nil {
		return nil, err
	}
	return &k, nil

}
