package core

// The AEMA buffer is the primary persistence interface to store match data.
// The buffer can be implemented as local storage, cloud storage or on a blockchain.
type Buffer interface {
	GetBuffer() map[string]string
	StoreModifiedBuffer()
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

func (b *StorageBuffer) GetBuffer() map[string]string {
	return b.Buffer
}
