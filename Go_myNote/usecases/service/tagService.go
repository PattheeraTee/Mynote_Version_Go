package service

import (
	"miw/entities"
	"miw/usecases/repository"
	"fmt"
	"gorm.io/gorm"
)

type TagUseCase interface {
	CreateTag(tag *entities.Tag) error
	GetAllTagsByUserId(userID uint) ([]entities.Tag, error)
	GetTagById(tagID, userID uint) (*entities.Tag, error) 
	UpdateTagName(tagID, userID uint, newName string) error
	DeleteTag(tagID, userID uint) error
}

type TagService struct {
	repo repository.TagRepository
	noteRepo  repository.NoteRepository
}

func NewTagService(repo repository.TagRepository,noteRepo repository.NoteRepository) *TagService {
	return &TagService{
		repo: repo, 
		noteRepo: noteRepo,
	}
}

// CreateTag: สร้าง Tag พร้อมตรวจสอบว่า User เป็นเจ้าของ
func (s *TagService) CreateTag(tag *entities.Tag) error {
	return s.repo.CreateTag(tag)
}

func (s *TagService) GetAllTagsByUserId(userID uint) ([]entities.Tag, error) {
	// ดึงแท็กทั้งหมดที่เป็นของผู้ใช้และแท็กที่เกี่ยวข้องกับโน้ตที่แชร์
	return s.repo.GetAllTagsByUserId(userID)
}


// GetTagById: ดึง Tag ตาม ID และตรวจสอบ UserID
func (s *TagService) GetTagById(tagID, userID uint) (*entities.Tag, error) {
	tag, err := s.repo.GetTagById(tagID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, err
	}

	// ตรวจสอบว่า Tag เป็นของ User หรือโน้ตที่แชร์กับ User
	isOwner := tag.UserID == userID
	isShared := false

	// ตรวจสอบว่าแท็กเกี่ยวข้องกับโน้ตที่แชร์ให้ User หรือไม่
	for _, note := range tag.Notes {
		allowed, err := s.noteRepo.IsUserAllowedToAccessNote(note.NoteID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check access permission: %v", err)
		}
		if allowed {
			isShared = true
			break
		}
	}

	if !isOwner && !isShared {
		return nil, fmt.Errorf("you are not authorized to view this tag")
	}

	return tag, nil
}

// UpdateTagName: แก้ไขชื่อ Tag โดยต้องเป็นเจ้าของเท่านั้น
func (s *TagService) UpdateTagName(tagID, userID uint, newName string) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของแท็กก่อนอัปเดต
	tag, err := s.GetTagById(tagID, userID)
	if err != nil {
		return err
	}

	return s.repo.UpdateTagName(tag.TagID, userID, newName)
}

// DeleteTag: ลบ Tag โดยต้องเป็นเจ้าของเท่านั้น
func (s *TagService) DeleteTag(tagID, userID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของแท็กก่อนลบ
	tag, err := s.GetTagById(tagID, userID)
	if err != nil {
		return err
	}

	return s.repo.DeleteTag(tag.TagID, userID)
}
