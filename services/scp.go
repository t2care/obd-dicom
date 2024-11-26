package services

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/dimsec"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomcommand"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

type scp struct {
	Port                 int
	listener             net.Listener
	onAssociationRequest func(request *network.AAssociationRQ) bool
	onAssociationRelease func(request *network.AAssociationRQ)
	onCFindRequest       func(request *network.AAssociationRQ, findLevel string, data *media.DcmObj) ([]*media.DcmObj, uint16)
	onCMoveRequest       func(request *network.AAssociationRQ, moveLevel string, data *media.DcmObj, moveDst *network.Destination) ([]string, uint16)
	onCStoreRequest      func(request *network.AAssociationRQ, data *media.DcmObj) uint16
}

// NewSCP - Creates an interface to scu
func NewSCP(port int) *scp {
	media.InitDict()

	return &scp{
		Port: port,
	}
}

func (s *scp) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			slog.Error(err.Error())
			continue
		}
		slog.Info("handleConnection, new connection", "ADDRESS", conn.RemoteAddr())
		go func() {
			if err := s.handleConnection(conn); err != nil {
				slog.Error(err.Error())
			}
		}()
	}
}

func (s *scp) Stop() error {
	return s.listener.Close()
}

func (s *scp) handleConnection(conn net.Conn) (err error) {
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	pdu := network.NewPDUService()
	pdu.SetConn(rw)

	if s.onAssociationRequest != nil {
		pdu.SetOnAssociationRequest(s.onAssociationRequest)
	}
	if s.onAssociationRelease != nil {
		pdu.SetOnAssociationRelease(s.onAssociationRelease)
	}

	var dco, ddo *media.DcmObj
	for err == nil {
		if dco, err = pdu.NextPDU(); dco == nil {
			continue
		}
		command := dco.GetUShort(tags.CommandField)
		status := dicomstatus.Success
		switch command {
		case dicomcommand.CStoreRequest:
			if ddo, err = dimsec.CStoreReadRQ(pdu, dco); err != nil {
				return
			}
			if s.onCStoreRequest != nil {
				status = s.onCStoreRequest(pdu.GetAAssociationRQ(), ddo)
			}
		case dicomcommand.CFindRequest:
			if ddo, err = dimsec.CFindReadRQ(pdu); err != nil {
				return
			}
			if s.onCFindRequest != nil {
				queryLevel := ddo.GetString(tags.QueryRetrieveLevel)
				var results []*media.DcmObj
				results, status = s.onCFindRequest(pdu.GetAAssociationRQ(), queryLevel, ddo)
				for _, result := range results {
					if err = pdu.WriteResp(command, dco, result, dicomstatus.Pending); err != nil {
						return
					}
				}
			}
		case dicomcommand.CMoveRequest:
			if ddo, err = dimsec.CMoveReadRQ(pdu); err != nil {
				return
			}
			if s.onCMoveRequest != nil {
				moveLevel := ddo.GetString(tags.QueryRetrieveLevel)
				dst := &network.Destination{CalledAE: dco.GetString(tags.MoveDestination)}
				var files []string
				files, status = s.onCMoveRequest(pdu.GetAAssociationRQ(), moveLevel, ddo, dst)
				scu := NewSCU(dst)
				scu.onCStoreResult = func(pending, completed, failed uint16) error {
					return pdu.WriteResp(command, dco, ddo, dicomstatus.Pending, completed, failed)
				}
				if err = scu.StoreSCU(files, 0); err != nil {
					status = dicomstatus.CMoveOutOfResourcesUnableToPerformSubOperations
				}
			}
		case dicomcommand.CEchoRequest:
		default:
			return fmt.Errorf("handleConnection, service not implemented: %v", command)
		}
		err = pdu.WriteResp(command, dco, nil, status)
	}
	return
}

func (s *scp) OnAssociationRequest(f func(request *network.AAssociationRQ) bool) {
	s.onAssociationRequest = f
}

func (s *scp) OnAssociationRelease(f func(request *network.AAssociationRQ)) {
	s.onAssociationRelease = f
}

func (s *scp) OnCFindRequest(f func(request *network.AAssociationRQ, findLevel string, data *media.DcmObj) ([]*media.DcmObj, uint16)) {
	s.onCFindRequest = f
}

func (s *scp) OnCMoveRequest(f func(request *network.AAssociationRQ, moveLevel string, data *media.DcmObj, moveDst *network.Destination) ([]string, uint16)) {
	s.onCMoveRequest = f
}

func (s *scp) OnCStoreRequest(f func(request *network.AAssociationRQ, data *media.DcmObj) uint16) {
	s.onCStoreRequest = f
}
