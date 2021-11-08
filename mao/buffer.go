package mao

// The AEMA buffer is the primary persistence interface to store match data.
// The buffer can be implemented as local storage, cloud storage or on a blockchain.
type Buffer interface {
	// GetBuffer can return an error if the buffer is busy writing
	// elements. See BusyBuffer error.
	GetBuffer() (map[string]string, error)
	// PutElements stores given key-value pairs. Existing keys will be
	// overridden, non-existing keys will be created.
	PutElements(kv map[string]string) error
	// DeleteElements removes given keys from buffer. Does nothing
	// if key doesn't exist.
	DeleteElements(keys ...string) error
}

type BusyBuffer struct {
	msg string
}

func (e *BusyBuffer) Error() string {
	return "buffer is busy: " + e.msg
}

// Implements an AEMA buffer as local, in-memory storage
type StorageBuffer struct {
	Buffer map[string]string
}

func CreateStorageBuffer() StorageBuffer {
	sb := StorageBuffer{}
	sb.Buffer = make(map[string]string)
	return sb
}

func (b *StorageBuffer) GetBuffer() (map[string]string, error) {
	return b.Buffer, nil
}
