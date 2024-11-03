package example

import (
	"github.com/beekeeper1010/lvs2/global"
	model "github.com/beekeeper1010/lvs2/model/example"
)

func Create(body model.ExampleRequest) error {
	return global.Db.Create(&model.Example{
		ExampleRequest: model.ExampleRequest{
			Language: body.Language,
			Code:     body.Code,
		},
	}).Error
}

func List() ([]model.Example, error) {
	examples := []model.Example{}
	err := global.Db.Find(&examples).Error
	return examples, err
}
