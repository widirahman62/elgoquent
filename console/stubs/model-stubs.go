package stubs

type ModelStubs struct {}

func (*ModelStubs) ODM(param ...string)string{
	return `
package `+param[0]+`

import (
    "`+param[1]+`/database/eloquent/model"
)

type Field struct {

}

func Model() *model.ODM {
	return &model.ODM{
		Connection:  "",
		Entity:      "",
		Field:       Field{},
		FillableTag: []string{},
		GuardedTag:  []string{},
	}
}
`
}
