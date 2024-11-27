package dimsec

import (
	"errors"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

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
