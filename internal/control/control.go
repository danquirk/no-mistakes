package control

import (
	"github.com/kunchenguid/no-mistakes/internal/cimonitor"
	"github.com/kunchenguid/no-mistakes/internal/ipc"
	"github.com/kunchenguid/no-mistakes/internal/types"
)

// TerminalStatus reports whether a run has reached a final state.
func TerminalStatus(status types.RunStatus) bool {
	switch status {
	case types.RunCompleted, types.RunFailed, types.RunCancelled:
		return true
	default:
		return false
	}
}

// OutcomeFor maps a terminal run status onto the headless command outcome word.
func OutcomeFor(status types.RunStatus) string {
	switch status {
	case types.RunCompleted:
		return "passed"
	case types.RunFailed:
		return "failed"
	case types.RunCancelled:
		return "cancelled"
	default:
		return string(status)
	}
}

// ActiveRunInfoForHead returns run when it is active for headSHA.
func ActiveRunInfoForHead(run *ipc.RunInfo, headSHA string) *ipc.RunInfo {
	if run == nil || TerminalStatus(run.Status) || run.HeadSHA != headSHA {
		return nil
	}
	return run
}

// ActiveRunIDForHead returns the active run ID matching headSHA, or empty.
func ActiveRunIDForHead(active *ipc.GetActiveRunResult, headSHA string) string {
	run := ActiveRunInfoForHead(active.Run, headSHA)
	if run == nil {
		return ""
	}
	return run.ID
}

// CIReadyToMerge reports whether CI is still running but all checks have passed.
func CIReadyToMerge(stepName string, stepStatus string, ciLogs []string) bool {
	return stepName == string(types.StepCI) &&
		stepStatus == string(types.StepStatusRunning) &&
		cimonitor.ChecksPassed(ciLogs)
}

// GateResolution decides how --yes answers an approval gate.
func GateResolution(gateStatus string, findingsJSON string, alreadyFixed bool) (types.ApprovalAction, []string) {
	if alreadyFixed || gateStatus == string(types.StepStatusFixReview) {
		return types.ActionApprove, nil
	}
	parsed, err := types.ParseFindingsJSON(findingsJSON)
	if err != nil || !types.HasActionableFindings(parsed) {
		return types.ActionApprove, nil
	}
	ids := make([]string, 0, len(parsed.Items))
	for _, f := range parsed.Items {
		if f.ID != "" {
			ids = append(ids, f.ID)
		}
	}
	if len(ids) == 0 {
		return types.ActionApprove, nil
	}
	return types.ActionFix, ids
}
