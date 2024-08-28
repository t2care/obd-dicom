package network

import (
	"bufio"
	"errors"
	"log/slog"
	"strconv"

	"github.com/one-byte-data/obd-dicom/dictionary/sopclass"
	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
	"github.com/one-byte-data/obd-dicom/media"
)

type AAssociationAC struct {
	ItemType           byte
	Reserved1          byte
	Length             uint32
	ProtocolVersion    uint16
	Reserved2          uint16
	CallingAE          [16]byte
	CalledAE           [16]byte
	Reserved3          [32]byte
	AppContext         *UIDItem
	PresContextAccepts []*PresentationContextAccept
	UserInfo           *UserInformation
}

// NewAAssociationAC NewAAssociationAC
func NewAAssociationAC() *AAssociationAC {
	return &AAssociationAC{
		ItemType:        0x02,
		Reserved1:       0x00,
		ProtocolVersion: 0x01,
		Reserved2:       0x00,
		AppContext: &UIDItem{
			itemType:  0x10,
			reserved1: 0x00,
			uid:       sopclass.DICOMApplicationContext.UID,
			length:    uint16(len(sopclass.DICOMApplicationContext.UID)),
		},
		PresContextAccepts: make([]*PresentationContextAccept, 0),
		UserInfo:           NewUserInformation(),
	}
}

func (aaac *AAssociationAC) GetAppContext() *UIDItem {
	return aaac.AppContext
}

func (aaac *AAssociationAC) SetAppContext(context *UIDItem) {
	aaac.AppContext = context
}

func (aaac *AAssociationAC) GetCallingAE() string {
	temp := []byte{}
	for _, b := range aaac.CallingAE {
		if b != 0x00 && b != 0x20 {
			temp = append(temp, b)
		}
	}
	return string(temp)
}

func (aaac *AAssociationAC) SetCallingAE(AET string) {
	copy(aaac.CallingAE[:], AET)
	for index, b := range aaac.CallingAE {
		if b == 0x00 {
			aaac.CallingAE[index] = 0x20
		}
	}
}

func (aaac *AAssociationAC) GetCalledAE() string {
	temp := []byte{}
	for _, b := range aaac.CalledAE {
		if b != 0x00 && b != 0x20 {
			temp = append(temp, b)
		}
	}
	return string(temp)
}

func (aaac *AAssociationAC) SetCalledAE(AET string) {
	copy(aaac.CalledAE[:], AET)
	for index, b := range aaac.CalledAE {
		if b == 0x00 {
			aaac.CalledAE[index] = 0x20
		}
	}
}

func (aaac *AAssociationAC) AddPresContextAccept(context *PresentationContextAccept) {
	aaac.PresContextAccepts = append(aaac.PresContextAccepts, context)
}

func (aaac *AAssociationAC) GetPresContextAccepts() []*PresentationContextAccept {
	return aaac.PresContextAccepts
}

func (aaac *AAssociationAC) GetUserInformation() *UserInformation {
	return aaac.UserInfo
}

func (aaac *AAssociationAC) SetUserInformation(UserInfo *UserInformation) {
	aaac.UserInfo = UserInfo
}

func (aaac *AAssociationAC) GetMaxSubLength() uint32 {
	return aaac.UserInfo.GetMaxSubLength().GetMaximumLength()
}

func (aaac *AAssociationAC) SetMaxSubLength(length uint32) {
	aaac.UserInfo.GetMaxSubLength().SetMaximumLength(length)
}

func (aaac *AAssociationAC) Size() uint32 {
	aaac.Length = 4 + 16 + 16 + 32
	aaac.Length += uint32(aaac.AppContext.GetSize())

	for _, PresContextAccept := range aaac.PresContextAccepts {
		aaac.Length += uint32(PresContextAccept.Size())
	}

	aaac.Length += uint32(aaac.UserInfo.Size())
	return aaac.Length + 6
}

func (aaac *AAssociationAC) Write(rw *bufio.ReadWriter) error {
	bd := media.NewEmptyBufData()

	slog.Info("ASSOC-AC:", "CallingAE", aaac.GetCallingAE(), "CalledAE", aaac.GetCalledAE())
	slog.Info("ASSOC-AC:", "ImpClass", aaac.UserInfo.GetImpClass().GetUID())
	slog.Info("ASSOC-AC:", "ImpVersion", aaac.UserInfo.GetImpVersion().GetUID())
	slog.Info("ASSOC-AC:", "MaxPDULength", aaac.GetUserInformation().GetMaxSubLength().GetMaximumLength())
	slog.Info("ASSOC-AC:", "MaxOpsInvoked", aaac.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsInvoked(), "MaxOpsPerformed", aaac.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsPerformed())

	bd.SetBigEndian(true)
	aaac.Size()
	bd.WriteByte(aaac.ItemType)
	bd.WriteByte(aaac.Reserved1)
	bd.WriteUint32(aaac.Length)
	bd.WriteUint16(aaac.ProtocolVersion)
	bd.WriteUint16(aaac.Reserved2)
	bd.Write(aaac.CalledAE[:], 16)
	bd.Write(aaac.CallingAE[:], 16)
	bd.Write(aaac.Reserved3[:], 32)

	if err := bd.Send(rw); err != nil {
		return err
	}

	if err := aaac.AppContext.Write(rw); err != nil {
		return err
	}
	for _, presContextAccept := range aaac.PresContextAccepts {
		if err := presContextAccept.Write(rw); err != nil {
			return err
		}
	}
	return aaac.UserInfo.Write(rw)
}

func (aaac *AAssociationAC) Read(ms *media.MemoryStream) (err error) {
	if aaac.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return aaac.ReadDynamic(ms)
}

func (aaac *AAssociationAC) ReadDynamic(ms *media.MemoryStream) (err error) {
	if aaac.Reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if aaac.Length, err = ms.GetUint32(); err != nil {
		return err
	}
	if aaac.ProtocolVersion, err = ms.GetUint16(); err != nil {
		return err
	}
	if aaac.Reserved2, err = ms.GetUint16(); err != nil {
		return err
	}

	ms.ReadData(aaac.CalledAE[:])
	ms.ReadData(aaac.CallingAE[:])
	ms.ReadData(aaac.Reserved3[:])

	Count := int(aaac.Length - 4 - 16 - 16 - 32)

	for Count > 0 {
		TempByte, err := ms.GetByte()
		if err != nil {
			return err
		}

		switch TempByte {
		case 0x10:
			aaac.AppContext.ReadDynamic(ms)
			Count = Count - int(aaac.AppContext.GetSize())
		case 0x21:
			PresContextAccept := NewPresentationContextAccept()
			PresContextAccept.ReadDynamic(ms)
			Count = Count - int(PresContextAccept.Size())
			aaac.PresContextAccepts = append(aaac.PresContextAccepts, PresContextAccept)
		case 0x50: // User Information
			aaac.UserInfo.ReadDynamic(ms)
			Count = Count - int(aaac.UserInfo.Size())
		default:
			Count = -1
			return errors.New("aaac::ReadDynamic, unknown Item " + strconv.Itoa(int(TempByte)))
		}
	}

	slog.Info("ASSOC-AC:", "CallingAE", aaac.GetCallingAE(), "CalledAE", aaac.GetCalledAE())
	slog.Info("ASSOC-AC:", "ImpClass", aaac.GetUserInformation().GetImpClass().GetUID())
	slog.Info("ASSOC-AC:", "ImpVersion", aaac.GetUserInformation().GetImpVersion().GetUID())
	slog.Info("ASSOC-AC:", "MaxPDULength", aaac.GetUserInformation().GetMaxSubLength().GetMaximumLength())
	slog.Info("ASSOC-AC:", "MaxOpsInvoked", aaac.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsInvoked(), "MaxOpsPerformed", aaac.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsPerformed())
	slog.Info("ASSOC-AC: ApplicationContext", "UID", aaac.AppContext.GetUID(), "Description", sopclass.GetSOPClassFromUID(aaac.AppContext.GetUID()).Description)
	for presIndex, presContextAccept := range aaac.PresContextAccepts {
		slog.Info("ASSOC-AC: AcceptedPresentationContext", "Index", presIndex+1)
		//slog.Info("ASSOC-AC: \tAccepted AbstractSyntax", "UID", presContextAccept.GetAbstractSyntax().GetUID(), "Description", sopclass.GetSOPClassFromUID(presContextAccept.GetAbstractSyntax().GetUID()).Description)
		slog.Info("ASSOC-AC: \tAccepted TransferSyntax", "UID", presContextAccept.GetTrnSyntax().GetUID(), "Description", transfersyntax.GetTransferSyntaxFromUID(presContextAccept.GetTrnSyntax().GetUID()).Description)
	}
	if Count == 0 {
		return nil
	}

	return errors.New("aarq::ReadDynamic, Count is not zero")
}
