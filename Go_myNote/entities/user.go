package entities


type User struct {
	UserID              uint    `json:"user_id" gorm:"primaryKey;autoIncrement"`
	Username            string  `json:"username"`
	Email               string  `json:"email" gorm:"unique"`
	Password            string  `json:"password"`
	GoogleCalendarToken string  `json:"google_calendar_token"`
	Notes               []Note  `gorm:"foreignKey:UserID"`
	SharedNotes         []ShareNote `gorm:"foreignKey:SharedWith"`
}

