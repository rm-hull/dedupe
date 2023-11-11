package db

type ScanStatus int64

const (
	InProgress ScanStatus = iota
	Errored
	Completed
	Aborted
)

func (s ScanStatus) String() string {
	switch s {
	case InProgress:
		return "in_progress"
	case Errored:
		return "errore"
	case Completed:
		return "completed"
	case Aborted:
		return "aborted"
	}
	return "unknown"
}
