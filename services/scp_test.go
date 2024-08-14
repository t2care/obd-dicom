package services

import (
	"testing"

	"github.com/one-byte-data/obd-dicom/network"
)

func Test_Association_ID(t *testing.T) {
	_, testSCP := StartSCP(t, 1043)
	var onAssociationRequestID int64
	var onAssociationReleaseID int64
	testSCP.OnAssociationRequest(func(request network.AAssociationRQ) bool {
		onAssociationRequestID = request.GetID()
		return true
	})
	testSCP.OnAssociationRelease(func(request network.AAssociationRQ) {
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
				IsCFind:   true,
				IsCMove:   true,
				IsCStore:  true,
				IsTLS:     false,
			})
			if err := d.EchoSCU(0); (err != nil) != tt.wantErr {
				t.Errorf("scu.EchoSCU() error = %v, wantErr %v", err, tt.wantErr)
			}
			if onAssociationRequestID != onAssociationReleaseID {
				t.Errorf("onAssociationRequestID = %v, onAssociationReleaseID = %v", onAssociationRequestID, onAssociationReleaseID)
			}
		})
	}
}
