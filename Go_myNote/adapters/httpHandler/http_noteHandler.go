package httpHandler

import (
	"fmt"
	"miw/entities"
	"miw/usecases/service"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type NoteResponse struct {
	NoteID    uint                `json:"note_id"`
	UserID    uint                `json:"user_id"`
	Title     string              `json:"title"`
	Content   string              `json:"content,omitempty"` // ซ่อนถ้าไม่มีค่า
	Color     string              `json:"color"`
	Priority  int                 `json:"priority"`
	IsTodo    bool                `json:"is_todo"`
	IsAllDone bool                `json:"is_all_done"` // เพิ่มฟิลด์นี้
	TodoItems []ToDoResponse      `json:"todo_items"`  // เพิ่มรายการ ToDo
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	DeletedAt string              `json:"deleted_at,omitempty"` // ซ่อนถ้าไม่มีค่า
	Tags	  []NoteTagResponse   `json:"tags"`
	Reminder  []entities.Reminder `json:"reminder"`
	Event     interface{}         `json:"event"`
}
type NoteTagResponse struct {
	TagID   uint   `json:"tag_id"`
	TagName string `json:"tag_name"`
}

type ReminderResponse struct {
	ReminderID uint   `json:"reminder_id"`
	NoteID     uint   `json:"note_id"`
	Content    string `json:"content"`
	DateTime   string `json:"datetime"`
}

type ToDoResponse struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	IsDone  bool   `json:"is_done"`
}

type HttpNoteHandler struct {
	noteUseCase service.NoteUseCase
}

func NewHttpNoteHandler(useCase service.NoteUseCase) *HttpNoteHandler {
	return &HttpNoteHandler{noteUseCase: useCase}
}

func (h *HttpNoteHandler) CreateNoteHandler(c *fiber.Ctx) error {
	note := new(entities.Note)

	// รับข้อมูลโน้ตจาก Body ของ request
	if err := c.BodyParser(note); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	// Validation: ถ้ามี `todo_items` ต้องไม่มี `content` และถ้าไม่มี `todo_items` ต้องมี `content`
	if len(note.TodoItems) > 0 && note.Content != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Note cannot have both content and todo_items"})
	}
	if len(note.TodoItems) == 0 && note.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Note must have either content or todo_items"})
	}

	// ดึง UserID จาก Context (Middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	note.UserID = userID

	// เรียกใช้ฟังก์ชันสร้างโน้ต
	if err := h.noteUseCase.CreateNote(note); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not create note")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Note created successfully",
		"note":    note,
	})
}


func (h *HttpNoteHandler) GetAllNoteHandler(c *fiber.Ctx) error {
	// ดึง UserID จาก Context (Middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// ดึงข้อมูลโน้ตทั้งหมดของ User
	notes, err := h.noteUseCase.GetAllNote(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Notes not found for this user")
	}

	// แปลงผลลัพธ์เป็น JSON Response
	var response []NoteResponse
	for _, note := range notes {
		// แปลง Tags จาก entities.Tag เป็น NoteTagResponse
		var tagResponses []NoteTagResponse
		for _, tag := range note.Tags {
			tagResponses = append(tagResponses, NoteTagResponse{
				TagID:   tag.TagID,
				TagName: tag.TagName,
			})
		}

		// แปลง TodoItems จาก entities.ToDo เป็น ToDoResponse
		var todoResponses []ToDoResponse
		for _, todo := range note.TodoItems {
			todoResponses = append(todoResponses, ToDoResponse{
				ID:      todo.ID,
				Content: todo.Content,
				IsDone:  todo.IsDone,
			})
		}

		response = append(response, NoteResponse{
			NoteID:    note.NoteID,
			UserID:    note.UserID,
			Title:     note.Title,
			Content:   note.Content,
			Color:     note.Color,
			Priority:  note.Priority,
			IsTodo:    note.IsTodo,
			IsAllDone: note.IsAllDone,
			TodoItems: todoResponses,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			DeletedAt: note.DeletedAt,
			Tags:      tagResponses, // เปลี่ยนจาก tag เป็น tagResponses
			Reminder:  note.Reminder,
			Event:     note.Event,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"notes": response,
	})
}


func (h *HttpNoteHandler) UpdateColorHandler(c *fiber.Ctx) error {
	noteID, _ := strconv.Atoi(c.Params("noteid"))
	userID, _ := c.Locals("user_id").(uint)

	data := new(struct {
		Color string `json:"color"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.noteUseCase.UpdateColor(uint(noteID), userID, data.Color)
	if err != nil {
		// ตรวจสอบข้อผิดพลาดและแสดงข้อความที่เหมาะสม
		switch {
		case err.Error() == "note not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
		case err.Error() == "you are not authorized to update this note":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this note"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update color"})
		}
	}

	return c.JSON(fiber.Map{"message": "Color updated successfully"})
}


func (h *HttpNoteHandler) UpdatePriorityHandler(c *fiber.Ctx) error {
	noteID, _ := strconv.Atoi(c.Params("noteid"))
	userID, _ := c.Locals("user_id").(uint)

	data := new(struct {
		Priority int `json:"priority"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.noteUseCase.UpdatePriority(uint(noteID), userID, data.Priority)
	if err != nil {
		switch {
		case err.Error() == "note not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
		case err.Error() == "you are not authorized to update this note":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this note"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update priority"})
		}
	}

	return c.JSON(fiber.Map{"message": "Priority updated successfully"})
}


func (h *HttpNoteHandler) UpdateTitleAndContentHandler(c *fiber.Ctx) error {
	noteID, _ := strconv.Atoi(c.Params("noteid"))
	userID, _ := c.Locals("user_id").(uint)

	data := new(struct {
		Title     string          `json:"title"`
		Content   string          `json:"content"`
		TodoItems []entities.ToDo `json:"todo_items"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validation: ห้ามส่ง content และ todo_items พร้อมกัน
	if len(data.TodoItems) > 0 && data.Content != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Note cannot have both content and todo_items"})
	}

	// เรียก UseCase เพื่ออัปเดตโน้ต
	err := h.noteUseCase.UpdateTitleAndContent(uint(noteID), userID, data.Title, data.Content, data.TodoItems)
	if err != nil {
		switch err.Error() {
		case "note not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
		case "you are not authorized to update this note":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this note"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update title/content"})
		}
	}

	// ดึงข้อมูลโน้ตที่อัปเดตแล้ว
	notes, err := h.noteUseCase.GetAllNote(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve updated note"})
	}

	// แปลงผลลัพธ์เป็น JSON Response
	var response []NoteResponse
	for _, note := range notes {
		// แปลง Tags จาก entities.Tag เป็น NoteTagResponse
		var tagResponses []NoteTagResponse
		for _, tag := range note.Tags {
			tagResponses = append(tagResponses, NoteTagResponse{
				TagID:   tag.TagID,
				TagName: tag.TagName,
			})
		}

		var todoResponses []ToDoResponse
		for _, todo := range note.TodoItems {
			todoResponses = append(todoResponses, ToDoResponse{
				ID:      todo.ID,
				Content: todo.Content,
				IsDone:  todo.IsDone,
			})
		}

		response = append(response, NoteResponse{
			NoteID:    note.NoteID,
			UserID:    note.UserID,
			Title:     note.Title,
			Content:   note.Content,
			Color:     note.Color,
			Priority:  note.Priority,
			IsTodo:    note.IsTodo,
			IsAllDone: note.IsAllDone,
			TodoItems: todoResponses,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			DeletedAt: note.DeletedAt,
			Tags:      tagResponses,
			Reminder:  note.Reminder,
			Event:     note.Event,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Title/Content updated successfully",
		"notes":   response,
	})
}



func (h *HttpNoteHandler) UpdateStatusHandler(c *fiber.Ctx) error {
	noteID, _ := strconv.Atoi(c.Params("noteid"))
	userID, _ := c.Locals("user_id").(uint)

	data := new(struct {
		IsTodo    *bool `json:"is_todo"`
		IsAllDone *bool `json:"is_all_done"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validation: ต้องส่งอย่างน้อยหนึ่งค่า
	if data.IsTodo == nil && data.IsAllDone == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one of 'is_todo' or 'is_all_done' must be provided"})
	}

	// เรียก Use Case
	err := h.noteUseCase.UpdateStatus(uint(noteID), userID, data.IsTodo, data.IsAllDone)
	if err != nil {
		switch err.Error() {
		case "note not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
		case "you are not authorized to update this note":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this note"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
		}
	}

	return c.JSON(fiber.Map{"message": "Status updated successfully"})
}

func (h *HttpNoteHandler) AddTagToNoteHandler(c *fiber.Ctx) error {
	var request struct {
		NoteID uint `json:"note_id"`
		TagID  uint `json:"tag_id"`
	}

	// รับข้อมูลจาก Body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ดึง UserID จาก Context
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เรียกใช้ Service Layer
	if err := h.noteUseCase.AddTagToNote(request.NoteID, request.TagID, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tag added successfully"})
}

func (h *HttpNoteHandler) RemoveTagFromNoteHandler(c *fiber.Ctx) error {
	var request struct {
		NoteID uint `json:"note_id"`
		TagID  uint `json:"tag_id"`
	}

	// Parse the request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ดึง UserID จาก Context
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Call the use case to remove the tag from the note
	if err := h.noteUseCase.RemoveTagFromNote(request.NoteID, request.TagID, userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to remove tag %d from note %d: %v", request.TagID, request.NoteID, err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tag removed successfully"})
}

func (h *HttpNoteHandler) UpdateTodoStatusHandler(c *fiber.Ctx) error {
    noteID, err := strconv.Atoi(c.Params("noteid"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note ID"})
    }

    todoID, err := strconv.Atoi(c.Params("todoid"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid todo ID"})
    }

    userID, ok := c.Locals("user_id").(uint)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
    }

    data := new(struct {
        IsDone bool `json:"is_done"`
    })

    if err := c.BodyParser(data); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    if err := h.noteUseCase.UpdateTodoStatus(uint(noteID), uint(todoID), userID, data.IsDone); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "Todo status updated successfully"})
}

func (h *HttpNoteHandler) DeleteNoteHandler(c *fiber.Ctx) error {
	// ดึง Note ID จากพารามิเตอร์
	noteID, err := strconv.Atoi(c.Params("noteid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note ID"})
	}

	// ดึง UserID จาก Context
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เรียกใช้ Use Case เพื่อลบโน้ต
	if err := h.noteUseCase.DeleteNoteById(uint(noteID), userID); err != nil {
		if err.Error() == "note not found or does not belong to the user" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to delete this note"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete note with ID %d: %v", noteID, err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Note deleted successfully"})
}

func (h *HttpNoteHandler) RestoreNoteHandler(c *fiber.Ctx) error {
	// ดึง Note ID จากพารามิเตอร์
	noteID, err := strconv.Atoi(c.Params("noteid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note ID"})
	}

	// ดึง UserID จาก Context
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เรียกใช้ Use Case เพื่อกู้คืนโน้ต
	if err := h.noteUseCase.RestoreNoteById(uint(noteID), userID); err != nil {
		if err.Error() == "note not found or does not belong to the user" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to restore this note"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to restore note with ID %d: %v", noteID, err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Note restored successfully"})
}

func (h *HttpNoteHandler) GetDeletedNotesHandler(c *fiber.Ctx) error {
	// ดึง UserID จาก Route พารามิเตอร์
	userID, err := strconv.Atoi(c.Params("userid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// ดึงข้อมูลโน้ตที่ถูกลบ
	notes, err := h.noteUseCase.GetDeletedNotes(uint(userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve deleted notes"})
	}

	// แปลงผลลัพธ์เป็น JSON Response
	var response []NoteResponse
	for _, note := range notes {
		var tagResponses []NoteTagResponse
		for _, tag := range note.Tags {
			tagResponses = append(tagResponses, NoteTagResponse{
				TagID:   tag.TagID,
				TagName: tag.TagName,
			})
		}

		var todoResponses []ToDoResponse
		for _, todo := range note.TodoItems {
			todoResponses = append(todoResponses, ToDoResponse{
				ID:      todo.ID,
				Content: todo.Content,
				IsDone:  todo.IsDone,
			})
		}

		response = append(response, NoteResponse{
			NoteID:    note.NoteID,
			UserID:    note.UserID,
			Title:     note.Title,
			Content:   note.Content,
			Color:     note.Color,
			Priority:  note.Priority,
			IsTodo:    note.IsTodo,
			IsAllDone: note.IsAllDone,
			TodoItems: todoResponses,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
			DeletedAt: note.DeletedAt,
			Tags:      tagResponses,
			Reminder:  note.Reminder,
			Event:     note.Event,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"deleted_notes": response,
	})
}


