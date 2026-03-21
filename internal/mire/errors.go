package mire

var (
	ErrRecordingDiscarded = RecordingDiscardedError{}
)

type RecordingDiscardedError struct{}

func (RecordingDiscardedError) Error() string {
	return "recording discarded"
}
