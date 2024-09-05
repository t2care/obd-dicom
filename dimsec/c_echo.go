package dimsec

import (
	"errors"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

// CEchoReadRQ CEcho request read
func CEchoReadRQ(DCO *media.DcmObj) bool {
	return DCO.GetUShort(tags.CommandField) == dicomcommand.CEchoRequest
}

// CEchoWriteRQ CEcho request write
func CEchoWriteRQ(pdu *network.PDUService) error {
	DCO := media.NewEmptyDCMObj()

	sopClassUID := ""
	for _, presContext := range pdu.GetAAssociationRQ().GetPresContexts() {
		sopClassUID = presContext.GetAbstractSyntax().GetUID()
	}
	valor := uint16(len(sopClassUID))
	if valor%2 == 1 {
		valor++
	}

	size := uint32(8 + valor + 8 + 2 + 8 + 2 + 8 + 2)

	DCO.WriteUint32(tags.CommandGroupLength, size)
	DCO.WriteString(tags.AffectedSOPClassUID, sopClassUID)
	DCO.WriteUint16(tags.CommandField, dicomcommand.CEchoRequest)
	DCO.WriteUint16(tags.MessageID, network.Uniq16odd())
	DCO.WriteUint16(tags.CommandDataSetType, 0x0101)

	return pdu.Write(DCO, 0x01)
}

// CEchoReadRSP CEcho response read
func CEchoReadRSP(pdu *network.PDUService) error {
	dco, err := pdu.NextPDU()
	if err != nil {
		return errors.New("CEchoReadRSP, failed pdu.Read(&DCO)")
	}
	if dco.GetUShort(tags.CommandField) == dicomcommand.CEchoResponse {
		if dco.GetUShort(tags.Status) == dicomstatus.Success {
			return nil
		}
	}
	return nil
}

// CEchoWriteRSP CEcho response write
func CEchoWriteRSP(pdu *network.PDUService, DCO *media.DcmObj) error {
	DCOR := media.NewEmptyDCMObj()

	DCOR.SetTransferSyntax(DCO.GetTransferSyntax())
	SOPClassUID := DCO.GetString(tags.AffectedSOPClassUID)
	valor := uint16(len(SOPClassUID))
	if valor > 0 {
		if valor%2 == 1 {
			valor++
		}

		size := uint32(8 + valor + 8 + 2 + 8 + 2 + 8 + 2)

		DCOR.WriteUint32(tags.CommandGroupLength, size)
		DCOR.WriteString(tags.AffectedSOPClassUID, SOPClassUID)
		DCOR.WriteUint16(tags.CommandField, dicomcommand.CEchoResponse)
		valor = DCO.GetUShort(tags.MessageID)
		DCOR.WriteUint16(tags.MessageIDBeingRespondedTo, valor)
		valor = DCO.GetUShort(tags.CommandDataSetType)
		DCOR.WriteUint16(tags.CommandDataSetType, valor)
		DCOR.WriteUint16(tags.Status, dicomstatus.Success)
		return pdu.Write(DCOR, 0x01)
	}
	return errors.New("CEchoReadRSP, unknown error")
}
