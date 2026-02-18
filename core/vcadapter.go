package core

import (
	"context"
)

// VCAdapter wraps Calls to satisfy utils.VoiceChatManager interface
type VCAdapter struct {
	calls *Calls
}

// NewVCAdapter creates a new adapter
func NewVCAdapter(calls *Calls) *VCAdapter {
	return &VCAdapter{calls: calls}
}

func (a *VCAdapter) JoinVC(ctx context.Context, chatID int64, file string, video bool) error {
	return a.calls.JoinVC(chatID, file, video)
}

func (a *VCAdapter) LeaveVC(ctx context.Context, chatID int64, force bool) error {
	return a.calls.LeaveVC(chatID)
}

func (a *VCAdapter) ChangeVC(ctx context.Context, chatID int64) error {
	return a.calls.LeaveVC(chatID)
}

func (a *VCAdapter) ReplayVC(ctx context.Context, chatID int64, file string, video bool) error {
	if err := a.calls.LeaveVC(chatID); err != nil {
		return err
	}
	return a.calls.JoinVC(chatID, file, video)
}
