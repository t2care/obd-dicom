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
		go s.handleConnection(conn)
	}
}

func (s *scp) Stop() error {
	return s.listener.Close()
}

func (s *scp) handleConnection(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	pdu := network.NewPDUService()
	pdu.SetConn(rw)

	if s.onAssociationRequest != nil {
		pdu.SetOnAssociationRequest(s.onAssociationRequest)
	}
	if s.onAssociationRelease != nil {
		pdu.SetOnAssociationRelease(s.onAssociationRelease)
	}

	var err error
	var dco *media.DcmObj
	for err == nil {
		dco, err = pdu.NextPDU()
		if dco == nil {
			continue
		}
		command := dco.GetUShort(tags.CommandField)
		switch command {
		case dicomcommand.CStoreRequest:
			ddo, err := dimsec.CStoreReadRQ(pdu, dco)
			if err != nil {
				slog.Error("handleConnection, C-Store failed to read request", "ERROR", err.Error())
				conn.Close()
				return
			}

			if s.onCStoreRequest == nil {
				panic("OnCStoreRequest() not implemented")
			}

			status := s.onCStoreRequest(pdu.GetAAssociationRQ(), ddo)
			if err := dimsec.CStoreWriteRSP(pdu, dco, status); err != nil {
				slog.Error("handleConnection, C-Store failed to write response", "ERROR", err.Error())
				conn.Close()
				return
			}
		case dicomcommand.CFindRequest:
			ddo, err := dimsec.CFindReadRQ(pdu)
			if err != nil {
				slog.Error("handleConnection, C-Find failed to read request!")
				conn.Close()
				return
			}
			queryLevel := ddo.GetString(tags.QueryRetrieveLevel)

			status := dicomstatus.Success

			if s.onCFindRequest == nil {
				panic("OnCFindRequest() not implemented")
			}

			results, status := s.onCFindRequest(pdu.GetAAssociationRQ(), queryLevel, ddo)

			for _, result := range results {
				if err := dimsec.CFindWriteRSP(pdu, dco, result, dicomstatus.Pending); err != nil {
					slog.Error("handleConnection, C-Find failed to write response", "ERROR", err.Error())
					conn.Close()
					return
				}
			}

			if err := dimsec.CFindWriteRSP(pdu, dco, media.NewEmptyDCMObj(), status); err != nil {
				slog.Error("handleConnection, C-Find failed to write response", "ERROR", err.Error())
				conn.Close()
				return
			}

		case dicomcommand.CMoveRequest:
			ddo, err := dimsec.CMoveReadRQ(pdu)
			if err != nil {
				slog.Error("handleConnection, C-Move failed to read request!")
				conn.Close()
				return
			}
			moveLevel := ddo.GetString(tags.QueryRetrieveLevel)

			if s.onCMoveRequest == nil {
				panic("OnCMoveRequest() not implemented")
			}

			dst := &network.Destination{CalledAE: dco.GetString(tags.MoveDestination)}
			files, status := s.onCMoveRequest(pdu.GetAAssociationRQ(), moveLevel, ddo, dst)
			scu := NewSCU(dst)
			scu.onCStoreResult = func(pending, completed, failed uint16) error {
				return dimsec.CMoveWriteRSP(pdu, dco, dicomstatus.Pending, pending, completed, failed)
			}
			if err = scu.StoreSCU(files, 0); err != nil {
				status = dicomstatus.CMoveOutOfResourcesUnableToPerformSubOperations
			}
			if err = dimsec.CMoveWriteRSP(pdu, dco, status, 0, 0, 0); err != nil {
				slog.Error("slog.ErrorhandleConnection, C-Move failed to write response", "ERROR", err.Error())
				conn.Close()
				return
			}
		case dicomcommand.CEchoRequest:
			if dimsec.CEchoReadRQ(dco) {
				if err := dimsec.CEchoWriteRSP(pdu, dco); err != nil {
					slog.Error("handleConnection, C-Echo failed to write response!")
					conn.Close()
					return
				}
			}
		default:
			slog.Error("handleConnection, service not implemented", "COMMAND", command)
			conn.Close()
			return
		}
	}

	if err != nil {
		conn.Close()
	}
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
