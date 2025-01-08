package gormRepository

import (
	"fmt"
	"miw/entities"
	"errors"
	"gorm.io/gorm"
)

type GormReminderRepository struct {
	db *gorm.DB
}

func NewGormReminderRepository(db *gorm.DB) *GormReminderRepository {
	return &GormReminderRepository{db: db}
}

func (r *GormReminderRepository ) GetReminderByID(reminderID uint) (*entities.Reminder, error) {
	var reminder entities.Reminder
	if err := r.db.First(&reminder, reminderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("reminder not found")
		}
		return nil, fmt.Errorf("failed to fetch reminder: %v", err)
	}
	return &reminder, nil
}

// เพิ่ม Reminder เข้า Note
func (r *GormReminderRepository) AddReminder(noteID uint, reminder *entities.Reminder) error {
    // ตรวจสอบว่า Note มีอยู่และไม่ถูกลบ
    var note entities.Note
    if err := r.db.Where("note_id = ? AND deleted_at = ?", noteID,"").First(&note).Error; err != nil {
        return fmt.Errorf("note not found or already deleted")
    }

    // เพิ่ม Reminder
    reminder.NoteID = noteID
    if err := r.db.Create(reminder).Error; err != nil {
        return fmt.Errorf("failed to add reminder: %v", err)
    }

    return nil
}

// ดึงข้อมูล Reminder โดย Note ID
func (r *GormReminderRepository ) GetReminderByNoteID(noteID uint) ([]entities.Reminder, error) {
	var reminders []entities.Reminder
	if err := r.db.Where("note_id = ?", noteID).Find(&reminders).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch reminders: %v", err)
	}
	return reminders, nil
}

// ลบ Reminder
func (r *GormReminderRepository ) DeleteReminder(reminderID uint) error {
    // ตรวจสอบว่ามี Reminder อยู่หรือไม่
    var reminder entities.Reminder
    if err := r.db.First(&reminder, reminderID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return fmt.Errorf("reminder with ID %d not found", reminderID)
        }
        return fmt.Errorf("failed to find reminder: %v", err)
    }

    // ดำเนินการลบ Reminder
    if err := r.db.Delete(&reminder).Error; err != nil {
        return fmt.Errorf("failed to delete reminder: %v", err)
    }

    return nil
}

func (r *GormReminderRepository ) UpdateReminder(reminder *entities.Reminder) error {
	if err := r.db.Save(reminder).Error; err != nil {
		return fmt.Errorf("failed to update reminder: %v", err)
	}
	return nil
}



