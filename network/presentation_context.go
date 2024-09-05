package network

import (
	"bufio"
	"errors"

	"github.com/t2care/obd-dicom/media"
)

type presentationContext struct {
	ItemType              byte //0x20
	Reserved1             byte
	Length                uint16
	PresentationContextID byte
	Reserved2             byte
	Reserved3             byte
	Reserved4             byte
	AbsSyntax             uidItem
	TrnSyntaxs            []*uidItem
}

// NewPresentationContext - NewPresentationContext
func NewPresentationContext() *presentationContext {
	return &presentationContext{
		ItemType:              0x20,
		PresentationContextID: Uniq8odd(),
	}
}

func (pc *presentationContext) GetPresentationContextID() byte {
	return pc.PresentationContextID
}

func (pc *presentationContext) SetPresentationContextID(id byte) {
	pc.PresentationContextID = id
}

func (pc *presentationContext) GetAbstractSyntax() *uidItem {
	return &pc.AbsSyntax
}

func (pc *presentationContext) SetAbstractSyntax(Abst string) {
	pc.AbsSyntax.SetType(0x30)
	pc.AbsSyntax.SetReserved(0x00)
	pc.AbsSyntax.SetUID(Abst)
	pc.AbsSyntax.SetLength(uint16(len(Abst)))
}

func (pc *presentationContext) AddTransferSyntax(Tran string) {
	TrnSyntax := NewUIDItem(Tran, 0x40)
	pc.TrnSyntaxs = append(pc.TrnSyntaxs, TrnSyntax)
}

func (pc *presentationContext) GetTransferSyntaxes() []*uidItem {
	return pc.TrnSyntaxs
}

func (pc *presentationContext) Size() uint16 {
	pc.Length = 4 + pc.AbsSyntax.GetSize()
	for _, TrnSyntax := range pc.TrnSyntaxs {
		pc.Length += TrnSyntax.GetSize()
	}
	return pc.Length + 4
}

func (pc *presentationContext) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	pc.Size()
	bd.WriteByte(pc.ItemType)
	bd.WriteByte(pc.Reserved1)
	bd.WriteUint16(pc.Length)
	bd.WriteByte(pc.PresentationContextID)
	bd.WriteByte(pc.Reserved2)
	bd.WriteByte(pc.Reserved3)
	bd.WriteByte(pc.Reserved4)
	if err := bd.Send(rw); err != nil {
		return err
	}
	if err := pc.AbsSyntax.Write(rw); err != nil {
		return err
	}
	for _, TrnSyntax := range pc.TrnSyntaxs {
		if err := TrnSyntax.Write(rw); err != nil {
			return err
		}
	}
	return nil
}

func (pc *presentationContext) Read(ms *media.MemoryStream) (err error) {
	if pc.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return pc.ReadDynamic(ms)
}

func (pc *presentationContext) ReadDynamic(ms *media.MemoryStream) (err error) {
	if pc.Reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if pc.Length, err = ms.GetUint16(); err != nil {
		return err
	}
	if pc.PresentationContextID, err = ms.GetByte(); err != nil {
		return err
	}
	if pc.Reserved2, err = ms.GetByte(); err != nil {
		return err
	}
	if pc.Reserved3, err = ms.GetByte(); err != nil {
		return err
	}
	if pc.Reserved4, err = ms.GetByte(); err != nil {
		return err
	}
	if err := pc.AbsSyntax.Read(ms); err != nil {
		return err
	}

	Count := pc.Length - 4 - pc.AbsSyntax.GetSize()
	for Count > 0 {
		var TrnSyntax uidItem
		TrnSyntax.Read(ms)
		Count = Count - TrnSyntax.GetSize()
		if TrnSyntax.GetSize() > 0 {
			pc.TrnSyntaxs = append(pc.TrnSyntaxs, &TrnSyntax)
		}
	}

	if Count == 0 {
		return nil
	}

	return errors.New("pc::ReadDynamic, Count is not zero")
}
