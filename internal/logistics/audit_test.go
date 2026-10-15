package logistics

import (
	"reflect"
	"testing"
)

func TestAuditKeys(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		wantReach   []string
		wantMissing []string
	}{
		{
			name:        "operations key reaches shipment state but cannot render review",
			scopes:      []string{"logistics.shipment_events.read", "logistics.exceptions.manage"},
			wantReach:   []string{"exception handling", "shipment events"},
			wantMissing: []string{"dated access-review document", "proof-of-delivery files"},
		},
		{
			name:        "review key covers the complete workflow",
			scopes:      []string{"logistics.shipment_events.read", "logistics.pod_files.read", "logistics.exceptions.manage", "pdf.generate"},
			wantReach:   []string{"dated access-review document", "exception handling", "proof-of-delivery files", "shipment events"},
			wantMissing: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AuditKeys([]Key{{ID: "key_1", Name: "dispatch", Scopes: tt.scopes}})[0]
			if !reflect.DeepEqual(got.Reaches, tt.wantReach) || !reflect.DeepEqual(got.Missing, tt.wantMissing) {
				t.Fatalf("got reaches %v missing %v", got.Reaches, got.Missing)
			}
		})
	}
}
