package billingdatasource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAddAWSPendingImportWarning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     types.String
		wantWarns  int
		wantDetail string
	}{
		{
			name:      "pending status emits warning",
			status:    types.StringValue(awsBillingDatasourceStatusPending),
			wantWarns: 1,
			wantDetail: awsPendingImportWarningDetail,
		},
		{
			name:      "active status is silent",
			status:    types.StringValue("ACTIVE"),
			wantWarns: 0,
		},
		{
			name:      "null status is silent",
			status:    types.StringNull(),
			wantWarns: 0,
		},
		{
			name:      "unknown status is silent",
			status:    types.StringUnknown(),
			wantWarns: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			addAWSPendingImportWarning(&diags, tt.status)

			if got, want := len(diags.Warnings()), tt.wantWarns; got != want {
				t.Fatalf("unexpected warning count: got %d, want %d", got, want)
			}

			if tt.wantWarns == 0 {
				return
			}

			if got, want := diags.Warnings()[0].Summary(), awsPendingImportWarningSummary; got != want {
				t.Fatalf("unexpected warning summary: got %q, want %q", got, want)
			}

			if got, want := diags.Warnings()[0].Detail(), tt.wantDetail; got != want {
				t.Fatalf("unexpected warning detail: got %q, want %q", got, want)
			}
		})
	}
}
