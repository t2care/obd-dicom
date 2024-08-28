package network

import (
	"bufio"

	"github.com/one-byte-data/obd-dicom/media"
)

type RoleSelect struct {
	ItemType  byte //0x54
	Reserved1 byte
	Length    uint16
	SCURole   byte
	SCPRole   byte
	uid       string
}

// NewRoleSelect - NewRoleSelect
func NewRoleSelect() *RoleSelect {
	return &RoleSelect{
		ItemType: 0x54,
	}
}

func (scpscu *RoleSelect) Size() uint16 {
	return scpscu.Length + 4
}

func (scpscu *RoleSelect) Write(rw *bufio.ReadWriter) bool {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	bd.WriteByte(scpscu.ItemType)
	bd.WriteByte(scpscu.Reserved1)
	bd.WriteUint16(scpscu.Length)
	bd.WriteUint16(uint16(len(scpscu.uid)))
	bd.Write([]byte(scpscu.uid), len(scpscu.uid))
	bd.WriteByte(scpscu.SCURole)
	bd.WriteByte(scpscu.SCPRole)

	if err := bd.Send(rw); err != nil {
		return false
	}
	return true
}

func (scpscu *RoleSelect) Read(ms *media.MemoryStream) (err error) {
	if scpscu.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return scpscu.ReadDynamic(ms)
}

func (scpscu *RoleSelect) ReadDynamic(ms *media.MemoryStream) (err error) {
	if scpscu.Reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if scpscu.Length, err = ms.GetUint16(); err != nil {
		return err
	}
	tl, err := ms.GetUint16()
	if err != nil {
		return err
	}

	tuid := make([]byte, tl)
	ms.ReadData(tuid)

	scpscu.uid = string(tuid)
	if scpscu.SCURole, err = ms.GetByte(); err != nil {
		return err
	}
	if scpscu.SCPRole, err = ms.GetByte(); err != nil {
		return err
	}
	return
}
