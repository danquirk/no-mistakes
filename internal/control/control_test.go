package control

import (
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/cimonitor"
	"github.com/kunchenguid/no-mistakes/internal/ipc"
	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestActiveRunIDForHeadRequiresMatchingHead(t *testing.T) {
	active := &ipc.GetActiveRunResult{Run: &ipc.RunInfo{ID: "run-old", Status: types.RunRunning, HeadSHA: "old-head"}}

	if got := ActiveRunIDForHead(active, "new-head"); got != "" {
		t.Fatalf("mismatched active run ID = %q, want empty", got)
	}
	if got := ActiveRunIDForHead(active, "old-head"); got != "run-old" {
		t.Fatalf("matching active run ID = %q, want run-old", got)
	}

	active.Run.Status = types.RunCompleted
	if got := ActiveRunIDForHead(active, "old-head"); got != "" {
		t.Fatalf("terminal active run ID = %q, want empty", got)
	}
}

func TestGateResolution(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		findingsJSON string
		alreadyFixed bool
		wantAction   types.ApprovalAction
		wantIDs      []string
	}{
		{
			name:         "actionable findings are fixed with every finding selected",
			status:       string(types.StepStatusAwaitingApproval),
			findingsJSON: `{"findings":[{"id":"review-1","severity":"warning","description":"design choice","action":"ask-user"},{"id":"review-2","severity":"info","description":"fyi","action":"no-op"}],"summary":"2"}`,
			wantAction:   types.ActionFix,
			wantIDs:      []string{"review-1", "review-2"},
		},
		{
			name:         "only non-actionable findings are approved",
			status:       string(types.StepStatusAwaitingApproval),
			findingsJSON: `{"findings":[{"id":"test-1","severity":"info","description":"fyi","action":"no-op"}],"summary":"1"}`,
			wantAction:   types.ActionApprove,
		},
		{
			name:       "no findings are approved",
			status:     string(types.StepStatusAwaitingApproval),
			wantAction: types.ActionApprove,
		},
		{
			name:         "fix review is approved",
			status:       string(types.StepStatusFixReview),
			findingsJSON: `{"findings":[{"id":"review-1","severity":"warning","description":"still here","action":"ask-user"}],"summary":"1"}`,
			wantAction:   types.ActionApprove,
		},
		{
			name:         "already fixed step is approved",
			status:       string(types.StepStatusAwaitingApproval),
			findingsJSON: `{"findings":[{"id":"review-1","severity":"warning","description":"still here","action":"ask-user"}],"summary":"1"}`,
			alreadyFixed: true,
			wantAction:   types.ActionApprove,
		},
		{
			name:         "actionable findings without ids are approved rather than fixing nothing",
			status:       string(types.StepStatusAwaitingApproval),
			findingsJSON: `{"findings":[{"severity":"warning","description":"no id","action":"ask-user"}],"summary":"1"}`,
			wantAction:   types.ActionApprove,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, ids := GateResolution(tt.status, tt.findingsJSON, tt.alreadyFixed)
			if action != tt.wantAction {
				t.Fatalf("action = %s, want %s", action, tt.wantAction)
			}
			if len(ids) != len(tt.wantIDs) {
				t.Fatalf("ids = %v, want %v", ids, tt.wantIDs)
			}
			for i := range ids {
				if ids[i] != tt.wantIDs[i] {
					t.Fatalf("ids = %v, want %v", ids, tt.wantIDs)
				}
			}
		})
	}
}

func TestCIReadyToMerge(t *testing.T) {
	passedLogs := []string{
		"monitoring CI for PR #42 (timeout: 4h)...",
		cimonitor.ChecksPassedMsg,
	}

	if !CIReadyToMerge(string(types.StepCI), string(types.StepStatusRunning), passedLogs) {
		t.Fatal("expected running CI with passed logs to be ready")
	}
	if CIReadyToMerge(string(types.StepCI), string(types.StepStatusCompleted), passedLogs) {
		t.Fatal("completed CI step should not be treated as early ready")
	}
	if CIReadyToMerge(string(types.StepPR), string(types.StepStatusRunning), passedLogs) {
		t.Fatal("non-CI step should not be ready")
	}
}
