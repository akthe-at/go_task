package data

import "fmt"

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

// StringToPriorityType converts a string to a PriorityType by mapping the string to a PriorityType
func StringToPriorityType(s string) (PriorityType, error) {
	switch s {
	case string(PriorityTypeLow):
		return PriorityTypeLow, nil
	case string(PriorityTypeMedium):
		return PriorityTypeMedium, nil
	case string(PriorityTypeHigh):
		return PriorityTypeHigh, nil
	case string(PriorityTypeUrgent):
		return PriorityTypeUrgent, nil
	default:
		return "", fmt.Errorf("invalid priority type: %s", s)
	}
}

// StringToStatusType converts a string to a StatusType by mapping a string to a StatusType
func StringToStatusType(input string) (StatusType, error) {
	switch input {
	case "todo":
		return StatusToDo, nil
	case "planning":
		return StatusPlanning, nil
	case "doing":
		return StatusDoing, nil
	case "done":
		return StatusDone, nil
	default:
		return "", fmt.Errorf("invalid status type ( %v ) is not one of the valid status values (todo, planning, doing, done)", input)
	}
}
