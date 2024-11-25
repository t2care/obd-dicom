package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomstatus"
)

func Test_Association_ID(t *testing.T) {
	_, testSCP := StartSCP(t, 1043)
	var onAssociationRequestID int64
	var onAssociationReleaseID int64
	testSCP.OnAssociationRequest(func(request *network.AAssociationRQ) bool {
		onAssociationRequestID = request.GetID()
		return true
	})
	testSCP.OnAssociationRelease(func(request *network.AAssociationRQ) {
		onAssociationReleaseID = request.GetID()
	})
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Asso 1",
			wantErr: false,
		},
		{
			name:    "Asso 2",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewSCU(&network.Destination{
				Name:      "Test Destination",
				CalledAE:  "TEST_SCP",
				CallingAE: "TEST_SCU",
				HostName:  "localhost",
				Port:      1043,
				IsCFind:   false,
				IsCMove:   false,
				IsCStore:  false,
				IsTLS:     false,
			})
			if err := d.EchoSCU(0); (err != nil) != tt.wantErr {
				t.Errorf("scu.EchoSCU() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(100 * time.Millisecond) // wait for association closed
			if onAssociationRequestID != onAssociationReleaseID {
				t.Errorf("onAssociationRequestID = %v, onAssociationReleaseID = %v", onAssociationRequestID, onAssociationReleaseID)
			}
		})
	}
}

func Test_QRSCP(t *testing.T) {
	port := 1044
	_, testSCP := StartSCP(t, port)
	testSCP.OnAssociationRequest(func(request *network.AAssociationRQ) bool { return true })
	testSCP.OnCFindRequest(func(request *network.AAssociationRQ, queryLevel string, query *media.DcmObj) ([]*media.DcmObj, uint16) {
		query.WriteString(tags.PatientName, "123")
		return []*media.DcmObj{query}, dicomstatus.Success
	})
	testSCP.OnCMoveRequest(func(request *network.AAssociationRQ, moveLevel string, query *media.DcmObj, moveDst *network.Destination) ([]string, uint16) {
		moveDst.CallingAE = request.GetCalledAE()
		moveDst.HostName = "127.0.0.1"
		moveDst.Port = 1105
		return []string{"../samples/test-losslessSV1.dcm"}, dicomstatus.Success
	})
	assert.NoError(t, dcmtk_findscu(port), "FindSCU should be ok")
	assert.NoError(t, dcmtk_movescu(port), "MoveSCU should be ok")
}

func dcmtk_findscu(port int) error {
	return exe("findscu", "-d", "-S", "-k", "QueryRetrieveLevel=STUDY", "-k", "PatientName=", "127.0.0.1", strconv.Itoa(port))
}

func dcmtk_movescu(port int) error {
	return exe("movescu", "-d", "-k", "StudyInstanceUID=STUDY", "-aem", "scp", "127.0.0.1", strconv.Itoa(port))
}

func exe(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	fmt.Println(string(out)) // For debug logging
	return nil
}

func StartSCP(t testing.TB, port int) (func(t testing.TB), *scp) {
	testSCP := NewSCP(port)
	go func() {
		if err := testSCP.Start(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // wait for server started
	return func(t testing.TB) {
		if err := testSCP.Stop(); err != nil {
			panic(err)
		}
	}, testSCP
}
