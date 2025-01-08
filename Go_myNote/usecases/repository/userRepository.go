package repository

import (
	"miw/entities"
)

type UserRepository interface {
	CreateUser(user *entities.User) error
	UpdateUser(user *entities.User) error
	GetUserById(userID uint) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
	GetUserEmailByID(userID uint) (string, error)
}