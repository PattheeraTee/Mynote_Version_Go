package service

import (
	"errors"
	"fmt"
	"log"
	"miw/entities"
	"miw/usecases/repository"
	"miw/utils"
	"time"

	"gorm.io/gorm"
)

type ReminderUseCase interface {
	GetReminderByID(reminderID uint) (*entities.Reminder, error) 
	AddReminder(noteID uint, userID uint, reminder *entities.Reminder) (*entities.Reminder, error)
	GetReminderByNoteID(userID uint, noteID uint) ([]entities.Reminder, error)
	UpdateReminder(userID uint, reminderID uint, reminderTime *string, recurring *bool, frequency *string) error
	DeleteReminder(userID uint, reminderID uint) error
}

type ReminderService struct {
	reminderRepo repository.ReminderRepository
	noteRepo     repository.NoteRepository
	userRepo     repository.UserRepository
}

func NewReminderService(reminderRepo repository.ReminderRepository, noteRepo repository.NoteRepository, userRepo repository.UserRepository) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		noteRepo:     noteRepo,
		userRepo:     userRepo,
	}
}

func (s *ReminderService) GetReminderByID(reminderID uint) (*entities.Reminder, error) {
	return s.reminderRepo.GetReminderByID(reminderID)
}

func (s *ReminderService) GetReminderByNoteID(userID uint, noteID uint) ([]entities.Reminder, error) {
	// ตรวจสอบว่า Note เป็นของผู้ใช้หรือไม่
	_, err := s.noteRepo.GetNoteByIdAndUser(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("note not found or does not belong to the user")
	}

	// ดึง Reminder ที่เกี่ยวข้อง
	reminders, err := s.reminderRepo.GetReminderByNoteID(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reminders: %v", err)
	}

	return reminders, nil
}

func (s *ReminderService) AddReminder(noteID uint, userID uint, reminder *entities.Reminder) (*entities.Reminder, error) {
	// ตรวจสอบว่า Note ID มีอยู่ในระบบและเป็นของผู้ใช้หรือไม่
	note, err := s.noteRepo.GetNoteByIdAndUser(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("note not found or does not belong to the user: %v", err)
	}

	// ตรวจสอบว่า Note มี Reminder อยู่แล้วหรือไม่
	existingReminders, err := s.reminderRepo.GetReminderByNoteID(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing reminders: %v", err)
	}

	if len(existingReminders) > 0 {
		return nil, fmt.Errorf("a reminder already exists for this note")
	}

	// ตรวจสอบเวลา Reminder
	thLocation, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return nil, fmt.Errorf("failed to load Thailand timezone: %v", err)
	}

	reminderTime, err := time.ParseInLocation("2006-01-02 15:04:05", reminder.ReminderTime, thLocation)
	if err != nil {
		return nil, fmt.Errorf("invalid reminder time format: %v", err)
	}

	if reminderTime.Before(time.Now().In(thLocation)) {
		return nil, fmt.Errorf("reminder time is in the past and cannot be added")
	}

	// บันทึก Reminder ลงฐานข้อมูล
	if err := s.reminderRepo.AddReminder(noteID, reminder); err != nil {
		return nil, fmt.Errorf("failed to add reminder to database: %v", err)
	}

	// ตั้งค่าแจ้งเตือน
	s.scheduleReminder(note, reminder, reminderTime)

	// คืนค่า Reminder ที่สร้างใหม่
	return reminder, nil
}

func (s *ReminderService) UpdateReminder(userID uint, reminderID uint, reminderTime *string, recurring *bool, frequency *string) error {
	// ตรวจสอบว่า Reminder มีอยู่จริงและเป็นของผู้ใช้งานนี้หรือไม่
	existingReminder, err := s.reminderRepo.GetReminderByID(reminderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("reminder with ID %d not found", reminderID)
		}
		return fmt.Errorf("failed to find reminder: %v", err)
	}

	// ตรวจสอบว่า Note เป็นของผู้ใช้คนนี้หรือไม่
	_, err = s.noteRepo.GetNoteByIdAndUser(existingReminder.NoteID, userID)
	if err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ตรวจสอบเวลาที่ส่งมา
	if reminderTime != nil {
		thLocation, _ := time.LoadLocation("Asia/Bangkok")
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", *reminderTime, thLocation)
		if err != nil {
			return fmt.Errorf("invalid reminder time format: %v", err)
		}
		if parsedTime.Before(time.Now().In(thLocation)) {
			return fmt.Errorf("reminder time cannot be in the past")
		}
		existingReminder.ReminderTime = *reminderTime
	}

	// อัปเดตค่าที่ส่งมา
	if recurring != nil {
		existingReminder.Recurring = *recurring
	}
	if frequency != nil {
		existingReminder.Frequency = *frequency
	}

	// บันทึกการเปลี่ยนแปลง
	if err := s.reminderRepo.UpdateReminder(existingReminder); err != nil {
		return fmt.Errorf("failed to update reminder: %v", err)
	}

	// ตั้งค่าแจ้งเตือนใหม่
	note, err := s.noteRepo.GetNoteById(existingReminder.NoteID)
	if err != nil {
		return fmt.Errorf("failed to fetch note: %v", err)
	}

	thLocation, _ := time.LoadLocation("Asia/Bangkok")
	parsedTime, _ := time.ParseInLocation("2006-01-02 15:04:05", existingReminder.ReminderTime, thLocation)
	s.scheduleReminder(note, existingReminder, parsedTime)

	return nil
}

func (s *ReminderService) scheduleReminder(note *entities.Note, reminder *entities.Reminder, reminderTime time.Time) {
	go func() {
		durationUntilReminder := time.Until(reminderTime)
		time.AfterFunc(durationUntilReminder, func() {
			s.sendReminder(note, reminder)

			if reminder.Recurring {
				s.scheduleRecurringReminder(note, reminder, reminderTime)
			}
		})
	}()
}

func (s *ReminderService) scheduleRecurringReminder(note *entities.Note, reminder *entities.Reminder, reminderTime time.Time) {
	thLocation, _ := time.LoadLocation("Asia/Bangkok")
	nextTime := reminderTime

	switch reminder.Frequency {
	case "daily":
		nextTime = reminderTime.AddDate(0, 0, 1)
	case "weekly":
		nextTime = reminderTime.AddDate(0, 0, 7)
	case "monthly":
		nextTime = reminderTime.AddDate(0, 1, 0)
	case "yearly":
		nextTime = reminderTime.AddDate(1, 0, 0)
	}

	if nextTime.After(time.Now().In(thLocation)) {
		s.scheduleReminder(note, reminder, nextTime)
	}
}

func (s *ReminderService) sendReminder(note *entities.Note, reminder *entities.Reminder) {
	userEmail, err := s.userRepo.GetUserEmailByID(note.UserID)
	if err != nil {
		log.Printf("Failed to get user email: %v", err)
		return
	}

	emailBody := "Reminder\n\n"
	emailBody += fmt.Sprintf("Title: %s\n", note.Title)

	if note.Content != "" {
		emailBody += fmt.Sprintf("Content: %s\n", note.Content)
	}

	if len(note.TodoItems) > 0 {
		emailBody += "Todo Items:\n"
		for _, todo := range note.TodoItems {
			status := "Not Done"
			if todo.IsDone {
				status = "Done"
			}
			emailBody += fmt.Sprintf("- %s [%s]\n", todo.Content, status)
		}
	}

	emailBody += fmt.Sprintf("\nReminder Time: %s\n", reminder.ReminderTime)

	err = utils.SendEmail(userEmail, "Reminder Notification", emailBody)
	if err != nil {
		log.Printf("Failed to send reminder email: %v", err)
	} else {
		log.Printf("Reminder email sent to %s\n", userEmail)
	}
}

func (s *ReminderService) DeleteReminder(userID uint, reminderID uint) error {
	existingReminder, err := s.reminderRepo.GetReminderByID(reminderID)
	// fmt.Println(existingReminder)
	// fmt.Println("ReminderID",reminderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("reminder with ID %d not found", reminderID)
		}
		return fmt.Errorf("failed to find reminder: %v", err)
	}

	// ตรวจสอบว่า Note เป็นของผู้ใช้หรือไม่
	_, err = s.noteRepo.GetNoteByIdAndUser(existingReminder.NoteID, userID)
	if err != nil {
		return fmt.Errorf("note not found or does not belong to the user")
	}

	// ลบ Reminder
	return s.reminderRepo.DeleteReminder(reminderID)
}
