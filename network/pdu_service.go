package network

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/one-byte-data/obd-dicom/dictionary/sopclass"
	"github.com/one-byte-data/obd-dicom/dictionary/tags"
	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
	"github.com/one-byte-data/obd-dicom/imp"
	"github.com/one-byte-data/obd-dicom/media"
	"github.com/one-byte-data/obd-dicom/network/pdutype"
)

type PDUService struct {
	AcceptedPresentationContexts []*PresentationContextAccept
	readWriter                   *bufio.ReadWriter
	ms                           media.MemoryStream
	pdutype                      int
	pdulength                    uint32
	AssocRQ                      *AAssociationRQ
	AssocAC                      *AAssociationAC
	AssocRJ                      *AAssociationRJ
	ReleaseRQ                    *AReleaseRQ
	ReleaseRP                    *AReleaseRP
	AbortRQ                      *AAbortRQ
	Pdata                        PDataTF
	Timeout                      int
	OnAssociationRequest         func(request *AAssociationRQ) bool
	OnAssociationRelease         func(request *AAssociationRQ)
}

// NewPDUService - creates a pointer to PDUService
func NewPDUService() *PDUService {
	return &PDUService{
		ms:        media.NewEmptyMemoryStream(),
		AssocRQ:   NewAAssociationRQ(),
		AssocAC:   NewAAssociationAC(),
		AssocRJ:   NewAAssociationRJ(),
		ReleaseRQ: NewAReleaseRQ(),
		ReleaseRP: NewAReleaseRP(),
		AbortRQ:   NewAAbortRQ(),
	}
}

var maxPduLength uint32 = 16384

func (pdu *PDUService) SetConn(rw *bufio.ReadWriter) {
	pdu.readWriter = rw
}

func (pdu *PDUService) GetTransferSyntax(pcid byte) *transfersyntax.TransferSyntax {
	for _, pca := range pdu.AcceptedPresentationContexts {
		if pca.GetPresentationContextID() == pcid {
			return transfersyntax.GetTransferSyntaxFromUID(pca.GetTrnSyntax().GetUID())
		}
	}
	return nil
}

func (pdu *PDUService) SetTimeout(timeout int) {
	pdu.Timeout = timeout
}

func (pdu *PDUService) Connect(IP string, Port string) error {
	conn, err := net.Dial("tcp", IP+":"+Port)
	if err != nil {
		return errors.New("pduservice::Connect - " + err.Error())
	}

	if pdu.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(time.Duration(int32(pdu.Timeout)) * time.Second))
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	pdu.readWriter = rw
	pdu.AssocRQ.SetMaxSubLength(maxPduLength)
	pdu.AssocRQ.SetImpClassUID(imp.GetImpClassUID())
	pdu.AssocRQ.SetImpVersionName(imp.GetImpVersion())

	if err = pdu.AssocRQ.Write(pdu.readWriter); err != nil {
		return err
	}

	pdu.ms = media.NewEmptyMemoryStream()

	if err := pdu.ms.ReadFully(rw, 10); err != nil {
		return err
	}

	ItemType, err := pdu.ms.GetByte()
	if err != nil {
		return err
	}

	if _, err = pdu.ms.GetByte(); err != nil {
		return err
	}

	if pdu.pdulength, err = pdu.ms.GetUint32(); err != nil {
		return err
	}

	switch ItemType {
	case pdutype.AssociationAccept:
		if err := pdu.readPDU(); err != nil {
			return err
		}
		pdu.ms.SetPosition(1)
		pdu.AssocAC.ReadDynamic(pdu.ms)
		if !pdu.interogateAAssociateAC() {
			return errors.New("pduservice::Connect - No accepted presentation contexts found")
		}
		return nil
	case pdutype.AssociationReject:
		if err := pdu.readPDU(); err != nil {
			return err
		}
		pdu.ms.SetPosition(1)
		pdu.AssocRJ.ReadDynamic(pdu.ms)
		return fmt.Errorf("pduservice::Connect - Association rejected - %s", pdu.AssocRJ.GetReason())
	case pdutype.AssociationAbortRequest:
		if err := pdu.readPDU(); err != nil {
			return err
		}
		pdu.ms.SetPosition(1)
		pdu.AbortRQ.ReadDynamic(pdu.ms)
		return fmt.Errorf("pduservice::Connect - Association aborted - %s", pdu.AbortRQ.GetReason())
	default:
		return fmt.Errorf("pduservice::Connect - Corrupt transmision - %b", ItemType)
	}
}

func (pdu *PDUService) Close() {
	pdu.ReleaseRQ.Write(pdu.readWriter)
	pdu.ReleaseRP.Read(pdu.ms)
}

func (pdu *PDUService) NextPDU() (command media.DcmObj, err error) {
	if pdu.Pdata.Buffer != nil {
		pdu.Pdata.Buffer.ClearMemoryStream()
	} else {
		pdu.Pdata.Buffer = media.NewEmptyBufData()
	}

	pdu.Pdata.MsgStatus = 0
	if pdu.Pdata.Length != 0 {
		DCO := media.NewEmptyDCMObj()
		pdu.Pdata.ReadDynamic(pdu.ms)
		if pdu.Pdata.MsgStatus > 0 {
			if !pdu.parseRawVRIntoDCM(DCO) {
				pdu.AbortRQ.Write(pdu.readWriter)
				return nil, errors.New("pduservice::Read - ParseRawVRIntoDCM failed")
			}
			return DCO, nil
		}
	}

	for {
		pdu.ms = media.NewEmptyMemoryStream()

		if err := pdu.ms.ReadFully(pdu.readWriter, 10); err != nil {
			return nil, err
		}

		pdu.ms.SetPosition(0)

		if pdu.pdutype, err = pdu.ms.Get(); err != nil {
			return nil, err
		}

		if _, err = pdu.ms.Get(); err != nil {
			return nil, err
		}

		if pdu.pdulength, err = pdu.ms.GetUint32(); err != nil {
			return nil, err
		}

		switch pdu.pdutype {
		case pdutype.AssocicationRequest:
			if err := pdu.readPDU(); err != nil {
				return nil, err
			}
			if err := pdu.AssocRQ.Read(pdu.ms); err != nil {
				return nil, err
			}
			if err := pdu.interogateAAssociateRQ(pdu.readWriter); err != nil {
				return nil, err
			}
			return nil, nil
		case pdutype.AssociationAccept:
			if err := pdu.readPDU(); err != nil {
				return nil, err
			}
			return nil, nil
		case pdutype.PDUDataTransfer:
			if err := pdu.readPDU(); err != nil {
				return nil, err
			}
			pdu.ms.SetPosition(1)
			if err := pdu.Pdata.ReadDynamic(pdu.ms); err != nil {
				return nil, err
			}
			if pdu.Pdata.MsgStatus > 0 {
				DCO := media.NewEmptyDCMObj()
				if !pdu.parseRawVRIntoDCM(DCO) {
					pdu.AbortRQ.Write(pdu.readWriter)
					return nil, errors.New("pduservice::Read - ParseRawVRIntoDCM failed")
				}
				return DCO, nil
			}
		case pdutype.AssociationReleaseRequest:
			slog.Info("ASSOC-R-RQ:", "CallingAE", pdu.AssocRQ.GetCallingAE(), "CalledAE", pdu.AssocRQ.GetCalledAE())
			if pdu.OnAssociationRelease != nil {
				pdu.OnAssociationRelease(pdu.AssocRQ)
			}
			pdu.ReleaseRQ.ReadDynamic(pdu.ms)
			pdu.ReleaseRP.Write(pdu.readWriter)
			return nil, errors.New("pduservice::Read - A-Release-RQ")
		case pdutype.AssociationReleaseResponse:
			slog.Info("ASSOC-R-RP:", "CallingAE", pdu.AssocRQ.GetCallingAE(), "CalledAE", pdu.AssocRQ.GetCalledAE())
			return nil, errors.New("pduservice::Read - A-Release-RP")
		case pdutype.AssociationAbortRequest:
			slog.Info("ASSOC-ABORT-RQ:", "CallingAE", pdu.AssocRQ.GetCallingAE(), "CalledAE", pdu.AssocRQ.GetCalledAE())
			return nil, errors.New("pduservice::Read - A-Abort-RQ")
		default:
			pdu.AbortRQ.Write(pdu.readWriter)
			return nil, errors.New("pduservice::Read - unknown ItemType")
		}
	}
}

func (pdu *PDUService) GetAAssociationRQ() *AAssociationRQ {
	return pdu.AssocRQ
}

func (pdu *PDUService) GetCalledAE() string {
	return pdu.AssocRQ.GetCalledAE()
}

func (pdu *PDUService) GetCallingAE() string {
	return pdu.AssocRQ.GetCallingAE()
}

func (pdu *PDUService) SetCalledAE(calledAE string) {
	pdu.AssocRQ.SetCalledAE(calledAE)
}

func (pdu *PDUService) SetCallingAE(callingAE string) {
	pdu.AssocRQ.SetCallingAE(callingAE)
}

func (pdu *PDUService) AddPresContexts(presentationContext PresentationContext) {
	pdu.AssocRQ.AddPresContexts(presentationContext)
}

func (pdu *PDUService) GetPresentationContextID() byte {
	return pdu.Pdata.PresentationContextID
}

func (pdu *PDUService) SetOnAssociationRequest(f func(request *AAssociationRQ) bool) {
	pdu.OnAssociationRequest = f
}

func (pdu *PDUService) SetOnAssociationRelease(f func(request *AAssociationRQ)) {
	pdu.OnAssociationRelease = f
}

func (pdu *PDUService) Write(DCO media.DcmObj, ItemType byte) error {
	if pdu.Pdata.Buffer != nil {
		pdu.Pdata.Buffer.ClearMemoryStream()
	} else {
		pdu.Pdata.Buffer = media.NewEmptyBufData()
	}

	if pdu.Pdata.PresentationContextID == 0 {
		return errors.New("pduservice::Write - PresentationContextID==0")
	}

	if !pdu.parseDCMIntoRaw(DCO) {
		return errors.New("pduservice::Write - ParseDCMIntoRaw failed")
	}

	pdu.Pdata.MsgHeader = ItemType
	if pdu.AssocAC.GetUserInformation().GetMaxSubLength().GetMaximumLength() > maxPduLength {
		pdu.AssocAC.SetMaxSubLength(maxPduLength)
	}

	// Fixed MaxLength - 6 20200811
	pdu.Pdata.BlockSize = pdu.AssocAC.GetMaxSubLength() - 6

	if ItemType > 0x00 {
		sopClass := sopclass.GetSOPClassFromUID(DCO.GetString(tags.AffectedSOPClassUID))
		slog.Info("PDU-Service: SOP Class", "UID", sopClass.UID, "Description", sopClass.Description, "CalledAE", pdu.GetCalledAE())
	}

	return pdu.Pdata.Write(pdu.readWriter)
}

func (pdu *PDUService) interogateAAssociateAC() bool {
	var PresentationContextID byte
	TS := ""

	for _, presContextAccept := range pdu.AssocAC.GetPresContextAccepts() {
		if presContextAccept.GetResult() == 0 {
			pdu.AcceptedPresentationContexts = append(pdu.AcceptedPresentationContexts, presContextAccept)
			if len(TS) == 0 {
				TS = presContextAccept.GetTrnSyntax().GetUID()
				PresentationContextID = presContextAccept.GetPresentationContextID()
			}
		}
	}
	if (len(TS) > 0) && (len(pdu.AcceptedPresentationContexts) > 0) {
		pdu.Pdata.PresentationContextID = PresentationContextID
		return true
	}
	return false
}

func (pdu *PDUService) interogateAAssociateRQ(rw *bufio.ReadWriter) error {
	if pdu.OnAssociationRequest == nil || !pdu.OnAssociationRequest(pdu.AssocRQ) {
		pdu.AssocRJ.Set(1, 7)
		return pdu.AssocRJ.Write(rw)
	}

	pdu.AssocAC.SetCalledAE(pdu.AssocRQ.GetCalledAE())
	pdu.AssocAC.SetCallingAE(pdu.AssocRQ.GetCallingAE())
	pdu.AssocAC.SetAppContext(pdu.AssocRQ.GetAppContext())
	pdu.AssocAC.SetUserInformation(pdu.AssocRQ.GetUserInformation())

	slog.Info("ASSOC-RQ:", "CallingAE", pdu.AssocRQ.GetCallingAE(), "CalledAE", pdu.AssocRQ.GetCalledAE())
	slog.Info("ASSOC-RQ:", "ImpClass", pdu.AssocRQ.GetUserInformation().GetImpClass().GetUID())
	slog.Info("ASSOC-RQ:", "ImpVersion", pdu.AssocRQ.GetUserInformation().GetImpVersion().GetUID())
	slog.Info("ASSOC-RQ:", "MaxPDULength", pdu.AssocRQ.GetUserInformation().GetMaxSubLength().GetMaximumLength())
	slog.Info("ASSOC-RQ:", "MaxOpsInvoked", pdu.AssocRQ.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsInvoked(), "MaxOpsPerformed", pdu.AssocRQ.GetUserInformation().GetAsyncOperationWindow().GetMaxNumberOperationsPerformed())

	for presIndex, PresContext := range pdu.AssocRQ.GetPresContexts() {
		slog.Info("ASSOC-RQ: PresentationContext", "Index", presIndex)

		sopClass := sopclass.GetSOPClassFromUID(PresContext.GetAbstractSyntax().GetUID())
		slog.Info("ASSOC-RQ: \tAbstractContext", "UID", sopClass.UID, "Description", sopClass.Description)
		for _, TransferSyn := range PresContext.GetTransferSyntaxes() {
			tsName := ""
			transferSyntax := transfersyntax.GetTransferSyntaxFromUID(TransferSyn.GetUID())
			if transferSyntax != nil {
				tsName = transferSyntax.Description
			}
			slog.Info("ASSOC-RQ: \tTransferSynxtax:", "UID", TransferSyn.GetUID(), "Description", tsName)
		}

		PresContextAccept := NewPresentationContextAccept()
		PresContextAccept.SetResult(4)
		PresContextAccept.SetAbstractSyntax(PresContext.GetAbstractSyntax().GetUID())
		TS := ""
		for _, TrnSyntax := range PresContext.GetTransferSyntaxes() {
			TS = TrnSyntax.GetUID()
			break
		}
		if TS == "" {
			TS = transfersyntax.ImplicitVRLittleEndian.UID
		}

		PresContextAccept.SetResult(0)
		PresContextAccept.SetTransferSyntax(TS)
		PresContextAccept.SetPresentationContextID(PresContext.GetPresentationContextID())
		pdu.AcceptedPresentationContexts = append(pdu.AcceptedPresentationContexts, PresContextAccept)
		pdu.AssocAC.AddPresContextAccept(PresContextAccept)
	}

	if len(pdu.AcceptedPresentationContexts) > 0 {
		MaxSubLength := NewMaximumSubLength()
		UserInfo := NewUserInformation()

		MaxSubLength.SetMaximumLength(maxPduLength)
		UserInfo.SetImpClassUID(imp.GetImpClassUID())
		UserInfo.SetImpVersionName(imp.GetImpVersion())
		UserInfo.SetMaxSubLength(MaxSubLength)
		pdu.AssocAC.SetUserInformation(UserInfo)
		return pdu.AssocAC.Write(rw)
	}

	slog.Info("pduservice::InterogateAAssociateRQ - No valid AcceptedPresentationContexts")
	return pdu.AssocRJ.Write(rw)
}

func (pdu *PDUService) parseDCMIntoRaw(DCO media.DcmObj) bool {
	pdu.Pdata.Buffer.WriteObj(DCO)
	return true
}

func (pdu *PDUService) parseRawVRIntoDCM(DCO media.DcmObj) bool {
	TrnSyntax := pdu.GetTransferSyntax(pdu.Pdata.PresentationContextID)
	if TrnSyntax == nil {
		slog.Info("pduservice::ParseRawVRIntoDCM - Transfer syntax length is 0")
		return false
	}
	DCO.SetTransferSyntax(TrnSyntax)
	pdu.Pdata.Buffer.SetPosition(0)
	return pdu.Pdata.Buffer.ReadObj(DCO) == nil
}

func (pdu *PDUService) readPDU() error {
	return pdu.ms.ReadFully(pdu.readWriter, int(pdu.pdulength)-4)
}
