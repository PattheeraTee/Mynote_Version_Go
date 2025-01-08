package repository

import (
	"miw/entities"
)

type ReminderRepository interface {
	AddReminder(noteID uint, reminder *entities.Reminder) error 
	GetReminderByNoteID(noteID uint) ([]entities.Reminder, error) 
	UpdateReminder(reminder *entities.Reminder) error
	DeleteReminder(reminderID uint) error 
	GetReminderByID(reminderID uint) (*entities.Reminder, error)
}
