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

// CStoreWriteRQ CStore request write
func CStoreWriteRQ(pdu *network.PDUService, DDO *media.DcmObj) error {
	DCO := media.NewEmptyDCMObj()

	sopClassUID := DDO.GetString(tags.SOPClassUID)

	valor := uint16(len(sopClassUID))
	if valor%2 == 1 {
		valor++
	}

	size := uint32(8 + valor + 8 + 2 + 8 + 2 + 8 + 2)

	SOPInstance := DDO.GetString(tags.SOPInstanceUID)
	length := uint32(len(SOPInstance))
	if length%2 == 1 {
		length++
		size = size + 8 + length
	}

	DCO.WriteUint32(tags.CommandGroupLength, size)
	DCO.WriteString(tags.AffectedSOPClassUID, sopClassUID)
	DCO.WriteUint16(tags.CommandField, dicomcommand.CStoreRequest)
	DCO.WriteUint16(tags.MessageID, network.Uniq16odd())
	DCO.WriteUint16(tags.Priority, priority.Medium)
	DCO.WriteUint16(tags.CommandDataSetType, 0x0102)

	if length > 0 {
		DCO.WriteString(tags.AffectedSOPInstanceUID, SOPInstance)
	}

	if err := pdu.Write(DCO, 0x01); err != nil {
		return err
	}
	return pdu.Write(DDO, 0x00)
}

// CStoreReadRSP CStore response read
func CStoreReadRSP(pdu *network.PDUService) (uint16, error) {
	dco, err := pdu.NextPDU()
	if err != nil {
		return dicomstatus.FailureUnableToProcess, err
	}
	// Is this a C-Store RSP?
	if dco.GetUShort(tags.CommandField) == dicomcommand.CStoreResponse {
		return dco.GetUShort(tags.Status), nil
	}
	return dicomstatus.FailureUnableToProcess, errors.New("CStoreReadRSP, unknown error")
}
