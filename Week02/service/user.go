package service

import (
	"week02/dao"

	"github.com/pkg/errors"
)

func GetInfo(id int) (dao.Users, error) {
	u, e := dao.SeleectUser(id)
	if e != nil {
		return u, errors.WithMessage(e, "get info is null")
	}
	return u, nil
}
