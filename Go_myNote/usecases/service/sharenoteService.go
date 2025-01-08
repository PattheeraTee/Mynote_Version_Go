package service

import (
	"fmt"
	"miw/usecases/repository"
)

type ShareNoteUseCase interface {
	ShareNoteWithEmail(noteID uint, ownerID uint, email string) ([]map[string]string, error)
	IsUserAllowedToEdit(noteID uint, userID uint) (bool, error)
	RemoveShareByEmail(noteID uint, ownerID uint, email string) error 
	GetSharedEmailsByNoteID(noteID uint) ([]map[string]string, error)
}

type ShareNoteService struct {
	shareRepo repository.ShareNoteRepository
	noteRepo  repository.NoteRepository
}

func NewShareNoteService(shareRepo repository.ShareNoteRepository, noteRepo repository.NoteRepository) *ShareNoteService {
    return &ShareNoteService{
        shareRepo: shareRepo,
        noteRepo:  noteRepo,
    }
}


// Share a note with another user by email
func (s *ShareNoteService) ShareNoteWithEmail(noteID uint, ownerID uint, email string) ([]map[string]string, error) {
	// ตรวจสอบว่า Note เป็นของ Owner
	if _, err := s.noteRepo.GetNoteByIdAndUser(noteID, ownerID); err != nil {
		return nil, fmt.Errorf("note not found or does not belong to the user")
	}

	// ดึงข้อมูล User จาก Email
	user, err := s.shareRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("email not found: %v", err)
	}

	// ตรวจสอบว่า Email เป็นของเจ้าของโน้ตหรือไม่
	if user.UserID == ownerID {
		return nil, fmt.Errorf("cannot share note with the owner")
	}

	// ตรวจสอบว่า Note นี้แชร์กับ Email นี้ไปแล้วหรือไม่
	isAlreadyShared, err := s.shareRepo.IsNoteSharedWithUser(noteID, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if note is already shared: %v", err)
	}
	if isAlreadyShared {
		return nil, fmt.Errorf("this email has already been shared with the note")
	}

	// แชร์ Note
	if err := s.shareRepo.ShareNoteWithUser(noteID, user.UserID); err != nil {
		return nil, fmt.Errorf("failed to share note: %v", err)
	}

	// ดึงอีเมลที่แชร์ทั้งหมดหลังจากการแชร์สำเร็จ
	sharedEmails, err := s.shareRepo.GetSharedEmailsByNoteID(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared emails: %v", err)
	}

	return sharedEmails, nil
}

// Check if a user has edit permissions
func (s *ShareNoteService) IsUserAllowedToEdit(noteID uint, userID uint) (bool, error) {
	isOwner, err := s.noteRepo.IsNoteOwnedByUser(noteID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	isShared, err := s.shareRepo.IsNoteSharedWithUser(noteID, userID)
	if err != nil {
		return false, err
	}
	return isShared, nil
}

func (s *ShareNoteService) RemoveShareByEmail(noteID uint, ownerID uint, email string) error {
    return s.shareRepo.RemoveShareByEmail(noteID, ownerID, email)
}


func (s *ShareNoteService) GetSharedEmailsByNoteID(noteID uint) ([]map[string]string, error) {
	emails, err := s.shareRepo.GetSharedEmailsByNoteID(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared emails: %v", err)
	}
	return emails, nil
}
