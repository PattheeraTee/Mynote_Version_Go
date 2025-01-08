package repository

import "miw/entities"

type ShareNoteRepository interface {
	GetUserByEmail(email string) (*entities.User, error)
	ShareNoteWithUser(noteID, sharedWith uint) error
	IsNoteSharedWithUser(noteID, userID uint) (bool, error)
	IsUserAllowedToEdit(noteID uint, userID uint) (bool, error)
	ShareNoteWithEmail(noteID uint, ownerID uint, email string) ([]map[string]string, error) 
	RemoveShareByEmail(noteID uint, ownerID uint, email string) error
	GetSharedEmailsByNoteID(noteID uint) ([]map[string]string, error)
}
