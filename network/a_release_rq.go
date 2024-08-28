package network

import (
	"bufio"

	"github.com/one-byte-data/obd-dicom/media"
)

type AReleaseRQ struct {
	ItemType  byte // 0x05
	Reserved1 byte
	Length    uint32
	Reserved2 uint32
}

// NewAReleaseRQ NewAReleaseRQ
func NewAReleaseRQ() *AReleaseRQ {
	return &AReleaseRQ{
		ItemType:  0x05,
		Reserved1: 0x00,
		Reserved2: 0x00,
	}
}

func (arrq *AReleaseRQ) Size() uint32 {
	arrq.Length = 4
	return arrq.Length + 6
}

func (arrq *AReleaseRQ) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	arrq.Size()
	bd.WriteByte(arrq.ItemType)
	bd.WriteByte(arrq.Reserved1)
	bd.WriteUint32(arrq.Length)
	bd.WriteUint32(arrq.Reserved2)

	return bd.Send(rw)
}

func (arrq *AReleaseRQ) Read(ms media.MemoryStream) (err error) {
	if arrq.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return arrq.ReadDynamic(ms)
}

func (arrq *AReleaseRQ) ReadDynamic(ms media.MemoryStream) (err error) {
	if arrq.Reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if arrq.Length, err = ms.GetUint32(); err != nil {
		return err
	}
	if arrq.Reserved2, err = ms.GetUint32(); err != nil {
		return err
	}
	return
}
