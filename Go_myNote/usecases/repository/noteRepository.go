package repository

import (
	"miw/entities"
)

type NoteRepository interface {
	CreateNote(note *entities.Note) error
	GetAllNoteByUserId(userID uint) ([]entities.Note, error)
	GetNoteById(noteID uint) (*entities.Note, error)
	UpdateNoteColor(noteID uint, userID uint, color string) error 
	UpdateNotePriority(noteID uint, userID uint, priority int) error 
	UpdateNoteTitleAndContent(note *entities.Note) error 
	UpdateNoteStatus(noteID uint, userID uint, isTodo *bool, isAllDone *bool) error
	UpdateTodoStatus(noteID uint, todoID uint, isDone bool) error
	DeleteNoteById(noteID uint) error
	RestoreNoteById(noteID uint) error 
	AddTagToNote(noteID uint, tagID uint, userID uint) error
	RemoveTagFromNote(noteID uint, tagID uint, userID uint) error
	GetNoteByIdAndUser(noteID uint, userID uint) (*entities.Note, error)
	IsNoteOwnedByUser(noteID uint, userID uint) (bool, error)
	IsUserAllowedToAccessNote(noteID uint, userID uint) (bool, error)
	GetDeletedNotesByUserID(userID uint) ([]entities.Note, error)
}
