"use client";

import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useRouter } from "next/navigation";
import { faBars, faPlus, faTrash, faStar, faClock, faUserFriends, faCalendar, faTag, faBell } from "@fortawesome/free-solid-svg-icons";
import axios from "axios";
import Swal from "sweetalert2";
import { useSearchParams } from "next/navigation";

interface User {
  id: string;
  username: string;
  email: string;
}

interface TodoItem {
  id: number;
  content: string;
  is_done: boolean;
}

interface Note {
  note_id: number;
  user_id: number;
  title: string;
  content: string | null;
  color: string;
  is_todo: boolean;
  is_all_done: boolean;
  todo_items: TodoItem[] | null;
  priority: number;
  created_at: string;
  updated_at: string;
  tags: { tag_id: number; tag_name: string }[];
  reminder?: Reminder[]; // Optional array of reminders
  shared_with: SharedEmail[];
}
interface SharedEmail {
  email: string;
  type: string;
}

interface Reminder {
  reminder_id?: number;
  reminder_time: string;
  recurring: boolean;
  frequency?: "daily" | "weekly" | "monthly" | "yearly";
}

const getThaiTime = (): string => {
  const date = new Date();
  date.setHours(date.getHours() + 7); // Adjust for Thai timezone
  return date.toISOString().slice(0, 16); // Format: YYYY-MM-DDTHH:mm
};

const ReminderModal = ({
  noteId,
  existingReminder,
  onClose,
  setNotes,
}: {
  noteId: number;
  existingReminder?: Reminder;
  onClose: () => void;
  setNotes: React.Dispatch<React.SetStateAction<Note[]>>;
}) => {
  const [reminderTime, setReminderTime] = useState<string>(
    existingReminder?.reminder_time || getThaiTime()
  );
  const [isRecurring, setIsRecurring] = useState<boolean>(
    existingReminder?.recurring || false
  );
  const [frequency, setFrequency] = useState<"daily" | "weekly" | "monthly" | "yearly">(
    existingReminder?.frequency || "daily"
  );

  const formatReminderTime = (date: string): string => {
    const selectedTime = new Date(date);
    const year = selectedTime.getFullYear();
    const month = String(selectedTime.getMonth() + 1).padStart(2, "0");
    const day = String(selectedTime.getDate()).padStart(2, "0");
    const hours = String(selectedTime.getHours()).padStart(2, "0");
    const minutes = String(selectedTime.getMinutes()).padStart(2, "0");
    return `${year}-${month}-${day} ${hours}:${minutes}:00`;
  };

  const handleSave = async () => {
    const now = new Date();
    const selectedTime = new Date(reminderTime);

    if (selectedTime < now) {
      Swal.fire("Invalid Time", "Reminder time cannot be in the past.", "error");
      return;
    }

    const reminderPayload: Reminder = {
      reminder_time: formatReminderTime(reminderTime),
      recurring: isRecurring,
      frequency: isRecurring ? frequency : undefined,
      reminder_id: existingReminder?.reminder_id ?? undefined,
    };

    try {
      let updatedReminder: Reminder | null = null;

      if (existingReminder?.reminder_id) {
        const response = await axios.put(
          `http://localhost:8000/reminder/${existingReminder.reminder_id}`,
          reminderPayload,
          { withCredentials: true }
        );
        updatedReminder = response.data.reminder;
        Swal.fire("Success", "Reminder updated successfully!", "success");
      } else {
        const response = await axios.post(
          `http://localhost:8000/note/reminder/${noteId}`,
          reminderPayload,
          { withCredentials: true }
        );
        updatedReminder = response.data.reminder;
        Swal.fire("Success", "Reminder created successfully!", "success");
      }

      if (updatedReminder) {
        setNotes((prevNotes) =>
          prevNotes.map((note) =>
            note.note_id === noteId
              ? { ...note, reminder: [updatedReminder] }
              : note
          )
        );
      }

      onClose();
    } catch (err) {
      console.error("Failed to save reminder:", err);
      Swal.fire("Error", "Failed to save the reminder. Please try again.", "error");
    }
  };


  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-white p-6 rounded-lg w-96 z-60"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold mb-4">Set Reminder</h2>

        <label className="block mb-2">Reminder Time:</label>
        <input
          type="datetime-local"
          value={reminderTime}
          onChange={(e) => setReminderTime(e.target.value)}
          className="w-full p-2 border border-gray-300 rounded-md mb-4"
        />

        <div className="flex items-center mb-4">
          <input
            type="checkbox"
            checked={isRecurring}
            onChange={(e) => setIsRecurring(e.target.checked)}
            className="mr-2"
          />
          <label>Recurring</label>
        </div>

        {isRecurring && (
          <div className="mb-4">
            <label className="block mb-2">Frequency:</label>
            <select
              value={frequency}
              onChange={(e) => setFrequency(e.target.value as any)}
              className="w-full p-2 border border-gray-300 rounded-md"
            >
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
              <option value="yearly">Yearly</option>
            </select>
          </div>
        )}

        <div className="flex justify-end space-x-4">
          <button
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="px-4 py-2 bg-indigo-600 text-white rounded-md"
            onClick={handleSave}
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
};

export default function Home() {
  const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(true); // State สำหรับเปิด/ปิด Sidebar
  const [user, setUser] = useState<User | null>(null); // เก็บข้อมูลผู้ใช้ (อาจเป็น null หากยังไม่ได้โหลด)
  const [notes, setNotes] = useState<Note[]>([]);
  const [isPopupOpen, setIsPopupOpen] = useState<boolean>(false); // State สำหรับควบคุม Popup
  const router = useRouter(); // ใช้งาน useRouter

  // Form State
  const [noteTitle, setNoteTitle] = useState<string>("");
  const [noteContent, setNoteContent] = useState<string>("");
  const [isTodo, setIsTodo] = useState<boolean>(false);
  const [todoItems, setTodoItems] = useState<{ id: number; content: string; is_done: boolean }[]>([]);
  const [selectedColor, setSelectedColor] = useState<string>("white");
  const [currentNoteId, setCurrentNoteId] = useState<number | null>(null); // รหัสโน้ตที่กำลังแก้ไข
  const [filterStatus, setFilterStatus] = useState<"all" | "completed" | "incomplete">("all");

  const [isTagPopupOpen, setIsTagPopupOpen] = useState(false);
  const [tagName, setTagName] = useState("");
  const [tags, setTags] = useState<{ tag_id: number; tag_name: string }[]>([]); // เก็บ Tags
  // const [selectedTags, setSelectedTags] = useState<number[]>([]); // เก็บแท็กที่เลือก

  const [showDeletedNotes, setShowDeletedNotes] = useState<boolean>(false); // State สำหรับสลับโหมด Trash
  const [deletedNotes, setDeletedNotes] = useState<Note[]>([]); // เก็บโน้ตที่ถูกลบ

  const [searchTerm, setSearchTerm] = useState<string>(""); // เก็บคำค้นหา
  const [searchTimeout, setSearchTimeout] = useState<NodeJS.Timeout | null>(null); // Timeout handler
  const [filteredNotes, setFilteredNotes] = useState<Note[]>([]); // เก็บโน้ตที่กรองตามการค้นหา

  const [isReminderModalOpen, setIsReminderModalOpen] = useState(false);
  const [currentReminder, setCurrentReminder] = useState<Reminder | null>(null);
  const formatReminderTime = (date: string): string => {
    const selectedTime = new Date(date);

    const year = selectedTime.getFullYear();
    const month = String(selectedTime.getMonth() + 1).padStart(2, "0");
    const day = String(selectedTime.getDate()).padStart(2, "0");
    const hours = String(selectedTime.getHours()).padStart(2, "0");
    const minutes = String(selectedTime.getMinutes()).padStart(2, "0");
    const seconds = "00"; // Fixed seconds to match API expectation

    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
  };

  const [showReminderSection, setShowReminderSection] = useState(false);
  const [showShareSection, setShowShareSection] = useState(false);
  const [showTagsSection, setShowTagsSection] = useState(false);
  const [showSignInSection, setShowSignInSection] = useState(false);

  const searchParams = useSearchParams();
  const noteIdparams = searchParams.get("noteId");
  useEffect(() => {
    if (noteIdparams) {
      setIsPopupOpen(true); // เปิด popup ถ้ามี noteId
    }
  }, [noteIdparams]);

  const fetchData = async (): Promise<string | null> => {
    try {
      const response = await axios.get<{ userId: string }>('http://localhost:3000/api/getCookie', { withCredentials: true });
      return response.data.userId; // Return userId หาก API สำเร็จ
    } catch (err) {
      console.error("Failed to fetch user:", err);
      return null; // Return null หากเกิดข้อผิดพลาด
    }
  };

  const getColorClass = (color: string): string => {
    switch (color) {
      case "blue":
        return "bg-[#B3E5FC]";
      case "green":
        return "bg-[#B9FBC0]";
      case "yellow":
        return "bg-[#FFF9C4]";
      case "red":
        return "bg-[#FFABAB]";
      case "gray":
        return "bg-[#CFD8DC]";
      default:
        return "bg-white"; // Default สีพื้นหลัง
    }
  };
  useEffect(() => {
    console.log("Updated notes:", notes);
  }, [notes]);

  const logout = async () => {
    try {
      const response = await axios.get("http://localhost:8000/logout", {
        withCredentials: true,
      });

      await axios.get('http://localhost:3000/api/removeCookie', { withCredentials: true });
      setUser(null); // รีเซ็ต user เป็น null
      router.push('/'); // เปลี่ยนเส้นทางกลับไปยังหน้า `/`
    } catch (err) {
      console.error("Failed to log out:", err);
    }
  };

  const fetchUserAndNotes = async (): Promise<void> => {
    const userId = await fetchData();
    if (!userId) return;

    try {
      // ดึงข้อมูล User
      const userResponse = await axios.get<{ id: string; username: string; email: string }>(
        `http://localhost:8000/user/${userId}`,
        { withCredentials: true }
      );
      const userData = userResponse.data;
      setUser({ id: userData.id, username: userData.username, email: userData.email });
      // ดึง Notes
      const notesResponse = await axios.get<{ notes: Note[] | null }>(
        `http://localhost:8000/note/${userId}`,
        { withCredentials: true }
      );

      const notes = notesResponse.data.notes || [];

      // ดึงข้อมูล shared_emails สำหรับแต่ละโน้ต
      const notesWithSharedEmails = await Promise.all(
        notes.map(async (note) => {
          try {
            const sharedEmailsResponse = await axios.get<{ shared_emails: { email: string; type: string }[] }>(
              `http://localhost:8000/note/${note.note_id}/shared-emails`,
              { withCredentials: true }
            );

            return {
              ...note,
              shared_with: sharedEmailsResponse.data.shared_emails, // ดึงข้อมูลตรงๆ
            };
          } catch (err) {
            console.error(`Failed to fetch shared_emails for note ${note.note_id}:`, err);
            return { ...note, shared_with: [] }; // ใช้ค่าเริ่มต้นหากดึงล้มเหลว
          }
        })
      );

      // จัดเรียง Notes ตาม priority และเวลาที่อัปเดต
      const sortedNotes = notesWithSharedEmails.sort((a, b) => {
        if (a.priority !== b.priority) {
          return b.priority - a.priority; // Priority สำคัญ (1) อยู่บนสุด
        }
        const dateA = new Date(b.updated_at || b.created_at).getTime();
        const dateB = new Date(a.updated_at || a.created_at).getTime();
        return dateA - dateB;
      });

      setNotes(sortedNotes);
    } catch (err) {
      console.error("Failed to fetch notes or user:", err);
    }
  };

  const openNewNotePopup = () => {
    resetForm();
    setCurrentNoteId(null); // ตั้งค่าให้เป็นโน้ตใหม่
    setIsPopupOpen(true);
  };

  // ดึงตัวอักษรแรกจากชื่อผู้ใช้
  const getFirstLetter = (name: string): string => (name ? name.charAt(0).toUpperCase() : "");
  const resetForm = () => {
    setNoteTitle("");
    setNoteContent("");
    setIsTodo(false);
    setTodoItems([]);
    setSelectedColor("white");
  };

  let nextId = Date.now(); // ใช้เวลาปัจจุบันเป็น id เริ่มต้น

  const handleAddTodoItem = () => {
    setTodoItems((prevItems) => [
      ...prevItems,
      { id: nextId++, content: "", is_done: false }, // เพิ่ม todo item ใหม่พร้อม id ที่ไม่ซ้ำกัน
    ]);
  };

  const handleTodoChange = (id: number, value: string) => {
    setTodoItems((prevItems) =>
      prevItems.map((item) =>
        item.id === id ? { ...item, content: value } : item // อัปเดตเฉพาะรายการที่ตรงกับ id
      )
    );
  };

  const handleRemoveTodoItem = (id: number) => {
    setTodoItems((prevItems) => prevItems.filter((item) => item.id !== id));
  };

  const handleDeleteNote = async (noteId: number) => {
    // แสดงการแจ้งเตือนยืนยัน
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: "You won't be able to revert this!",
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#d33',
      cancelButtonColor: '#9ca3af',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel',
    });

    if (result.isConfirmed) {
      try {
        await axios.delete(`http://localhost:8000/note/${noteId}`, { withCredentials: true });

        // ลบโน้ตออกจาก state
        setNotes((prevNotes) => prevNotes.filter((note) => note.note_id !== noteId));

        // แสดงข้อความสำเร็จ
        Swal.fire('Deleted!', 'Your note has been deleted.', 'success');
      } catch (err) {
        console.error(`Failed to delete note with ID ${noteId}:`, err);

        // แสดงข้อความแสดงข้อผิดพลาด
        Swal.fire('Error', 'Failed to delete the note. Please try again.', 'error');
      }
    }
  };

  const [originalNote, setOriginalNote] = useState<Note | null>(null);

  const handleEditNote = (note: Note) => {
    setCurrentNoteId(note.note_id);
    setNoteTitle(note.title || "");
    setNoteContent(note.content || "");
    setIsTodo(note.is_todo);
    setTodoItems(note.todo_items || []);
    setSelectedColor(note.color || "white");
    setOriginalNote(note); // บันทึกค่าเดิมของโน้ต
    setIsPopupOpen(true);
  };

  const updateNoteColor = async (noteId: number, color: string) => {
    try {
      await axios.put(`http://localhost:8000/note/color/${noteId}`, { color }, {
        withCredentials: true,
      });
      // อัปเดตสีในโน้ตที่มีอยู่ใน state
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId ? { ...note, color } : note
        )
      );
    } catch (err) {
      console.error(`Failed to update note color for ID ${noteId}:`, err);
    }
  };

  const autoSaveNote = async () => {
    if (!noteTitle.trim()) {
      Swal.fire("Warning", "Please provide a title for the note.", "warning");
      return;
    }

    if (!isTodo && !noteContent.trim()) {
      Swal.fire("Warning", "Please provide content for the note.", "warning");
      return;
    }

    if (isTodo && todoItems.length === 0) {
      Swal.fire("Warning", "Please add at least one Todo item.", "warning");
      return;
    }

    if (isTodo && todoItems.length > 0 && todoItems.some((item) => !item.content.trim())) {
      Swal.fire("Warning", "Please provide content for all Todo items.", "warning");
      return;
    }

    try {
      const filteredTodoItems = todoItems
        .filter((item) => item.content.trim() !== "")
        .map((item) => ({
          content: item.content,
          is_done: item.is_done,
        }));

      const newNote = {
        title: noteTitle || "Untitled",
        content: isTodo ? null : noteContent || "",
        color: selectedColor,
        priority: 0,
        is_todo: isTodo,
        is_all_done: false,
        todo_items: isTodo ? filteredTodoItems : null,
      };

      console.log("Payload sent to API (newNote):", JSON.stringify(newNote, null, 2));

      let createdNote: Note;
      if (!currentNoteId) {
        const response = await axios.post<{ message: string; note: Note }>(
          "http://localhost:8000/note",
          newNote,
          { withCredentials: true }
        );
        createdNote = response.data.note;
      } else {
        const response = await axios.put<{ message: string; note: Note }>(
          `http://localhost:8000/note/title-content/${currentNoteId}`,
          { ...newNote, todo_items: filteredTodoItems },
          { withCredentials: true }
        );
        createdNote = response.data.note;
      }

      setNotes((prevNotes) => {
        const updatedNotes = [...prevNotes, createdNote];

        // จัดเรียงโน้ตใหม่ตาม priority และเวลาที่อัปเดต
        return updatedNotes.sort((a, b) => {
          if (a.priority !== b.priority) {
            return b.priority - a.priority; // Priority สำคัญ (1) อยู่บนสุด
          }
          const dateA = new Date(b.updated_at || b.created_at).getTime();
          const dateB = new Date(a.updated_at || a.created_at).getTime();
          return dateA - dateB;
        });
      });

      console.log("Note saved successfully:", createdNote);
    } catch (err) {
      console.error("Failed to save note:", err);
      Swal.fire("Error", "Failed to save the note. Please try again.", "error");
    }
  };

  const updateNoteContent = async () => {
    if (!currentNoteId) return;

    // กรองรายการ todoItems ที่มีเนื้อหา
    const filteredTodoItems = todoItems
      .filter((item) => item.content.trim() !== "")
      .map((item) => ({
        content: item.content,
        is_done: item.is_done,
      }));

    // เตรียมข้อมูลที่ต้องอัปเดต
    const updatedData = isTodo
      ? { todo_items: filteredTodoItems }
      : { content: noteContent };

    console.log("Filtered Todo Items (copyable):", JSON.stringify(filteredTodoItems, null, 2));
    console.log("Updated Data (copyable):", JSON.stringify(updatedData, null, 2));

    // ตรวจสอบว่าไม่มีการเปลี่ยนแปลงข้อมูล
    if (
      originalNote &&
      noteTitle === originalNote.title &&
      (!isTodo
        ? noteContent === originalNote.content
        : JSON.stringify(filteredTodoItems) === JSON.stringify(originalNote.todo_items))
    ) {
      console.log("No changes detected. Skipping update.");
      setIsPopupOpen(false);
      return;
    }

    try {
      // ข้อมูลที่จะส่งไปใน API
      const requestData = { title: noteTitle, ...updatedData };
      console.log("Request Data (copyable):", JSON.stringify(requestData, null, 2));

      // ส่งคำขออัปเดต
      const response = await axios.put(
        `http://localhost:8000/note/title-content/${currentNoteId}`,
        requestData,
        { withCredentials: true }
      );

      const updatedNote = response.data.notes;

      console.log("Response Data (copyable):", JSON.stringify(updatedNote, null, 2));

      // อัปเดตสถานะโน้ตใน UI
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === currentNoteId ? { ...note, ...updatedNote } : note
        )
      );

      setOriginalNote(updatedNote); // อัปเดต originalNote ด้วยค่าที่เพิ่งอัปเดต
      console.log("Note updated successfully:", updatedNote);

      await fetchUserAndNotes();
    } catch (err) {
      console.error("Failed to update note:", err);
      Swal.fire("Error", "Failed to update the note. Please try again.", "error");
    }

    setIsPopupOpen(false);
  };

  const updateNoteStatus = async (noteId: number, statusUpdate: { is_todo?: boolean; is_all_done?: boolean }) => {
    try {
      // ส่งเฉพาะค่าที่ต้องการอัปเดต
      await axios.put(
        `http://localhost:8000/note/status/${noteId}`,
        statusUpdate,
        { withCredentials: true }
      );

      // อัปเดต state ใน frontend
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId
            ? { ...note, ...statusUpdate }
            : note
        )
      );

      console.log(`Note ${noteId} status updated successfully with`, statusUpdate);
    } catch (err) {
      console.error("Failed to update note status:", err);
      Swal.fire("Error", "Failed to update the note status. Please try again.", "error");
    }
  };

  const togglePriority = async (noteId: number) => {
    const noteToUpdate = notes.find((note) => note.note_id === noteId);

    if (!noteToUpdate) return;

    const newPriority = noteToUpdate.priority === 1 ? 0 : 1; // Toggle priority

    try {
      await axios.put(
        `http://localhost:8000/note/priority/${noteId}`,
        { priority: newPriority },
        { withCredentials: true }
      );

      // อัปเดต priority ใน state
      setNotes((prevNotes) => {
        const updatedNotes = prevNotes.map((note) =>
          note.note_id === noteId ? { ...note, priority: newPriority } : note
        );

        // เรียงลำดับโน้ตใหม่โดยพิจารณา priority และ updated_at
        return updatedNotes.sort((a, b) => {
          if (a.priority !== b.priority) {
            return b.priority - a.priority; // Priority สำคัญ (1) อยู่บนสุด
          }
          const dateA = new Date(b.updated_at || b.created_at).getTime();
          const dateB = new Date(a.updated_at || a.created_at).getTime();
          return dateA - dateB;
        });
      });

      console.log("Priority updated successfully.");
    } catch (err) {
      console.error("Failed to update priority:", err);
      Swal.fire("Error", "Failed to update priority. Please try again.", "error");
    }
  };

  const toggleTodoStatus = async (noteId: number, todoId: number, isDone: boolean) => {
    try {

      if (!todoId) {
        console.error(`Todo item not found for todoId: ${todoId}`);
        return;
      }

      await axios.put(
        `http://localhost:8000/note/${noteId}/todo/${todoId}/status`,
        { is_done: !isDone }, // สลับค่า is_done
        { withCredentials: true }
      );

      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId
            ? {
              ...note,
              todo_items: note.todo_items
                ? note.todo_items.map((todo) =>
                  todo.id === todoId
                    ? { ...todo, is_done: !isDone }
                    : todo
                )
                : [],
            }
            : note
        )
      );

      console.log(`Todo item ${todoId} status updated successfully.`);
    } catch (err) {
      console.error("Failed to update todo status:", err);
      Swal.fire("Error", "Failed to update the todo status. Please try again.", "error");
    }
  };

  const fetchTags = async () => {
    try {
      const response = await axios.get("http://localhost:8000/tag", {
        withCredentials: true,
      });
      setTags(response.data || []); // ตั้งค่า Tags ใน State
    } catch (error) {
      console.error("Failed to fetch tags:", error);
      Swal.fire("Error", "Failed to load tags. Please try again.", "error");
    }
  };

  const handleCreateTag = async () => {
    if (!tagName.trim()) {
      Swal.fire("Warning", "Tag name cannot be empty.", "warning");
      return;
    }

    try {
      // ส่งคำขอไปที่ API เพื่อสร้างแท็กใหม่
      const response = await axios.post(
        "http://localhost:8000/tag",
        { tag_name: tagName },
        { withCredentials: true }
      );

      Swal.fire("Success", "Tag created successfully!", "success");

      // ดึงข้อมูลแท็กจาก response โดยตรง
      const newTag = response.data.tag;

      // เพิ่มแท็กใหม่ลงใน `tags` state ทันที
      setTags((prevTags) => [...prevTags, newTag]);

      setIsTagPopupOpen(false); // ปิด Popup หลังสร้างสำเร็จ
      setTagName(""); // รีเซ็ตชื่อแท็ก
    } catch (error: any) {
      // ตรวจสอบถ้า backend ส่ง status 500 และมีข้อความบอกว่าแท็กถูกสร้างมาแล้ว
      if (error.response && error.response.status === 500) {
        Swal.fire("Error", error.response.data.message || "This tag already exists.", "error");
      } else {
        console.error("Failed to create tag:", error);
        Swal.fire("Error", "Failed to create tag. Please try again.", "error");
      }
    }
  };

  const handleEditTag = async (tagId: number, currentTagName: string) => {
    const { value: newTagName } = await Swal.fire({
      title: "Edit Tag",
      input: "text",
      inputValue: currentTagName,
      showCancelButton: true,
      confirmButtonText: "Save",
      cancelButtonText: "Cancel",
      inputValidator: (value) => {
        if (!value.trim()) {
          return "Tag name cannot be empty!";
        }
      },
    });

    if (newTagName) {
      try {
        // อัปเดตชื่อแท็กใน Backend
        await axios.put(
          `http://localhost:8000/tag/${tagId}`,
          { new_tagname: newTagName },
          { withCredentials: true }
        );

        // อัปเดตแท็กใน `tags` state
        setTags((prevTags) =>
          prevTags.map((tag) =>
            tag.tag_id === tagId ? { ...tag, tag_name: newTagName } : tag
          )
        );

        // อัปเดตแท็กใน `notes` state
        setNotes((prevNotes) =>
          prevNotes.map((note) => ({
            ...note,
            tags: note.tags
              ? note.tags.map((tag) =>
                tag.tag_id === tagId ? { ...tag, tag_name: newTagName } : tag
              )
              : [], // Handle the case where tags is null or undefined
          }))
        );

        Swal.fire("Success", "Tag updated successfully!", "success");
      } catch (err) {
        console.error("Failed to update tag:", err);
        Swal.fire("Error", "Failed to update the tag. Please try again.", "error");
      }
    }
  };

  useEffect(() => {
    fetchTags();
    fetchUserAndNotes();
  }, []);

  const handleDeleteTag = async (tagId: number) => {
    const result = await Swal.fire({
      title: "Are you sure?",
      text: "You won't be able to revert this!",
      icon: "warning",
      showCancelButton: true,
      confirmButtonColor: "#d33",
      cancelButtonColor: "#9ca3af",
      confirmButtonText: "Yes, delete it!",
      cancelButtonText: "Cancel",
    });

    if (result.isConfirmed) {
      try {
        await axios.delete(`http://localhost:8000/tag/${tagId}`, { withCredentials: true });

        // ลบแท็กออกจาก state
        setTags((prevTags) => prevTags.filter((tag) => tag.tag_id !== tagId));

        Swal.fire("Deleted!", "Your tag has been deleted.", "success");
      } catch (error) {
        console.error("Failed to delete tag:", error);
        Swal.fire("Error", "Failed to delete the tag. Please try again.", "error");
      }
    }
  };

  const handleAddTagToNote = async (noteId: number, tagId: number, tagName: string) => {
    try {
      await axios.post(
        "http://localhost:8000/note/add-tag",
        { note_id: noteId, tag_id: tagId },
        { withCredentials: true }
      );

      // อัปเดตสถานะใน state
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId
            ? {
              ...note,
              tags: [...(note.tags || []), { tag_id: tagId, tag_name: tagName }]
            }
            : note
        )
      );

      console.log(`Tag ${tagId} added to Note ${noteId}`);

    } catch (err) {
      console.error(`Failed to add tag to note: ${err}`);
      Swal.fire("Error", "Failed to add tag to the note. Please try again.", "error");
    }
  };

  const handleRemoveTagFromNote = async (noteId: number, tagId: number) => {
    try {
      await axios.post("http://localhost:8000/note/remove-tag", { note_id: noteId, tag_id: tagId }, { withCredentials: true });

      // อัปเดตสถานะใน state
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId
            ? { ...note, tags: note.tags?.filter((tag) => tag.tag_id !== tagId) }
            : note
        )
      );
      console.log(`Tag ${tagId} removed from Note ${noteId}`);
    } catch (err) {
      console.error(`Failed to remove tag from note: ${err}`);
      Swal.fire("Error", "Failed to remove tag from the note. Please try again.", "error");
    }
  };

  const fetchDeletedNotes = async (): Promise<void> => {
    const userId = await fetchData(); // ดึง userId จาก Cookie
    if (!userId) return;

    try {
      const response = await axios.get<{ deleted_notes: Note[] }>(
        `http://localhost:8000/note/deleted/${userId}`,
        { withCredentials: true }
      );
      const deletedNotes = response.data.deleted_notes || []; // เข้าถึงข้อมูลใน deleted_notes
      setDeletedNotes(deletedNotes);
    } catch (err) {
      console.error("Failed to fetch deleted notes:", err);
      Swal.fire("Error", "Failed to load deleted notes. Please try again.", "error");
    }
  };

  const restoreNote = async (noteId: number) => {
    try {
      await axios.put(
        `http://localhost:8000/note/restore/${noteId}`,
        {},
        { withCredentials: true }
      );
      Swal.fire("Success", "Note restored successfully!", "success");

      // ลบโน้ตออกจาก Trash
      setDeletedNotes((prevNotes) => prevNotes.filter((n) => n.note_id !== noteId));

      // เพิ่มโน้ตที่กู้คืนกลับไปยัง notes
      const restoredNote = deletedNotes.find((n) => n.note_id === noteId);
      if (restoredNote) {
        setNotes((prevNotes) => {
          const updatedNotes = [...prevNotes, restoredNote];

          // จัดเรียงโน้ตใหม่ตาม priority และ updated_at
          return updatedNotes.sort((a, b) => {
            if (a.priority !== b.priority) {
              return b.priority - a.priority; // Priority สำคัญ (1) อยู่บนสุด
            }
            const dateA = new Date(b.updated_at || b.created_at).getTime();
            const dateB = new Date(a.updated_at || a.created_at).getTime();
            return dateA - dateB;
          });
        });
      }
    } catch (err) {
      console.error("Failed to restore note:", err);
      Swal.fire("Error", "Failed to restore the note. Please try again.", "error");
    }
  };

  useEffect(() => {
    if (searchTimeout) {
      clearTimeout(searchTimeout); // ยกเลิก Timeout ก่อนหน้า
    }

    // หน่วงเวลา 1 วินาทีหลังจากเริ่มพิมพ์
    const timeout = setTimeout(() => {
      let filtered = notes;

      // กรองโน้ตตามสถานะ filterStatus ก่อน
      if (filterStatus === "completed") {
        filtered = notes.filter((note) => note.is_all_done);
      } else if (filterStatus === "incomplete") {
        filtered = notes.filter((note) => !note.is_all_done);
      }

      // กรองโน้ตที่ตรงกับ searchTerm
      const lowerSearchTerm = searchTerm.toLowerCase();
      filtered = filtered.filter((note) => {
        if (!note || typeof note.title !== "string") {
          return false; // Skip invalid notes
        }
        return (
          note.title.toLowerCase().includes(lowerSearchTerm) ||
          (note.content && note.content.toLowerCase().includes(lowerSearchTerm)) ||
          (note.todo_items &&
            note.todo_items.some((item) =>
              item.content.toLowerCase().includes(lowerSearchTerm)
            ))
        );
      });


      setFilteredNotes(filtered); // อัปเดตโน้ตที่กรองแล้ว
    }, 200); // หน่วงเวลา 1 วินาที

    setSearchTimeout(timeout); // เก็บ Timeout ที่ตั้งไว้

    return () => {
      clearTimeout(timeout); // เคลียร์ Timeout เมื่อ searchTerm หรือ notes เปลี่ยน
    };
  }, [searchTerm, notes, filterStatus]);

  const handleCloseReminderModal = () => {
    setIsReminderModalOpen(false);

    // รีเฟรชค่า reminder ใน currentReminder
    if (currentNoteId) {
      const updatedNote = notes.find((note) => note.note_id === currentNoteId);
      if (updatedNote) {
        setCurrentReminder(updatedNote.reminder?.[0] || null);
      }
    }
  };

  const handleRemoveReminder = async (reminderId: number) => {
    try {
      // Call the API to delete the reminder by reminderId
      await axios.delete(`http://localhost:8000/reminder/${reminderId}`, {
        withCredentials: true,
      });

      // Update the notes state to remove the reminder for the respective note
      setNotes((prevNotes) =>
        prevNotes.map((note) => {
          // If the note has the reminder matching the given reminderId, remove it
          if (note.reminder) {
            return {
              ...note,
              reminder: note.reminder.filter(
                (reminder) => reminder.reminder_id !== reminderId
              ),
            };
          }
          return note; // For notes without reminders, return them as is
        })
      );

      // Clear the current reminder and show success message
      setCurrentReminder(null);
      Swal.fire("Success", "Reminder removed successfully!", "success");
    } catch (err) {
      console.error("Failed to remove reminder:", err);
      Swal.fire("Error", "Failed to remove the reminder. Please try again.", "error");
    }
  };

  const [shareEmail, setShareEmail] = useState<string>("");

  const handleShareNote = async () => {
    if (!shareEmail.trim() || !currentNoteId) {
      Swal.fire("Warning", "Please enter a valid email.", "warning");
      return;
    }

    try {
      const payload = {
        note_id: currentNoteId,
        email: shareEmail,
      };

      // เพิ่ม Shared Emails
      await axios.post("http://localhost:8000/note/share", payload, {
        withCredentials: true,
      });

      // ดึง Notes ใหม่หลังจาก Share
      await fetchUserAndNotes();

      Swal.fire("Success", "Note shared successfully!", "success");

      setShareEmail(""); // รีเซ็ต Email Input
    } catch (err) {
      console.error("Failed to share note:", err);
      Swal.fire("Error", "Failed to share the note. Please try again.", "error");
    }
  };

  const handleRemoveSharedUser = async (noteId: number, email: string) => {
    try {
      await axios.post(
        "http://localhost:8000/note/remove-share",
        { note_id: noteId, email },
        { withCredentials: true }
      );

      // อัปเดต shared_with ใน state
      setNotes((prevNotes) =>
        prevNotes.map((note) =>
          note.note_id === noteId
            ? {
              ...note,
              shared_with: note.shared_with?.filter((user) => user.email !== email),
            }
            : note
        )
      );

      Swal.fire("Success", "User removed from shared note.", "success");
    } catch (err) {
      console.error("Failed to remove shared user:", err);
      Swal.fire("Error", "Failed to remove the user. Please try again.", "error");
    }
  };

  const handleGoogleSignIn = () => {
    // บันทึก state ที่ต้องการใน localStorage
    localStorage.setItem("currentState", JSON.stringify({
      isPopupOpen,
      currentNoteId,
      noteTitle,
      noteContent,
      isTodo,
      todoItems,
      selectedColor,
    }));

    // Redirect ไปที่ backend `/authorize`
    const state = encodeURIComponent(
      JSON.stringify({ redirectTo: "/note" })
    ); // ใช้ state เพื่อเก็บค่า redirect URL
    window.location.href = `http://localhost:8000/authorize?state=${state}`;
  };

  useEffect(() => {
    const savedState = localStorage.getItem("currentState");

    if (savedState) {
      const parsedState = JSON.parse(savedState);

      // นำ state ที่บันทึกไว้มาใช้
      setIsPopupOpen(parsedState.isPopupOpen);
      setCurrentNoteId(parsedState.currentNoteId);
      setNoteTitle(parsedState.noteTitle);
      setNoteContent(parsedState.noteContent);
      setIsTodo(parsedState.isTodo);
      setTodoItems(parsedState.todoItems);
      setSelectedColor(parsedState.selectedColor);

      // ลบข้อมูลใน localStorage หลังจากดึงมาใช้
      localStorage.removeItem("currentState");
    }
  }, []);

  const [sessionId, setSessionId] = useState(null);
  useEffect(() => {
    const fetchSession = async () => {
      try {
        const response = await axios.get("http://localhost:3000/api/session", {
          withCredentials: true,
        });
        setSessionId(response.data.session_id || null);
      } catch (error) {
        // หากเกิดข้อผิดพลาดให้กำหนด sessionId เป็น null และไม่แสดง error ใน UI
        setSessionId(null);
        console.warn("Session not found or failed to fetch. Proceeding without session.");
      }
    };

    fetchSession();
  }, []);

  const [formData, setFormData] = useState({
    summary: "",
    location: "",
    description: "",
    start: "",
    end: "",
  });
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({ ...prevData, [name]: value }));
  };
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Format start and end times to include ":00+07:00"
      const formatTime = (time: string) => `${time}:00+07:00`;

      // Dynamically populate summary and description
      const note = notes.find((note) => note.note_id === currentNoteId);
      const summary = note?.title || noteTitle;
      let description = note?.content || noteContent || "";

      if (!description && (note?.todo_items || todoItems.length > 0)) {
        // Combine todo_items content into description if content is empty
        const items = note?.todo_items || todoItems;
        description = items.map((item) => `- ${item.content}`).join("\n");
      }

      const payload = {
        summary,
        location: "Thailand",
        description,
        start: formatTime(formData.start),
        end: formatTime(formData.end),
      };

      console.log("Payload being sent:", payload);

      // Send the request using axios
      const response = await axios.post("http://localhost:8000/create", payload, {
        headers: {
          "Content-Type": "application/json",
        },
        withCredentials: true, // Include cookies for session validation
      });

      console.log("Response from server:", response.data);

      if (response.status === 200) {
        alert("Event created successfully!");
      } else {
        alert(`Failed to create event: ${response.data.message}`);
      }
    } catch (error: any) {
      console.error("Error creating event:", error);
      alert(error.response?.data?.message || "An error occurred. Please try again.");
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      {/* Navbar */}
      <header className="bg-white px-4 py-3 flex items-center justify-between border-b border-gray-300">
        {/* Left Section: Sidebar Toggle + Title */}
        <div className="flex items-center">
          <button
            className="mr-4 text-gray-600 hover:text-gray-800"
            onClick={() => setIsSidebarOpen(!isSidebarOpen)} // Toggle Sidebar
          >
            <FontAwesomeIcon icon={faBars} />
          </button>
          <h1 className="text-xl font-semibold text-gray-800">Super Note</h1>
        </div>

        {/* Right Section: User Info + Logout */}
        <div className="flex items-center space-x-4">
          {user ? (
            <>
              {/* User Circle */}
              <div className="w-10 h-10 bg-gray-500 text-white flex items-center justify-center rounded-full">
                {getFirstLetter(user.username)}
              </div>
              {/* User Name */}
              <span className="text-gray-800 font-medium">{user.username}</span>
              {/* Logout Button */}
              <button
                className="bg-red-500 text-white hover:bg-red-600 text-sm font-medium p-2 rounded-md"
                onClick={logout} // เรียกฟังก์ชัน logout เมื่อกดปุ่ม
              >
                Logout
              </button>
            </>
          ) : (
            <span className="text-gray-500 text-sm">Loading...</span>
          )}
        </div>
      </header>

      {/* Sidebar + Content */}
      <div className="flex flex-1">
        {/* Sidebar */}
        {isSidebarOpen && (
          <aside className="w-64 bg-white p-4 transition-transform duration-300 ease-in-out">
            <ul className="space-y-4">
              <li
                className={`cursor-pointer ${!showDeletedNotes ? "font-semibold text-indigo-600" : "text-gray-600"}`}
                onClick={() => {
                  setShowDeletedNotes(false); // ปิดโหมด Trash
                  fetchUserAndNotes(); // ดึงโน้ตปกติ
                }}
              >
                Note
              </li>
              <li className="text-gray-600">Reminder</li>
              <li>
                <div className="flex justify-between items-center">
                  <span className="text-gray-600 ">Tags</span>
                  <button
                    onClick={() => setIsTagPopupOpen(true)} // เปิด Popup สร้าง Tag
                    className="text-indigo-600 hover:underline text-sm"
                  >
                    Add
                  </button>
                </div>
                <ul className="mt-2 space-y-2">
                  {tags.map((tag) => (
                    <li
                      key={tag.tag_id}
                      className="flex justify-between items-center text-gray-700 bg-gray-100 px-4 py-1 rounded-md cursor-pointer"
                      onClick={() => handleEditTag(tag.tag_id, tag.tag_name)} // แก้ไขแท็ก
                    >
                      <span>{tag.tag_name}</span>
                      <button
                        onClick={(e) => {
                          e.stopPropagation(); // ป้องกันการเรียก handleEditTag
                          handleDeleteTag(tag.tag_id); // ลบแท็ก
                        }}
                        className="text-gray-500 hover:text-gray-700"
                      >
                        X
                      </button>
                    </li>
                  ))}
                </ul>
              </li>

              <li
                className={` cursor-pointer ${showDeletedNotes ? "font-semibold text-indigo-600" : "text-gray-600"}`}
                onClick={() => {
                  setShowDeletedNotes(true); // เปิดโหมด Trash
                  fetchDeletedNotes(); // ดึงโน้ตที่ถูกลบ
                }}
              >
                Trash
              </li>
            </ul>
          </aside>
        )}
        {isTagPopupOpen && (
          <div
            className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"
            onClick={() => setIsTagPopupOpen(false)} // ปิด Popup เมื่อคลิกข้างนอก
          >
            <div
              className="bg-white rounded-lg p-6 w-1/3"
              onClick={(e) => e.stopPropagation()} // ป้องกันการปิด Popup เมื่อคลิกด้านใน
            >
              <h2 className="text-lg font-bold mb-4">Create New Tag</h2>
              <input
                type="text"
                placeholder="Tag Name"
                value={tagName}
                onChange={(e) => setTagName(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md mb-4"
              />
              <div className="flex justify-end space-x-4">
                <button
                  onClick={() => setIsTagPopupOpen(false)} // ปิด Popup
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateTag} // เรียกฟังก์ชันสร้าง Tag
                  className="px-4 py-2 bg-indigo-600 text-white rounded-md"
                >
                  Create
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Main Content */}
        <main className="flex-1 p-6 transition-all duration-300">
          {/* Search Bar */}
          <div className="mb-4">
            <input
              type="text"
              placeholder="ค้นหาโน้ต..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)} // อัปเดตคำค้นหา
              className="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Tabs */}
          <div className="flex space-x-8 mb-4">
            <button
              className={`pb-1 border-b-2 ${filterStatus === "all" ? "text-indigo-600 border-indigo-600" : "text-gray-600 border-transparent"}`}
              onClick={() => setFilterStatus("all")}
            >
              ทั้งหมด
            </button>
            <button
              className={`pb-1 border-b-2 ${filterStatus === "incomplete" ? "text-indigo-600 border-indigo-600" : "text-gray-600 border-transparent"}`}
              onClick={() => setFilterStatus("incomplete")}
            >
              ยังไม่เสร็จ
            </button>
            <button
              className={`pb-1 border-b-2 ${filterStatus === "completed" ? "text-indigo-600 border-indigo-600" : "text-gray-600 border-transparent"}`}
              onClick={() => setFilterStatus("completed")}
            >
              เสร็จแล้ว
            </button>
          </div>

          {/* Notes */}
          <div className="grid grid-cols-3 gap-4">
            {showDeletedNotes
              ? deletedNotes.map((note, index) => (
                <div
                  key={note.note_id || `note-${index}`}
                  className={`p-4 rounded-md shadow ${getColorClass(note.color)} h-64 relative`}
                >
                  <button
                    className={`absolute top-2 right-2 ${note.priority === 1 ? "text-yellow-500" : "text-gray-400"}`}
                    onClick={(e) => e.stopPropagation()} // ปิดการ Restore
                  >
                    <FontAwesomeIcon icon={faStar} />
                  </button>
                  {/* Title */}
                  <h2 className="font-bold text-lg">{note.title || "Untitled Note"}</h2>

                  {/* Tags */}
                  {note.tags?.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-2">
                      {note.tags.map((tag) => (
                        <span
                          key={tag.tag_id}
                          className="bg-gray-100 text-gray-700 px-2 py-1 rounded-full text-xs bg-opacity-80"
                        >
                          {tag.tag_name}
                        </span>
                      ))}
                    </div>
                  )}

                  {/* Content or Todo Items */}
                  {note.is_todo && note.todo_items ? (
                    <ul className="mt-2 space-y-1">
                      {note.todo_items.map((item, itemIndex) => (
                        <li
                          key={item.id || `item-${itemIndex}`}
                          className="flex items-center space-x-2"
                        >
                          {/* Checkbox */}
                          <input
                            type="checkbox"
                            checked={item.is_done}
                            readOnly
                            className="cursor-default"
                          />
                          {/* Todo Item Content */}
                          <span
                            className={
                              item.is_done ? "line-through text-gray-500" : "text-gray-700"
                            }
                          >
                            {item.content || "No content"}
                          </span>
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="mt-2 text-gray-600">{note.content}</p>
                  )}

                  {/* Restore Button */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      restoreNote(note.note_id);
                    }}
                    className="absolute bottom-2 right-2 bg-indigo-600 text-white px-2 py-1 rounded-md"
                  >
                    Restore
                  </button>
                </div>
              ))
              :
              filteredNotes.map((note, index) => (
                <div
                  key={note.note_id || `note-${index}`}
                  className={`px-4 py-3 rounded-md shadow-md h-64 border border-gray-200 relative ${getColorClass(note.color)} hover:shadow-lg transition-shadow duration-200`}
                  onClick={() => handleEditNote(note)}
                >

                  {/* Checkbox */}
                  <input
                    type="checkbox"
                    className="absolute -top-2 -left-2 w-5 h-5 cursor-pointer"
                    checked={note.is_all_done}
                    onClick={(e) => e.stopPropagation()}
                    onChange={(e) => {
                      e.stopPropagation();
                      updateNoteStatus(note.note_id, { is_all_done: e.target.checked });
                    }}
                  />
                  {/* Tags */}
                  {note.tags?.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                      {note.tags.map((tag) => (
                        <span
                          key={tag.tag_id}
                          className="bg-gray-200 text-gray-700 px-1.5 py-0.5 rounded-full text-xs bg-opacity-80"
                        >
                          {tag.tag_name}
                        </span>
                      ))}
                    </div>
                  )}
                  {/* Priority Icon */}
                  <button
                    className={`absolute top-2 right-2 text-gray-400 hover:text-yellow-500`}
                    onClick={(e) => {
                      e.stopPropagation();
                      togglePriority(note.note_id);
                    }}
                  >
                    <FontAwesomeIcon
                      icon={faStar}
                      className={`${note.priority === 1 ? "text-yellow-500" : ""}`}
                    />
                  </button>

                  {/* Shared Icon */}
                  {note.shared_with && note.shared_with.length > 1 && (
                    <div className="absolute bottom-2 right-2 flex items-center space-x-1">
                      <FontAwesomeIcon
                        icon={faUserFriends}
                        className="text-gray-400"
                        title={`${note.shared_with.length - 1} other users`}
                      />
                      <span className="text-xs text-gray-500">
                        +{note.shared_with.length - 1}
                      </span>
                    </div>
                  )}
                  {/* Trash Icon */}
                  <button
                    className="absolute bottom-2 left-2 text-gray-500 hover:text-gray-700"
                    onClick={(e) => {
                      handleDeleteNote(note.note_id);
                      e.stopPropagation();
                    }}
                  >
                    <FontAwesomeIcon icon={faTrash} />
                  </button>

                  {/* Title */}
                  <h2 className="font-medium text-md text-gray-800 truncate">
                    {note.title || "Untitled Note"}
                  </h2>

                  {/* Content */}
                  <p className="mt-1 text-sm text-gray-600 line-clamp-3">
                    {note.content}
                  </p>

                  {/* Todo Items */}
                  {note.is_todo && note.todo_items && (
                    <ul className="mt-2 space-y-1">
                      {note.todo_items.slice(0, 5).map((item, idx) => (
                        <li
                          key={item.id || `item-${idx}`}
                          className={`flex items-center text-sm space-x-2 ${item.is_done ? "line-through text-gray-400" : "text-gray-700"
                            }`}
                        >
                          <input
                            type="checkbox"
                            checked={item.is_done}
                            readOnly
                            className="cursor-default"
                          />
                          <span className="truncate">{item.content || "No content"}</span>
                        </li>
                      ))}
                      {note.todo_items.length > 5 && (
                        <li className="text-xs text-gray-500">+ More tasks</li>
                      )}
                    </ul>
                  )}

                  {/* Reminder */}
                  {note.reminder && note.reminder.length > 0 && (
                    <div className="absolute bottom-2 left-8 flex items-center text-xs text-gray-500">
                      <FontAwesomeIcon icon={faClock} className="mr-1" />
                      {new Date(note.reminder[0]?.reminder_time).toLocaleString("en-GB", {
                        day: "2-digit",
                        month: "2-digit",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </div>
                  )}
                </div>

              ))}
          </div>
        </main>
      </div>

      {/* Floating Button */}
      <button className="fixed bottom-4 right-4 bg-indigo-600 text-white p-4 rounded-full shadow-lg" onClick={openNewNotePopup}
      >
        <FontAwesomeIcon icon={faPlus} className="w-6 h-5" />
      </button>

      {/* Popup Modal */}
      {isPopupOpen && (
        <div
          className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50 overflow-scroll"
          onClick={() => {
            if (currentNoteId) {
              updateNoteContent(); // Call function to update note
            } else {
              autoSaveNote(); // Save new note
            }
            setIsPopupOpen(false); // Close the popup
          }}
        >
          <div
            className={`${getColorClass(selectedColor)} rounded-md shadow-lg w-96 p-6`}
            onClick={(e) => e.stopPropagation()} // Prevent closing when clicking inside
          >
            {/* Title Input */}
            <input
              type="text"
              placeholder="Title"
              value={noteTitle}
              onChange={(e) => setNoteTitle(e.target.value)}
              className="w-full text-gray-700 border-b border-gray-300 focus:outline-none focus:border-indigo-500 p-2 mb-4"
            />


            {/* Content or Todo List */}
            {!isTodo ? (
              <textarea
                placeholder="Content"
                value={noteContent}
                onChange={(e) => setNoteContent(e.target.value)}
                className="w-full text-gray-700 border-b border-gray-300 focus:outline-none focus:border-indigo-500 p-2"
              />
            ) : (
              <>
                {/* Todo Items */}
                <ul className="space-y-2 mb-4">
                  {todoItems
                    .slice()
                    .sort((a, b) => {
                      // Sort by is_done (unchecked first) and then by id
                      if (a.is_done === b.is_done) {
                        return a.id - b.id;
                      }
                      return a.is_done ? 1 : -1;
                    })
                    .map((item, index) => (
                      <li
                        key={`${item.id}`}
                        className="flex items-center gap-2"
                      >
                        {/* Checkbox */}
                        <input
                          type="checkbox"
                          checked={item.is_done}
                          onChange={async () => {
                            // Call toggleTodoStatus to update backend
                            await toggleTodoStatus(currentNoteId!, item.id, item.is_done);

                            // Update local state for immediate UI feedback
                            setTodoItems((prevItems) =>
                              prevItems.map((todo) =>
                                todo.id === item.id
                                  ? { ...todo, is_done: !item.is_done } // Toggle is_done
                                  : todo
                              )
                            );
                          }}
                          className="cursor-pointer"
                        />
                        {/* Editable Todo Item */}
                        <input
                          type="text"
                          placeholder="Todo item"
                          value={item.content}
                          onChange={(e) => handleTodoChange(item.id, e.target.value)}
                          className={`flex-grow text-gray-700 p-0.5 focus:outline-none focus:border-indigo-500 ${item.is_done ? "line-through text-gray-500" : ""
                            }`}
                        />
                        {/* Remove Button */}
                        <button
                          onClick={() => handleRemoveTodoItem(item.id)}
                          className="text-gray-400 hover:text-gray-600"
                        >
                          &times;
                        </button>
                      </li>
                    ))}
                </ul>
                {/* Add Todo Item */}
                <button
                  className="w-full bg-gray-100 text-gray-600 rounded-md py-2 text-sm hover:bg-gray-200"
                  onClick={handleAddTodoItem}
                >
                  Add Todo Item
                </button>
              </>
            )}


            {/* Color Picker */}
            <div className="flex items-center gap-2 mt-4">
              {["blue", "green", "yellow", "red", "gray", "white"].map((color) => (
                <button
                  key={color}
                  className={`w-8 h-8 rounded-full ${getColorClass(color)} ${selectedColor === color ? "ring-2 ring-indigo-500" : ""
                    }`}
                  onClick={() => {
                    setSelectedColor(color);
                    if (currentNoteId) {
                      updateNoteColor(currentNoteId, color);
                    }
                  }}
                />
              ))}
            </div>

            {/* Is Todo Toggle */}
            <div className="flex items-center justify-between mt-4 mb-4">
              <span className="text-sm font-medium text-gray-700">Enable Todo</span>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={isTodo}
                  onChange={(e) => {
                    const newIsTodo = e.target.checked;
                    setIsTodo(newIsTodo); // อัปเดต state
                    if (currentNoteId) {
                      updateNoteStatus(currentNoteId, { is_todo: newIsTodo }); // อัปเดต API
                    }
                  }}
                  className="sr-only peer" // ใช้ sr-only เพื่อซ่อน checkbox แต่ยังสามารถโฟกัสได้
                />
                <div className="w-9 h-5 bg-gray-200 rounded-full peer peer-checked:bg-indigo-600 transition-all"></div>
                <div className="w-4 h-4 bg-white border border-gray-300 rounded-full absolute left-0.5 top-0.5 peer-checked:translate-x-4 peer-checked:border-white transition-transform"></div>
              </label>
            </div>

            <div className="flex items-center gap-6 text-gray-500">
              {[
                { icon: faTag, action: () => setShowTagsSection(!showTagsSection), tooltip: "Tags" },
                { icon: faBell, action: () => setShowReminderSection(!showReminderSection), tooltip: "Reminders" },
                { icon: faUserFriends, action: () => setShowShareSection(!showShareSection), tooltip: "Share" },
                { icon: faCalendar, action: () => setShowSignInSection(!showSignInSection), tooltip: "Calendar" },
              ].map((item, index) => (
                <div
                  key={index}
                  className="relative group cursor-pointer"
                  onClick={item.action}
                >
                  <div className="flex items-center justify-center w-10 h-10 bg-gray-200 rounded-full shadow-md transition-transform transform hover:scale-110 hover:bg-indigo-500 hover:text-white">
                    <FontAwesomeIcon icon={item.icon} className="w-5 h-5" />
                  </div>
                  {/* Tooltip */}
                  <div className="absolute bottom-12 left-1/2 transform -translate-x-1/2 bg-gray-800 text-white text-xs rounded-lg px-2 py-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {item.tooltip}
                  </div>
                </div>
              ))}
            </div>
            {/* Tag Selection */}
            {showTagsSection && (
              <div className="mt-4">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700">Tags</span>
                  <button
                    onClick={() => setIsTagPopupOpen(true)} // เปิด Popup เลือก Tag
                    className="text-indigo-500 hover:underline text-xs"
                  >
                    + Add
                  </button>
                </div>

                {/* แสดงแท็กที่เลือก */}
                <ul className="flex flex-wrap gap-2">
                  {currentNoteId && notes.length > 0 &&
                    notes
                      .find((note) => note.note_id === currentNoteId)
                      ?.tags?.map((tag) => (
                        <li
                          key={tag.tag_id}
                          className="flex items-center text-xs bg-gray-200 text-gray-600 px-2 py-1 rounded-full"
                        >
                          <span>{tag.tag_name}</span>
                          <button
                            onClick={() => handleRemoveTagFromNote(currentNoteId, tag.tag_id)}
                            className="ml-1 text-gray-400 hover:text-gray-600"
                          >
                            &times;
                          </button>
                        </li>
                      ))}
                </ul>

              </div>
            )}
            {/* Reminder Section */}
            {showReminderSection && (

              <div className="mt-4">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700">Reminder</span>
                  {currentNoteId && (
                    () => {
                      const currentNote = notes.find((n) => n.note_id === currentNoteId);
                      const currentReminder = currentNote?.reminder?.[0];

                      return currentReminder ? (
                        <div className="flex items-center gap-2">
                          <span className="text-sm text-gray-600">
                            {currentReminder?.reminder_time
                              ? new Date(currentReminder.reminder_time).toLocaleString("en-GB", {
                                day: "2-digit",
                                month: "2-digit",
                                year: "numeric",
                                hour: "2-digit",
                                minute: "2-digit",
                                second: "2-digit",
                                hour12: false,
                              })
                              : "No reminder set"}
                          </span>
                          <div className="flex gap-2">

                            <button
                              className="text-indigo-500 hover:text-indigo-700 text-xs"
                              onClick={() => {
                                setCurrentReminder(currentReminder);
                                setIsReminderModalOpen(true);
                              }}
                            >
                              Edit
                            </button>
                            <button
                              className="text-red-500 hover:text-red-700 text-xs"
                              onClick={() => {
                                if (currentReminder?.reminder_id !== undefined) {
                                  handleRemoveReminder(currentReminder.reminder_id);
                                } else {
                                  console.error("Reminder ID is undefined");
                                }
                              }}
                            >
                              Remove
                            </button>
                          </div>

                        </div>
                      ) : (
                        <button
                          className="text-indigo-600 hover:underline text-sm"
                          onClick={() => {
                            setCurrentReminder(null);
                            setIsReminderModalOpen(true);
                          }}
                        >
                          Add Reminder
                        </button>
                      );
                    })()}
                </div>
              </div>
            )}

            {/* Share Note Section */}
            {showShareSection && (
              <div className="mt-4">
                <div className="flex justify-between items-center mb-2">
                  <span className="text-sm font-medium text-gray-700">Share Note</span>
                </div>
                <div className="flex gap-2 mb-2">
                  <input
                    type="email"
                    placeholder="Enter user email"
                    value={shareEmail}
                    onChange={(e) => setShareEmail(e.target.value)}
                    className="flex-grow p-2 border border-gray-300 rounded-md text-sm"
                  />
                  <button
                    className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm hover:bg-indigo-700"
                    onClick={handleShareNote}
                  >
                    Share
                  </button>
                </div>

                {/* Shared Emails */}
                <div className="mt-2">
                  <span className="text-sm font-medium text-gray-700">Shared With:</span>
                  <ul className="flex flex-wrap gap-2 mt-1">
                    {notes
                      .find((note) => note.note_id === currentNoteId)
                      ?.shared_with?.map((user) => (
                        <li
                          key={user.email}
                          className={`flex items-center px-2 py-1 rounded-full text-xs ${user.type === "owner" ? "bg-yellow-200 text-yellow-800" : "bg-gray-100 text-gray-700"
                            }`}
                        >
                          <span>
                            {user.email} {user.type === "owner" ? "(Owner)" : ""}
                          </span>
                          {user.type === "shared" && (
                            <button
                              className="ml-2 text-red-500 hover:text-red-700"
                              onClick={() => handleRemoveSharedUser(currentNoteId!, user.email)} // เรียกฟังก์ชันลบ user
                            >
                              &times;
                            </button>
                          )}
                        </li>
                      ))}
                  </ul>
                </div>



              </div>
            )}

            {/* Google Sign-In or Start time & End time Event -> Add to google calendar*/}
            {showSignInSection && (
              <div className="flex justify-between items-center mt-4">
                {sessionId ? (
                  <div>
                    <span className="text-sm text-gray-700">
                      Sign in with Google to create an event
                    </span>
                    <form onSubmit={handleSubmit} className="space-y-4">
                      <div>
                        <label htmlFor="start" className="block text-sm font-medium text-gray-700">
                          Start Time
                        </label>
                        <input
                          type="datetime-local"
                          id="start"
                          name="start"
                          value={formData.start}
                          onChange={handleChange}
                          required
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        />
                      </div>
                      <div>
                        <label htmlFor="end" className="block text-sm font-medium text-gray-700">
                          End Time
                        </label>
                        <input
                          type="datetime-local"
                          id="end"
                          name="end"
                          value={formData.end}
                          onChange={handleChange}
                          required
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                        />
                      </div>
                      <button
                        type="submit"
                        className="w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                      >
                        Create Event
                      </button>
                    </form>
                  </div>
                ) : (
                  <button
                    onClick={handleGoogleSignIn}
                    className="bg-green-600 text-white px-4 py-2 rounded-md text-sm hover:bg-green-700"
                  >
                    Sign in with Google
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

      )}
      {/* Popup เลือกแท็ก */}
      {isTagPopupOpen && (
        <div
          className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50"
          onClick={() => setIsTagPopupOpen(false)} // ปิด Popup เมื่อคลิกข้างนอก
        >
          <div
            className="bg-white rounded-lg p-6 w-1/3"
            onClick={(e) => e.stopPropagation()} // ป้องกันการปิด Popup เมื่อคลิกด้านใน
          >
            <h2 className="text-lg font-bold mb-4">Select Tags</h2>
            <ul className="flex flex-col gap-2">
              {tags
                .filter((tag) =>
                  !notes.find((note) => note.note_id === currentNoteId)?.tags?.some((t) => t.tag_id === tag.tag_id)
                ) // แสดงเฉพาะแท็กที่ยังไม่ได้ติด
                .map((tag) => (
                  <li
                    key={tag.tag_id}
                    className="bg-gray-100 text-gray-700 px-4 py-2 rounded-md cursor-pointer hover:bg-gray-200"
                    onClick={() => {
                      const tagName = tag.tag_name;
                      handleAddTagToNote(currentNoteId!, tag.tag_id, tagName); // ติดแท็ก
                      setIsTagPopupOpen(false); // ปิด Popup
                    }}
                  >
                    {tag.tag_name}
                  </li>
                ))}
            </ul>
            <div className="flex justify-end space-x-4 mt-4">
              <button
                onClick={() => setIsTagPopupOpen(false)} // ปิด Popup
                className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
      {isReminderModalOpen && (
        <ReminderModal
          noteId={currentNoteId!}
          existingReminder={currentReminder!}
          onClose={handleCloseReminderModal}
          setNotes={setNotes} // ส่ง setNotes ไปยัง ReminderModal
        />
      )}

    </div>
  );
}
