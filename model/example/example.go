package example

import "gorm.io/gorm"

type ExampleRequest struct {
	Language string `json:"language" gorm:"comment:语言"`
	Code     string `json:"code" gorm:"comment:编码"`
}

type Example struct {
	gorm.Model
	ExampleRequest
}
