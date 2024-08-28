package network

import (
	"bufio"
	"log/slog"

	"github.com/one-byte-data/obd-dicom/dictionary/sopclass"
	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
	"github.com/one-byte-data/obd-dicom/media"
)

type PresentationContextAccept struct {
	ItemType              byte //0x21
	Reserved1             byte
	Length                uint16
	PresentationContextID byte
	Reserved2             byte
	Result                byte
	Reserved4             byte
	AbsSyntax             UIDItem
	TrnSyntax             UIDItem
}

// NewPresentationContextAccept creates a PresentationContextAccept
func NewPresentationContextAccept() *PresentationContextAccept {
	return &PresentationContextAccept{
		ItemType:              0x21,
		PresentationContextID: Uniq8(),
		Result:                2,
	}
}

func (pc *PresentationContextAccept) GetPresentationContextID() byte {
	return pc.PresentationContextID
}

func (pc *PresentationContextAccept) SetPresentationContextID(id byte) {
	pc.PresentationContextID = id
}

func (pc *PresentationContextAccept) GetResult() byte {
	return pc.Result
}

func (pc *PresentationContextAccept) SetResult(result byte) {
	pc.Result = result
}

func (pc *PresentationContextAccept) GetTrnSyntax() *UIDItem {
	return &pc.TrnSyntax
}

// Size gets the size of presentation
func (pc *PresentationContextAccept) Size() uint16 {
	pc.Length = 4
	pc.Length += pc.TrnSyntax.GetSize()
	return pc.Length + 4
}

func (pc *PresentationContextAccept) GetAbstractSyntax() *UIDItem {
	return &pc.AbsSyntax
}

func (pc *PresentationContextAccept) SetAbstractSyntax(Abst string) {
	pc.AbsSyntax.SetType(0x30)
	pc.AbsSyntax.SetReserved(0x00)
	pc.AbsSyntax.SetUID(Abst)
	pc.AbsSyntax.SetLength(uint16(len(Abst)))
}

func (pc *PresentationContextAccept) SetTransferSyntax(Tran string) {
	pc.TrnSyntax.SetType(0x40)
	pc.TrnSyntax.SetReserved(0)
	pc.TrnSyntax.SetUID(Tran)
	pc.TrnSyntax.SetLength(uint16(len(Tran)))
}

func (pc *PresentationContextAccept) Write(rw *bufio.ReadWriter) (err error) {
	bd := media.NewEmptyBufData()

	bd.SetBigEndian(true)
	pc.Size()
	bd.WriteByte(pc.ItemType)
	bd.WriteByte(pc.Reserved1)
	bd.WriteUint16(pc.Length)
	bd.WriteByte(pc.PresentationContextID)
	bd.WriteByte(pc.Reserved2)
	bd.WriteByte(pc.Result)
	bd.WriteByte(pc.Reserved4)

	sopName := ""
	tsName := ""
	if sopClass := sopclass.GetSOPClassFromUID(pc.GetAbstractSyntax().GetUID()); sopClass != nil {
		sopName = sopClass.Description
	}
	if transferSyntax := transfersyntax.GetTransferSyntaxFromUID(pc.GetTrnSyntax().GetUID()); transferSyntax != nil {
		tsName = transferSyntax.Description
	}

	slog.Info("ASSOC-AC: \tAccepted AbstractContext:", "UID", pc.GetAbstractSyntax().GetUID(), "Description", sopName)
	slog.Info("ASSOC-AC: \tAccepted TransferSynxtax:", "UID", pc.GetTrnSyntax().GetUID(), "Description", tsName)

	if err = bd.Send(rw); err == nil {
		return pc.TrnSyntax.Write(rw)
	}
	return
}

func (pc *PresentationContextAccept) Read(ms *media.MemoryStream) (err error) {
	if pc.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return pc.ReadDynamic(ms)
}

func (pc *PresentationContextAccept) ReadDynamic(ms *media.MemoryStream) (err error) {
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
	if pc.Result, err = ms.GetByte(); err != nil {
		return err
	}
	if pc.Reserved4, err = ms.GetByte(); err != nil {
		return err
	}
	return pc.TrnSyntax.Read(ms)
}
