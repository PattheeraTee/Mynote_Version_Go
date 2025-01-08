package gormRepository

import (
	"gorm.io/gorm"
	"miw/entities"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) CreateUser(user *entities.User) error {
	// บันทึก Note ลงในฐานข้อมูล
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormUserRepository) UpdateUser(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *GormUserRepository) GetUserById(userID uint) (*entities.User, error) {
	var user entities.User
	if err := r.db.Preload("Notes").
		Preload("Notes.Tags", func(db *gorm.DB) *gorm.DB {
            return db.Select("tag_id, tag_name") // ไม่ดึง Notes ใน Tags
        }).
        Preload("Notes.Reminder").
        Preload("Notes.Event").
	First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetUserEmailByID(userID uint) (string, error) {
	var user entities.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.Email, nil
}