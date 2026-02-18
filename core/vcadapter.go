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

// JoinVC joins a voice chat
func (a *VCAdapter) JoinVC(ctx context.Context, chatID int64, file string, video bool) error {
	return a.calls.JoinVC(chatID, file, video)
}

// LeaveVC leaves a voice chat
func (a *VCAdapter) LeaveVC(ctx context.Context, chatID int64, force bool) error {
	return a.calls.LeaveVC(chatID)
}

// ChangeVC changes (skips) the current track
func (a *VCAdapter) ChangeVC(ctx context.Context, chatID int64) error {
	return a.calls.LeaveVC(chatID)
}

// ReplayVC replays the current track
func (a *VCAdapter) ReplayVC(ctx context.Context, chatID int64, file string, video bool) error {
	if err := a.calls.LeaveVC(chatID); err != nil {
		return err
	}
	return a.calls.JoinVC(chatID, file, video)
}
