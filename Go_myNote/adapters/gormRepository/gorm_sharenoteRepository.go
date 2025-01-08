package gormRepository

import (
    "fmt"
    "gorm.io/gorm"
    "miw/entities"
)

type GormShareNoteRepository struct {
    db *gorm.DB
}

func NewGormShareNoteRepository(db *gorm.DB) *GormShareNoteRepository {
    return &GormShareNoteRepository{db: db}
}

// Check if the email exists in the system
func (r *GormShareNoteRepository) GetUserByEmail(email string) (*entities.User, error) {
    var user entities.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("email not found")
        }
        return nil, err
    }
    return &user, nil
}

// Share a note with a user
func (r *GormShareNoteRepository) ShareNoteWithUser(noteID, sharedWith uint) error {
    share := entities.ShareNote{
        NoteID:     noteID,
        SharedWith: sharedWith,
    }
    if err := r.db.Create(&share).Error; err != nil {
        return fmt.Errorf("failed to share note: %v", err)
    }
    return nil
}

// Check if a note is shared with a user
func (r *GormShareNoteRepository) IsNoteSharedWithUser(noteID, userID uint) (bool, error) {
    var count int64
    if err := r.db.Model(&entities.ShareNote{}).Where("note_id = ? AND shared_with = ?", noteID, userID).Count(&count).Error; err != nil {
        return false, fmt.Errorf("failed to check shared status: %v", err)
    }
    return count > 0, nil
}


func (r *GormShareNoteRepository) IsUserAllowedToEdit(noteID uint, userID uint) (bool, error) {
    var count int64
    err := r.db.Model(&entities.Note{}).Where("note_id = ? AND user_id = ?", noteID, userID).Count(&count).Error
    if err != nil {
        return false, err
    }
    if count > 0 {
        return true, nil
    }

    err = r.db.Model(&entities.ShareNote{}).Where("note_id = ? AND shared_with = ?", noteID, userID).Count(&count).Error
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

func (r *GormShareNoteRepository) ShareNoteWithEmail(noteID uint, ownerID uint, email string) ([]map[string]string, error) {
    // ตรวจสอบว่า Note เป็นของ Owner
    var note entities.Note
    if err := r.db.Where("note_id = ? AND user_id = ?", noteID, ownerID).First(&note).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("note not found or does not belong to the owner")
        }
        return nil, err
    }

    // ดึงข้อมูล User จาก Email
    var user entities.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("email not found")
        }
        return nil, err
    }

    // แชร์ Note
    share := entities.ShareNote{
        NoteID:     noteID,
        SharedWith: user.UserID,
    }
    if err := r.db.Create(&share).Error; err != nil {
        return nil, fmt.Errorf("failed to share note: %v", err)
    }

    // ดึงอีเมลที่แชร์ทั้งหมด
    var sharedUsers []struct {
        Email string
    }
    if err := r.db.Table("users").
        Select("users.email").
        Joins("JOIN share_notes ON users.user_id = share_notes.shared_with").
        Where("share_notes.note_id = ?", noteID).
        Find(&sharedUsers).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch shared emails: %v", err)
    }

    // เตรียมข้อมูลอีเมลที่แชร์
    sharedEmails := make([]map[string]string, 0, len(sharedUsers))
    for _, user := range sharedUsers {
        sharedEmails = append(sharedEmails, map[string]string{"email": user.Email, "type": "shared"})
    }

    return sharedEmails, nil
}


func (r *GormShareNoteRepository) RemoveShareByEmail(noteID uint, ownerID uint, email string) error {
    // ตรวจสอบว่า Note เป็นของ Owner
    var note entities.Note
    if err := r.db.Where("note_id = ? AND user_id = ?", noteID, ownerID).First(&note).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return fmt.Errorf("note not found or does not belong to the owner")
        }
        return err
    }

    var user entities.User
    if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return fmt.Errorf("email not found")
        }
        return err
    }

    // ลบการแชร์
    if err := r.db.Where("note_id = ? AND shared_with = ?", noteID, user.UserID).Delete(&entities.ShareNote{}).Error; err != nil {
        return fmt.Errorf("failed to remove share: %v", err)
    }
    return nil
}


func (r *GormShareNoteRepository) GetSharedEmailsByNoteID(noteID uint) ([]map[string]string, error) {
    var sharedEmails []map[string]string

    // ดึงข้อมูลอีเมลของเจ้าของโน้ต
    var ownerEmail string
    err := r.db.Table("users").
        Select("users.email").
        Joins("JOIN notes ON users.user_id = notes.user_id").
        Where("notes.note_id = ?", noteID).
        Row().
        Scan(&ownerEmail)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch owner email: %v", err)
    }

    // ดึงข้อมูลอีเมลของผู้ที่แชร์โน้ตด้วย
    var sharedUsers []struct {
        Email string
    }
    err = r.db.Table("users").
        Select("users.email").
        Joins("JOIN share_notes ON users.user_id = share_notes.shared_with").
        Where("share_notes.note_id = ?", noteID).
        Find(&sharedUsers).Error
    if err != nil {
        return nil, fmt.Errorf("failed to fetch shared emails: %v", err)
    }

    // รวมอีเมลเจ้าของพร้อมระบุว่าเป็น "owner"
    sharedEmails = append(sharedEmails, map[string]string{"email": ownerEmail, "type": "owner"})

    // รวมอีเมลของผู้ใช้ที่แชร์
    for _, user := range sharedUsers {
        sharedEmails = append(sharedEmails, map[string]string{"email": user.Email, "type": "shared"})
    }

    return sharedEmails, nil
}


