package dimsec

import (
	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

// CMoveReadRSP CMove response read
func CMoveReadRSP(pdu *network.PDUService, pending *int) (*media.DcmObj, uint16, error) {
	status := dicomstatus.FailureUnableToProcess
	dco, err := pdu.NextPDU()
	if err != nil {
		return nil, dicomstatus.FailureUnableToProcess, err
	}

	if dco.GetUShort(tags.CommandField) == dicomcommand.CMoveResponse {
		if dco.GetUShort(tags.CommandDataSetType) != 0x0101 {
			ddo, err := pdu.NextPDU()
			if err != nil {
				return nil, dicomstatus.FailureUnableToProcess, err
			}
			status = dco.GetUShort(tags.Status)
			*pending = int(dco.GetUShort(tags.NumberOfRemainingSuboperations))
			return ddo, status, nil
		}
		status = dco.GetUShort(tags.Status)
		*pending = -1
	}

	return nil, status, nil
}
