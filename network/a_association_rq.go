package network

import (
	"bufio"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/t2care/obd-dicom/dictionary/sopclass"
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media"
)

type AAssociationRQ struct {
	ItemType        byte // 0x01
	Reserved1       byte
	Length          uint32
	ProtocolVersion uint16 // 0x01
	Reserved2       uint16
	CallingAE       [16]byte // 16 bytes transfered
	CalledAE        [16]byte // 16 bytes transfered
	Reserved3       [32]byte
	AppContext      *uidItem
	PresContexts    []*presentationContext
	UserInfo        *userInformation
	ID              int64
}

// NewAAssociationRQ - NewAAssociationRQ
func NewAAssociationRQ() *AAssociationRQ {
	return &AAssociationRQ{
		ItemType:        0x01,
		Reserved1:       0x00,
		ProtocolVersion: 0x01,
		Reserved2:       0x00,
		AppContext: &uidItem{
			itemType:  0x10,
			reserved1: 0x00,
			uid:       sopclass.DICOMApplicationContext.UID,
			length:    uint16(len(sopclass.DICOMApplicationContext.UID)),
		},
		PresContexts: make([]*presentationContext, 0),
		UserInfo:     NewUserInformation(),
		ID:           time.Now().UnixNano(),
	}
}

func (aarq *AAssociationRQ) GetAppContext() *uidItem {
	return aarq.AppContext
}

func (aarq *AAssociationRQ) SetAppContext(context *uidItem) {
	aarq.AppContext = context
}

func (aarq *AAssociationRQ) GetCallingAE() string {
	temp := []byte{}
	for _, b := range aarq.CallingAE {
		if b != 0x00 && b != 0x20 {
			temp = append(temp, b)
		}
	}
	return string(temp)
}

func (aarq *AAssociationRQ) SetCallingAE(AET string) {
	copy(aarq.CallingAE[:], AET)
	for index, b := range aarq.CallingAE {
		if b == 0x00 {
			aarq.CallingAE[index] = 0x20
		}
	}
}

func (aarq *AAssociationRQ) GetCalledAE() string {
	temp := []byte{}
	for _, b := range aarq.CalledAE {
		if b != 0x00 && b != 0x20 {
			temp = append(temp, b)
		}
	}
	return string(temp)
}

func (aarq *AAssociationRQ) SetCalledAE(AET string) {
	copy(aarq.CalledAE[:], AET)
	for index, b := range aarq.CalledAE {
		if b == 0x00 {
			aarq.CalledAE[index] = 0x20
		}
	}
}

func (aarq *AAssociationRQ) GetPresContexts() []*presentationContext {
	return aarq.PresContexts
}

func (aarq *AAssociationRQ) GetUserInformation() *userInformation {
	return aarq.UserInfo
}

func (aarq *AAssociationRQ) SetUserInformation(userInfo *userInformation) {
	aarq.UserInfo = userInfo
}

func (aarq *AAssociationRQ) GetMaxSubLength() uint32 {
	return aarq.UserInfo.GetMaxSubLength().GetMaximumLength()
}

func (aarq *AAssociationRQ) SetMaxSubLength(length uint32) {
	aarq.UserInfo.GetMaxSubLength().SetMaximumLength(length)
}

func (aarq *AAssociationRQ) GetImpClass() *uidItem {
	return aarq.UserInfo.GetImpClass()
}

func (aarq *AAssociationRQ) SetImpClassUID(uid string) {
	aarq.UserInfo.SetImpClassUID(uid)
}

func (aarq *AAssociationRQ) SetImpVersionName(name string) {
	aarq.UserInfo.SetImpVersionName(name)
}

func (aarq *AAssociationRQ) Size() uint32 {
	aarq.Length = 4 + 16 + 16 + 32
	aarq.Length += uint32(aarq.AppContext.GetSize())

	for _, PresContext := range aarq.PresContexts {
		aarq.Length += uint32(PresContext.Size())
	}

	aarq.Length += uint32(aarq.UserInfo.Size())
	return aarq.Length + 6
}

func (aarq *AAssociationRQ) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	slog.Info("ASSOC-RQ:", "CallingAE", aarq.GetCallingAE(), "CalledAE", aarq.GetCalledAE())
	slog.Info("ASSOC-RQ:", "ImpClass", aarq.GetUserInformation().GetImpClass().GetUID())
	slog.Info("ASSOC-RQ:", "ImpVersion", aarq.GetUserInformation().GetImpVersion().GetUID())
	slog.Info("ASSOC-RQ:", "MaxPDULength", aarq.GetUserInformation().GetMaxSubLength().GetMaximumLength())
	slog.Info("ASSOC-RQ:", "MaxOpsInvoked", aarq.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsInvoked(), "MaxOpsPerformed", aarq.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsPerformed())

	bd.SetBigEndian(true)
	aarq.Size()
	bd.WriteByte(aarq.ItemType)
	bd.WriteByte(aarq.Reserved1)
	bd.WriteUint32(aarq.Length)
	bd.WriteUint16(aarq.ProtocolVersion)
	bd.WriteUint16(aarq.Reserved2)
	bd.Write(aarq.CalledAE[:], 16)
	bd.Write(aarq.CallingAE[:], 16)
	bd.Write(aarq.Reserved3[:], 32)

	if err := bd.Send(rw); err != nil {
		return err
	}

	slog.Info("ASSOC-RQ: ApplicationContext", "UID", aarq.AppContext.GetUID(), "Description", sopclass.GetSOPClassFromUID(aarq.AppContext.GetUID()).Description)
	if err := aarq.AppContext.Write(rw); err != nil {
		return err
	}
	for presIndex, presContext := range aarq.PresContexts {
		slog.Info("ASSOC-RQ: PresentationContext", "Index", presIndex+1)
		slog.Info("ASSOC-RQ: \tAbstractSyntax:", "UID", presContext.GetAbstractSyntax().GetUID(), "Description", sopclass.GetSOPClassFromUID(presContext.GetAbstractSyntax().GetUID()).Description)
		for _, transSyntax := range presContext.GetTransferSyntaxes() {
			slog.Info("ASSOC-RQ: \tTransferSyntax:", "UID", transSyntax.GetUID(), "Description", transfersyntax.GetTransferSyntaxFromUID(transSyntax.GetUID()).Description)
		}
		if err := presContext.Write(rw); err != nil {
			return err
		}
	}
	return aarq.UserInfo.Write(rw)
}

func (aarq *AAssociationRQ) Read(ms *media.MemoryStream) (err error) {
	if aarq.ProtocolVersion, err = ms.GetUint16(); err != nil {
		return err
	}
	if aarq.Reserved2, err = ms.GetUint16(); err != nil {
		return err
	}

	ms.ReadData(aarq.CalledAE[:])
	ms.ReadData(aarq.CallingAE[:])
	ms.ReadData(aarq.Reserved3[:])

	Count := int(ms.GetSize() - 4 - 16 - 16 - 32)
	for Count > 0 {
		TempByte, err := ms.GetByte()
		if err != nil {
			return err
		}

		switch TempByte {
		case 0x10:
			aarq.AppContext.SetType(TempByte)
			aarq.AppContext.ReadDynamic(ms)
			Count = Count - int(aarq.AppContext.GetSize())
		case 0x20:
			PresContext := NewPresentationContext()
			PresContext.ReadDynamic(ms)
			Count = Count - int(PresContext.Size())
			aarq.PresContexts = append(aarq.PresContexts, PresContext)
		case 0x50: // User Information
			aarq.UserInfo.ReadDynamic(ms)
			return nil
		default:
			slog.Error("aarq::ReadDynamic, unknown Item " + strconv.Itoa(int(TempByte)))
			Count = -1
		}
	}

	if Count == 0 {
		return nil
	}

	return errors.New("aarq::ReadDynamic, Count is not zero")
}

func (aarq *AAssociationRQ) AddPresContexts(presentationContext *presentationContext) {
	aarq.PresContexts = append(aarq.PresContexts, presentationContext)
}

func (aarq *AAssociationRQ) GetID() int64 {
	return aarq.ID
}
