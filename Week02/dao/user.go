package dao

import (
	"database/sql"

	"github.com/pkg/errors"
)

type Users struct {
	Id   int
	Name string
}

func querySelectUser(user *Users) error {
	return sql.ErrNoRows
}

func SeleectUser(id int) (Users, error) {
	var user Users
	err := querySelectUser(&user)
	return user, errors.Wrap(err, "select user is null")
}
