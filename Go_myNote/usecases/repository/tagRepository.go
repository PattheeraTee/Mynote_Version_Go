package repository

import (
	"miw/entities"
)

type TagRepository interface {
	CreateTag(tag *entities.Tag) error
	GetAllTagsByUserId(userID uint) ([]entities.Tag, error)
	GetTagById(tagID uint) (*entities.Tag, error) 
	GetTagsByUser(userID uint) ([]entities.Tag, error)
	UpdateTagName(tagID, userID uint, newName string) error
	DeleteTag(tagID, userID uint) error
}
