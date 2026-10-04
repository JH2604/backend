package model

type Student struct {
	StudentID   string `json:"student_id" gorm:"primaryKey;size=20" `
	StudentName string `json:"student_name"`
}
