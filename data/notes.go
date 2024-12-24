package data

const (
	TaskNoteType       NoteType     = 1
	AreaNoteType       NoteType     = 2
	PriorityTypeLow    PriorityType = "low"
	PriorityTypeMedium PriorityType = "medium"
	PriorityTypeHigh   PriorityType = "high"
	PriorityTypeUrgent PriorityType = "urgent"
	StatusToDo         StatusType   = "todo"
	StatusPlanning     StatusType   = "planning"
	StatusDoing        StatusType   = "doing"
	StatusDone         StatusType   = "done"
)

type (
	PriorityType string
	StatusType   string
	NoteType     int
)

func (nt NoteType) String() string {
	return [...]string{"Task Note", "Area Note"}[nt]
}
