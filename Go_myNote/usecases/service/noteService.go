package service

import (
	"fmt"
	"miw/entities"
	"miw/usecases/repository"
	"time"

	"gorm.io/gorm"
)

type NoteUseCase interface {
	CreateNote(note *entities.Note) error
	GetAllNote(userid uint) ([]entities.Note, error)
	UpdateColor(noteID uint, userID uint, color string) error
	UpdatePriority(noteID uint, userID uint, priority int) error
	UpdateTitleAndContent(noteID uint, userID uint, title string, content string, todoItems []entities.ToDo) error
	UpdateStatus(noteID uint, userID uint, isTodo *bool, isAllDone *bool) error
	UpdateTodoStatus(noteID uint, todoID uint, userID uint, isDone bool) error
	DeleteNoteById(noteID uint, userID uint) error
	RestoreNoteById(noteID uint, userID uint) error
	AddTagToNote(noteID uint, tagID uint, userID uint) error
	RemoveTagFromNote(noteID uint, tagID uint, userID uint) error
	GetDeletedNotes(userID uint) ([]entities.Note, error)
}

type NoteService struct {
	noteRepo         repository.NoteRepository
	shareNoteService ShareNoteUseCase // เพิ่มฟิลด์นี้
}

func NewNoteService(noteRepo repository.NoteRepository, shareNoteService ShareNoteUseCase) *NoteService {
	return &NoteService{
		noteRepo:         noteRepo,
		shareNoteService: shareNoteService,
	}
}

func (s *NoteService) CreateNote(note *entities.Note) error {
	timeCreate := time.Now().Format("2006-01-02 15:04:05")
	note.CreatedAt = timeCreate

	// คำนวณ IsAllDone จาก TodoItems
	note.IsAllDone = false
	// for _, todo := range note.TodoItems {
	// 	if !todo.IsDone {
	// 		note.IsAllDone = false
	// 		break
	// 	}
	// }
	fmt.Println("Note: ", note)

	return s.noteRepo.CreateNote(note)
}

func (s *NoteService) GetAllNote(userid uint) ([]entities.Note, error) {
	return s.noteRepo.GetAllNoteByUserId(userid)
}

func (s *NoteService) UpdateColor(noteID uint, userID uint, color string) error {
	// ตรวจสอบว่าโน้ตนั้นมีอยู่จริงไหม
	note, err := s.noteRepo.GetNoteById(noteID)
	if err != nil {
		// ระบุข้อผิดพลาดให้ชัดเจน
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("note not found")
		}
		return fmt.Errorf("failed to retrieve note: %v", err)
	}
	// ตรวจสอบสิทธิ์การแก้ไข
	isAllowed, err := s.shareNoteService.IsUserAllowedToEdit(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isAllowed {
		return fmt.Errorf("you are not authorized to update this note")
	}

	// อัปเดตสี
	err = s.noteRepo.UpdateNoteColor(noteID, note.UserID, color)
	if err != nil {
		return fmt.Errorf("failed to update note color: %v", err)
	}

	return nil
}

func (s *NoteService) UpdatePriority(noteID uint, userID uint, priority int) error {
	// ตรวจสอบว่าโน้ตนั้นมีอยู่จริง
	note, err := s.noteRepo.GetNoteById(noteID)
	if err != nil {
		return fmt.Errorf("note not found")
	}

	// ตรวจสอบสิทธิ์การแก้ไข
	isAllowed, err := s.shareNoteService.IsUserAllowedToEdit(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isAllowed {
		return fmt.Errorf("you are not authorized to update this note")
	}

	// ดำเนินการอัปเดต Priority
	return s.noteRepo.UpdateNotePriority(noteID, note.UserID, priority)
}

func (s *NoteService) UpdateTitleAndContent(noteID uint, userID uint, title string, content string, todoItems []entities.ToDo) error {
	// ตรวจสอบว่าโน้ตนั้นมีอยู่จริง
	note, err := s.noteRepo.GetNoteById(noteID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("note not found") // ไม่มีโน้ต
		}
		return fmt.Errorf("failed to retrieve note: %v", err)
	}

	// ตรวจสอบสิทธิ์การแก้ไข
	isAllowed, err := s.shareNoteService.IsUserAllowedToEdit(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isAllowed {
		return fmt.Errorf("you are not authorized to update this note") // ไม่มีสิทธิ์
	}

	// Validation: ห้ามส่ง content และ todo_items พร้อมกัน
	if len(todoItems) > 0 && content != "" {
		return fmt.Errorf("note cannot have both content and todo_items")
	}

	// อัปเดต Title หากมีการส่งค่า
	if title != "" {
		note.Title = title
	}

	// ถ้ามี Content ให้ลบ TodoItems และอัปเดต Content
	if content != "" {
		note.Content = content
		note.TodoItems = nil // ลบ TodoItems
	}

	// ถ้ามี TodoItems ให้ลบ Content และอัปเดต TodoItems
	if len(todoItems) > 0 {
		note.TodoItems = todoItems
		note.Content = "" // ลบ Content
	}

	// อัปเดต UpdatedAt
	note.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	// บันทึกการอัปเดต
	return s.noteRepo.UpdateNoteTitleAndContent(note)
}

func (s *NoteService) UpdateStatus(noteID uint, userID uint, isTodo *bool, isAllDone *bool) error {
	// ตรวจสอบว่าโน้ตนั้นมีอยู่จริง
	note, err := s.noteRepo.GetNoteById(noteID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("note not found") // ไม่มีโน้ต
		}
		return fmt.Errorf("failed to retrieve note: %v", err)
	}

	// ตรวจสอบสิทธิ์การแก้ไข
	isAllowed, err := s.shareNoteService.IsUserAllowedToEdit(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %v", err)
	}
	if !isAllowed {
		return fmt.Errorf("you are not authorized to update this note") // ไม่มีสิทธิ์
	}

	// อัปเดตสถานะใน Repository Layer
	return s.noteRepo.UpdateNoteStatus(noteID, note.UserID, isTodo, isAllDone)
}

func (s *NoteService) UpdateTodoStatus(noteID uint, todoID uint, userID uint, isDone bool) error {
	// ตรวจสอบว่า Note เป็นของ User หรือไม่
	isOwner, err := s.noteRepo.IsUserAllowedToAccessNote(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %v", err)
	}
	if !isOwner {
		return fmt.Errorf("you are not authorized to update this todo")
	}

	// อัปเดตสถานะของ Todo
	if err := s.noteRepo.UpdateTodoStatus(noteID, todoID, isDone); err != nil {
		return fmt.Errorf("failed to update todo status: %v", err)
	}

	return nil
}

func (s *NoteService) DeleteNoteById(noteID uint, userID uint) error {
	// ตรวจสอบว่า Note เป็นของ User หรือไม่
	_, err := s.noteRepo.GetNoteByIdAndUser(noteID, userID)
	if err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ดำเนินการลบโน้ต
	if err := s.noteRepo.DeleteNoteById(noteID); err != nil {
		return fmt.Errorf("failed to delete note: %v", err)
	}
	return nil
}

func (s *NoteService) RestoreNoteById(noteID uint, userID uint) error {
	// ตรวจสอบว่า Note เป็นของ User หรือไม่
	_, err := s.noteRepo.GetNoteByIdAndUser(noteID, userID)
	if err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ดำเนินการกู้คืนโน้ต
	if err := s.noteRepo.RestoreNoteById(noteID); err != nil {
		return fmt.Errorf("failed to restore note: %v", err)
	}
	return nil
}

func (s *NoteService) AddTagToNote(noteID uint, tagID uint, userID uint) error {
	return s.noteRepo.AddTagToNote(noteID, tagID, userID)
}

func (s *NoteService) RemoveTagFromNote(noteID uint, tagID uint, userID uint) error {
	return s.noteRepo.RemoveTagFromNote(noteID, tagID, userID)
}

func (s *NoteService) GetDeletedNotes(userID uint) ([]entities.Note, error) {
	return s.noteRepo.GetDeletedNotesByUserID(userID)
}

