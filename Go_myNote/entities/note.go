package entities

type Note struct {
	NoteID     uint       `json:"note_id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Color      string     `json:"color"`
	Priority   int        `json:"priority"`
	IsTodo     bool       `json:"is_todo"`
	TodoItems  []ToDo     `gorm:"foreignKey:NoteID;constraint:OnDelete:CASCADE;" json:"todo_items"` // เชื่อมโยงกับ ToDo
	IsAllDone 	bool 	  `json:"is_all_done"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
	DeletedAt  string     `json:"deleted_at"`
	Tags       []Tag      `gorm:"many2many:note_tags;joinForeignKey:NoteID;joinReferences:TagID;constraint:OnDelete:CASCADE;"`
	Reminder  []Reminder `gorm:"foreignKey:NoteID"`
	Event      Event      `gorm:"foreignKey:NoteID;constraint:OnDelete:CASCADE;"`
}

type ToDo struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    NoteID    uint   `json:"note_id"`           // เชื่อมโยงกับ Note
    Content   string `json:"content"`           // เนื้อหาของ To-Do
    IsDone    bool   `json:"is_done"`           // สถานะเสร็จสิ้นหรือไม่
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type Reminder struct {
	ReminderID   uint   `json:"reminder_id" gorm:"primaryKey"`
	NoteID       uint   `json:"note_id"`
	ReminderTime string `json:"reminder_time"`
	Recurring    bool   `json:"recurring"`
	Frequency    string `json:"frequency"`
}

type Tag struct {
	TagID   uint   `json:"tag_id" gorm:"primaryKey"`
	TagName string `json:"tag_name" gorm:"unique"`
	UserID  uint   `json:"user_id"`
    Notes   []Note `gorm:"many2many:note_tags;joinForeignKey:TagID;joinReferences:NoteID;constraint:OnDelete:CASCADE;"`
}

type ShareNote struct {
	ShareNoteID uint `json:"share_note_id" gorm:"primaryKey"`
	NoteID      uint `json:"note_id"`
	SharedWith  uint `json:"shared_with"`
}

type Event struct {
	EventID   uint   `json:"event_id" gorm:"primaryKey"`
	NoteID    uint   `json:"note_id" gorm:"unique"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
