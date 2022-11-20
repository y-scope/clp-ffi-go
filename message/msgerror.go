package message

// MsgError defines errors created in the go FFI code, but may need to mirror a
// cpp type in the future.
//go:generate stringer -type=MsgError
type MsgError int

const (
	_ MsgError = iota
	DecodeError
	NilRef
)

func (self MsgError) Error() string {
	return self.String()
}
