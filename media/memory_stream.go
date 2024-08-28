package media

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type MemoryStream struct {
	Data     []byte
	Position int
	Size     int
}

// NewEmptyMemoryStream - Creates an inteface to a new empty memoryStream
func NewEmptyMemoryStream() *MemoryStream {
	return &MemoryStream{
		Data:     make([]byte, 0),
		Position: 0,
		Size:     0,
	}
}

// NewMemoryStreamFromBytes - Creates an interface to a new memoryStream from bytes
func NewMemoryStreamFromBytes(data []byte) *MemoryStream {
	return &MemoryStream{
		Data:     data,
		Position: 0,
		Size:     len(data),
	}
}

// NewMemoryStreamFromFile - Creates an interface to a new memoryStream from file
func NewMemoryStreamFromFile(fileName string) (*MemoryStream, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return &MemoryStream{
		Data:     data,
		Position: 0,
		Size:     len(data),
	}, nil
}

func (ms *MemoryStream) GetByte() (byte, error) {
	if ms.Position >= ms.Size {
		return 0, errors.New("no more data to read")
	}
	b := ms.Data[ms.Position]
	ms.Position++
	return b, nil
}

func (ms *MemoryStream) GetUint16() (uint16, error) {
	if ms.Position+1 >= ms.Size {
		return 0, errors.New("no more data to read")
	}
	b := make([]byte, 2)
	copy(b, ms.Data[ms.Position:ms.Position+2])
	ms.Position += 2
	return binary.BigEndian.Uint16(b), nil
}

func (ms *MemoryStream) GetUint32() (uint32, error) {
	if ms.Position+3 >= ms.Size {
		return 0, errors.New("no more data to read")
	}
	b := make([]byte, 4)
	copy(b, ms.Data[ms.Position:ms.Position+4])
	ms.Position += 4
	return binary.BigEndian.Uint32(b), nil
}

func (ms *MemoryStream) Get() (int, error) {
	if ms.Position >= ms.Size {
		return 0, errors.New("no more data to read")
	}
	b := ms.Data[ms.Position]
	ms.Position++
	return int(b), nil
}

func (ms *MemoryStream) GetInt() (int, error) {
	if ms.Position+3 >= ms.Size {
		return 0, errors.New("no more data to read")
	}
	b := ms.Data[ms.Position : ms.Position+4]
	ms.Position += 4
	return int(binary.BigEndian.Uint32(b)), nil
}

func (ms *MemoryStream) ReadData(dst []byte) error {
	if ms.Position+len(dst) > ms.Size {
		return errors.New("no more data to read")
	}
	copy(dst, ms.Data[ms.Position:ms.Position+len(dst)])
	ms.Position += len(dst)
	return nil
}

func (ms *MemoryStream) ReadFully(rw *bufio.ReadWriter, length int) error {
	data := make([]byte, length)
	if _, err := io.ReadFull(rw, data); err != nil {
		return err
	}
	rw.Flush()
	ms.Data = append(ms.Data, data...)
	ms.Size += length
	return nil
}

func (ms *MemoryStream) GetData() []byte {
	return ms.Data
}

func (ms *MemoryStream) GetPosition() int {
	return ms.Position
}

func (ms *MemoryStream) SetPosition(position int) {
	ms.Position = position
}

func (ms *MemoryStream) GetSize() int {
	return ms.Size
}

func (ms *MemoryStream) SetSize(size int) {
	ms.Size = size
}

// Read - Read from MemoryStream into Buffer count bytes
func (ms *MemoryStream) Read(count int) ([]byte, error) {
	buffer := make([]byte, count)
	if count+ms.Position > ms.Size {
		return nil, errors.New("MemoryStream::Read, count+ms.Position > ms.Size")
	}
	copy(buffer, ms.Data[ms.Position:ms.Position+count])
	ms.Position = ms.Position + count
	return buffer, nil
}

func (ms *MemoryStream) Append(data []byte) (int, error) {
	count := len(data)
	if count == 0 {
		return -1, errors.New("MemoryStream::Append, nothing to write")
	}

	if ms.Data == nil {
		return -1, errors.New("MemoryStream:::Append, no Data to append to")
	}

	ms.Data = append(ms.Data, data...)

	return count, nil
}

// Write - Write from Buffer into MemoryStream count bytes
func (ms *MemoryStream) Write(buffer []byte, count int) (int, error) {
	if len(buffer) == 0 {
		return -1, errors.New("MemoryStream::Write, nothing to write")
	}

	if ms.Data == nil {
		return -1, errors.New("MemoryStream:::Write, no Data to append to")
	}

	if ms.Position >= ms.Size {
		ms.Data = append(ms.Data, buffer...)
		ms.Size = ms.Size + count
	} else {
		temp := ms.Data[:ms.Position]
		temp = append(temp, buffer[:count]...)
		temp = append(temp, ms.Data[ms.Position+count:ms.Size]...)
		copy(ms.Data, temp)
	}
	ms.Position = ms.Position + count
	return count, nil
}

// Clear - Clears the memoryStream
func (ms *MemoryStream) Clear() {
	ms.Data = ms.Data[:0]
	ms.Position = 0
	ms.Size = 0
}
