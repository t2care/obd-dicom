package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media"
	"github.com/t2care/obd-dicom/network"
	"github.com/t2care/obd-dicom/network/dicomstatus"
	"github.com/t2care/obd-dicom/utils"
)

const (
	scp_aet  = "SCP"
	scp_port = 1104
)

var scp_dst = &network.Destination{Port: 1104, CalledAE: scp_aet, CallingAE: "SCU"}

func TestEchoSCU(t *testing.T) {
	assert.Error(t, NewSCU(&network.Destination{Port: scp_port}).EchoSCU(1), "Should not have C-Echo Success")
	assert.NoError(t, NewSCU(scp_dst).EchoSCU(1), "Should have C-Echo Success")
}

func TestFindSCU(t *testing.T) {
	type args struct {
		Query   *media.DcmObj
		timeout int
	}
	tests := []struct {
		name       string
		args       args
		wantResult int
		wantStatus uint16
		wantErr    bool
	}{
		{
			name: "Should C-Find All",
			args: args{
				Query:   utils.DefaultCFindRequest(),
				timeout: 0,
			},
			wantResult: 2,
			wantStatus: dicomstatus.Success,
			wantErr:    false,
		},
		{
			name: "Should find study with date 20050323",
			args: args{
				Query:   cFindReqByDate(),
				timeout: 0,
			},
			wantResult: 1,
			wantStatus: dicomstatus.Success,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []*media.DcmObj
			d := NewSCU(scp_dst)
			d.SetOnCFindResult(func(result *media.DcmObj) {
				results = append(results, result)
			})
			_, status, err := d.FindSCU(tt.args.Query, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("scu.FindSCU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if status != tt.wantStatus {
				t.Errorf("scu.FindSCU() = %v, want %v", status, tt.wantStatus)
			}
			if len(results) != tt.wantResult {
				t.Errorf("scu.FindSCU() count obj= %v, want %v", len(results), tt.wantResult)
			}
		})
	}
}

func cFindReqByDate() *media.DcmObj {
	queryDate := media.NewEmptyDCMObj()
	queryDate.WriteString(tags.QueryRetrieveLevel, "STUDY")
	queryDate.WriteString(tags.StudyDate, "20050323")
	return queryDate
}

func TestStoreSCU(t *testing.T) {
	type args struct {
		FileNames        []string
		transfersyntaxes []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Should store multiples files",
			args: args{
				FileNames:        []string{"../samples/test.dcm", "../samples/test2.dcm"},
				transfersyntaxes: []string{transfersyntax.ExplicitVRLittleEndian.UID},
			},
			wantErr: false,
		},
		{
			name: "Should store lossless",
			args: args{
				FileNames:        []string{"../samples/test-losslessSV1.dcm"},
				transfersyntaxes: []string{transfersyntax.JPEGLosslessSV1.UID},
			},
			wantErr: false,
		},
		{
			name: "Should transcode file to send",
			args: args{
				FileNames:        []string{"../samples/test-losslessSV1.dcm"},
				transfersyntaxes: []string{transfersyntax.ImplicitVRLittleEndian.UID},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewSCU(scp_dst)
			if err := d.StoreSCU(tt.args.FileNames, 0, tt.args.transfersyntaxes...); (err != nil) != tt.wantErr {
				t.Errorf("scu.StoreSCU() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
