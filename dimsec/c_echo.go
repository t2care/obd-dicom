package dimsec

import (
	"errors"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

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
