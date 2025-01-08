package gormRepository

import (
	"fmt"
	"miw/entities"
	"time"

	"gorm.io/gorm"
)

type GormNoteRepository struct {
	db *gorm.DB
}

func NewGormNoteRepository(db *gorm.DB) *GormNoteRepository {
	return &GormNoteRepository{db: db}
}

func (r *GormNoteRepository) CreateNote(note *entities.Note) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        fmt.Println("Starting Transaction")

        // สร้าง Note
        session := tx.Session(&gorm.Session{FullSaveAssociations: false})
        if err := session.Create(note).Error; err != nil {
            return fmt.Errorf("failed to create note: %v", err)
        }

        // จัดการ TodoItems
        if len(note.TodoItems) > 0 {
            todoItems := make([]entities.ToDo, 0)
            for _, todo := range note.TodoItems {
                todo.NoteID = note.NoteID
                todo.ID = 0 // รีเซ็ต ID

                // ตรวจสอบว่ารายการไม่มีซ้ำ
                var count int64
                err := tx.Model(&entities.ToDo{}).Where("note_id = ? AND content = ?", todo.NoteID, todo.Content).Count(&count).Error
                if err != nil {
                    return fmt.Errorf("failed to check for duplicate TodoItem: %v", err)
                }
                if count == 0 {
                    todoItems = append(todoItems, todo)
                } else {
                    fmt.Printf("Duplicate ToDo found for Content: %s\n", todo.Content)
                }
            }

            // บันทึก TodoItems
            if len(todoItems) > 0 {
                if err := tx.Create(&todoItems).Error; err != nil {
                    return fmt.Errorf("failed to create todo items: %v", err)
                }
            }
        }

        fmt.Println("Transaction Completed")
        return nil
    })
}


func (r *GormNoteRepository) GetAllNoteByUserId(userID uint) ([]entities.Note, error) {
	var notes []entities.Note

	// Fetch notes owned by the user
	if err := r.db.Where("user_id = ? AND deleted_at = ?", userID, "").Preload("Tags").Preload("Reminder").Preload("TodoItems").Find(&notes).Error; err != nil {
		return nil, err
	}

	// Fetch notes shared with the user
	var sharedNoteIDs []uint
	if err := r.db.Model(&entities.ShareNote{}).Where("shared_with = ?", userID).Pluck("note_id", &sharedNoteIDs).Error; err != nil {
		return nil, err
	}
	if len(sharedNoteIDs) > 0 {
		var sharedNotes []entities.Note
		if err := r.db.Where("note_id IN ? AND deleted_at = ?", sharedNoteIDs, "").Preload("Tags").Preload("Reminder").Preload("TodoItems").Find(&sharedNotes).Error; err != nil {
			return nil, err
		}
		notes = append(notes, sharedNotes...)
	}

	return notes, nil
}

func (r *GormNoteRepository) GetNoteById(noteID uint) (*entities.Note, error) {
	var note entities.Note
	if err := r.db.Where("note_id = ?", noteID).
		Preload("Tags").
		Preload("Reminder").
		Preload("Event").
		Preload("TodoItems"). // เพิ่มการโหลด TodoItems
		First(&note, noteID).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *GormNoteRepository) UpdateNoteColor(noteID uint, userID uint, color string) error {
	result := r.db.Model(&entities.Note{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]interface{}{
			"color":      color,
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})
	// fmt.Println(noteID, userID, color)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("note not found or user not authorized")
	}

	return nil
}

func (r *GormNoteRepository) UpdateNotePriority(noteID uint, userID uint, priority int) error {
	result := r.db.Model(&entities.Note{}).
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Updates(map[string]interface{}{
			"priority":   priority,
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("note not found or user not authorized")
	}

	return nil

}

func (r *GormNoteRepository) UpdateNoteTitleAndContent(note *entities.Note) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// อัปเดต Note
		if err := tx.Save(note).Error; err != nil {
			return fmt.Errorf("failed to update note: %v", err)
		}

		// ถ้ามี TodoItems ให้จัดการ
		if len(note.TodoItems) > 0 {
			// ลบ TodoItems เก่าที่เชื่อมโยงกับ NoteID
			if err := tx.Where("note_id = ?", note.NoteID).Delete(&entities.ToDo{}).Error; err != nil {
				return fmt.Errorf("failed to delete old todo items: %v", err)
			}

			// ตั้งค่า NoteID และรีเซ็ต ID เป็น 0 สำหรับการเพิ่มใหม่
			for i := range note.TodoItems {
				note.TodoItems[i].ID = 0
				note.TodoItems[i].NoteID = note.NoteID
			}

			// เพิ่ม TodoItems ใหม่
			if err := tx.Create(&note.TodoItems).Error; err != nil {
				return fmt.Errorf("failed to create new todo items: %v", err)
			}
		}

		// หากไม่มี TodoItems แต่มี Content ให้ลบ TodoItems เก่า
		if len(note.TodoItems) == 0 && note.Content != "" {
			if err := tx.Where("note_id = ?", note.NoteID).Delete(&entities.ToDo{}).Error; err != nil {
				return fmt.Errorf("failed to delete old todo items: %v", err)
			}
		}

		return nil
	})
}

func (r *GormNoteRepository) UpdateNoteStatus(noteID uint, userID uint, isTodo *bool, isAllDone *bool) error {
	updates := map[string]interface{}{}

	// กำหนดค่า is_todo และ is_all_done
	if isTodo != nil {
		updates["is_todo"] = *isTodo
	}
	if isAllDone != nil {
		updates["is_all_done"] = *isAllDone
	}

	// เพิ่ม updated_at
	updates["updated_at"] = time.Now().Format("2006-01-02 15:04:05")

	// อัปเดตโน้ต
	result := r.db.Model(&entities.Note{}).Where("note_id = ?", noteID).Updates(updates)

	// ตรวจสอบผลลัพธ์
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("note not found or user not authorized")
	}

	return nil
}

func (r *GormNoteRepository) UpdateTodoStatus(noteID uint, todoID uint, isDone bool) error {
    result := r.db.Model(&entities.ToDo{}).
        Where("id = ? AND note_id = ?", todoID, noteID).
        Updates(map[string]interface{}{
            "is_done":    isDone,
            "updated_at": time.Now().Format("2006-01-02 15:04:05"),
        })

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return fmt.Errorf("todo not found or does not belong to the note")
    }

    return nil
}

func (r *GormNoteRepository) DeleteNoteById(noteID uint) error {
	// ใช้เวลาปัจจุบันในรูปแบบ string
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// อัปเดตฟิลด์ DeletedAt ด้วยเวลาปัจจุบัน
	result := r.db.Model(&entities.Note{}).Where("note_id = ? AND deleted_at = ?", noteID, "").Update("deleted_at", currentTime)

	// ตรวจสอบว่าพบโน้ตหรือไม่
	if result.RowsAffected == 0 {
		return fmt.Errorf("note with ID %d not found or already deleted", noteID)
	}

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete note with ID %d: %v", noteID, result.Error)
	}

	return nil
}

func (r *GormNoteRepository) RestoreNoteById(noteID uint) error {
	// ตรวจสอบว่ามี Note ที่ตรงกับ ID หรือไม่
	var note entities.Note
	if err := r.db.Unscoped().Where("note_id = ?", noteID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("note with ID %d not found", noteID)
		}
		return fmt.Errorf("failed to check note with ID %d: %v", noteID, err)
	}

	// ใช้คำสั่ง Unscoped() เพื่ออัปเดต DeletedAt ให้เป็น nil
	if err := r.db.Unscoped().Model(&entities.Note{}).Where("note_id = ?", noteID).Update("deleted_at", "").Error; err != nil {
		return fmt.Errorf("failed to restore note with ID %d: %v", noteID, err)
	}

	return nil
}

func (r *GormNoteRepository) AddTagToNote(noteID uint, tagID uint, userID uint) error {
	// ตรวจสอบว่า Note เป็นของ User หรือไม่
	var note entities.Note
	if err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ตรวจสอบว่า Tag เป็นของ User หรือไม่
	var tag entities.Tag
	if err := r.db.Where("tag_id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return fmt.Errorf("tag not found or does not belong to the user")
	}

	// เพิ่ม Tag เข้า Note
	if err := r.db.Model(&note).Association("Tags").Append(&tag); err != nil {
		return fmt.Errorf("failed to add tag to note: %v", err)
	}

	return nil
}

func (r *GormNoteRepository) RemoveTagFromNote(noteID uint, tagID uint, userID uint) error {
	// ตรวจสอบว่า Note เป็นของ User หรือไม่
	var note entities.Note
	if err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ตรวจสอบว่า Tag เป็นของ User หรือไม่
	var tag entities.Tag
	if err := r.db.Where("tag_id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		return fmt.Errorf("tag not found or does not belong to the user")
	}

	// ลบ Tag ออกจาก Note
	if err := r.db.Model(&note).Association("Tags").Delete(&tag); err != nil {
		return fmt.Errorf("failed to remove tag from note: %v", err)
	}

	return nil
}

func (r *GormNoteRepository) GetNoteByIdAndUser(noteID uint, userID uint) (*entities.Note, error) {
	var note entities.Note
	if err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).
		Preload("Tags").
		Preload("Reminder").
		Preload("Event").
		Preload("TodoItems").
		First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("note not found or does not belong to the user")
		}
		return nil, err
	}
	return &note, nil
}

func (r *GormNoteRepository) IsNoteOwnedByUser(noteID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entities.Note{}).Where("id = ? AND user_id = ?", noteID, userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormNoteRepository) IsUserAllowedToAccessNote(noteID uint, userID uint) (bool, error) {
	// ตรวจสอบว่า User เป็นเจ้าของ Note หรือไม่
	var count int64
	err := r.db.Model(&entities.Note{}).Where("note_id = ? AND user_id = ?", noteID, userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// ตรวจสอบว่า Note ถูกแชร์ให้ User หรือไม่
	err = r.db.Model(&entities.ShareNote{}).Where("note_id = ? AND shared_with = ?", noteID, userID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *GormNoteRepository) GetDeletedNotesByUserID(userID uint) ([]entities.Note, error) {
	var notes []entities.Note
	// ดึงโน้ตที่ถูกลบเท่านั้น
	if err := r.db.Where("user_id = ? AND deleted_at != ?", userID, "").
    Preload("Tags").
    Preload("Reminder").
    Preload("TodoItems").
    Find(&notes).Error; err != nil {
    return nil, err
}

	return notes, nil
}
