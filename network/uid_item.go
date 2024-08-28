package network

import (
	"bufio"

	"github.com/one-byte-data/obd-dicom/media"
)

type UIDItem struct {
	itemType  byte
	reserved1 byte
	length    uint16
	uid       string
}

func NewUIDItem(uid string, itemType byte) *UIDItem {
	return &UIDItem{
		itemType: itemType,
		uid:      uid,
		length:   uint16(len(uid)),
	}
}

func (u *UIDItem) GetLength() uint16 {
	return u.length
}

func (u *UIDItem) GetReserved() byte {
	return u.reserved1
}

func (u *UIDItem) GetSize() uint16 {
	return u.length + 4
}

func (u *UIDItem) GetType() byte {
	return u.itemType
}

func (u *UIDItem) GetUID() string {
	return u.uid
}

func (u *UIDItem) SetReserved(reserve byte) {
	u.reserved1 = reserve
}

func (u *UIDItem) SetLength(length uint16) {
	u.length = length
}

func (u *UIDItem) SetType(itemType byte) {
	u.itemType = itemType
}

func (u *UIDItem) SetUID(uid string) {
	u.uid = uid
}

func (u *UIDItem) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	bd.WriteByte(u.itemType)
	bd.WriteByte(u.reserved1)
	bd.WriteUint16(u.length)
	bd.WriteString(u.uid)

	return bd.Send(rw)
}

func (u *UIDItem) Read(ms *media.MemoryStream) (err error) {
	if u.itemType, err = ms.GetByte(); err != nil {
		return err
	}
	return u.ReadDynamic(ms)
}

func (u *UIDItem) ReadDynamic(ms *media.MemoryStream) (err error) {
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
