package network

import (
	"bufio"

	"github.com/one-byte-data/obd-dicom/media"
)

type uidItem struct {
	itemType  byte
	reserved1 byte
	length    uint16
	uid       string
}

func NewUIDItem(uid string, itemType byte) *uidItem {
	return &uidItem{
		itemType: itemType,
		uid:      uid,
		length:   uint16(len(uid)),
	}
}

func (u *uidItem) GetLength() uint16 {
	return u.length
}

func (u *uidItem) GetReserved() byte {
	return u.reserved1
}

func (u *uidItem) GetSize() uint16 {
	return u.length + 4
}

func (u *uidItem) GetType() byte {
	return u.itemType
}

func (u *uidItem) GetUID() string {
	return u.uid
}

func (u *uidItem) SetReserved(reserve byte) {
	u.reserved1 = reserve
}

func (u *uidItem) SetLength(length uint16) {
	u.length = length
}

func (u *uidItem) SetType(itemType byte) {
	u.itemType = itemType
}

func (u *uidItem) SetUID(uid string) {
	u.uid = uid
}

func (u *uidItem) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	bd.WriteByte(u.itemType)
	bd.WriteByte(u.reserved1)
	bd.WriteUint16(u.length)
	bd.WriteString(u.uid)

	return bd.Send(rw)
}

func (u *uidItem) Read(ms *media.MemoryStream) (err error) {
	if u.itemType, err = ms.GetByte(); err != nil {
		return err
	}
	return u.ReadDynamic(ms)
}

func (u *uidItem) ReadDynamic(ms *media.MemoryStream) (err error) {
	if u.reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if u.length, err = ms.GetUint16(); err != nil {
		return err
	}

	buffer := make([]byte, u.length)
	ms.ReadData(buffer)
	u.uid = string(buffer)

	return
}
