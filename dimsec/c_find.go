package dimsec

import (
	"errors"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

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
