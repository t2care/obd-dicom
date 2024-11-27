package dimsec

import (
	"errors"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
	"github.com/t2care/obd-dicom/network/priority"
)

// CFindWriteRQ CFind request write
func CFindWriteRQ(pdu *network.PDUService, DDO *media.DcmObj) error {
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
	DCO.WriteUint16(tags.CommandField, dicomcommand.CFindRequest)
	DCO.WriteUint16(tags.MessageID, network.Uniq16odd())
	DCO.WriteUint16(tags.Priority, priority.Medium)
	DCO.WriteUint16(tags.CommandDataSetType, 0x0102)

	if err := pdu.Write(DCO, 0x01); err != nil {
		return err
	}
	return pdu.Write(DDO, 0x00)
}

// CFindReadRSP CFind response read
func CFindReadRSP(pdu *network.PDUService) (*media.DcmObj, uint16, error) {
	dco, err := pdu.NextPDU()
	if err != nil {
		return nil, dicomstatus.FailureUnableToProcess, err
	}

	// Is this a C-Find RSP?
	if dco.GetUShort(tags.CommandField) == dicomcommand.CFindResponse {
		if dco.GetUShort(tags.CommandDataSetType) != 0x0101 {
			ddo, err := pdu.NextPDU()
			if err != nil {
				return nil, dicomstatus.FailureUnableToProcess, err
			}
			return ddo, dco.GetUShort(tags.Status), nil
		}
		return nil, dco.GetUShort(tags.Status), nil
	}
	return nil, dicomstatus.FailureUnableToProcess, errors.New("CFindReadRSP, unknown error")
}
