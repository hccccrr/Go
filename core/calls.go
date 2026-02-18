package core

import (
	"fmt"
	"log"
	"sync"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/ntgcalls"
)

type Calls struct {
	client  *tg.Client
	binding *ntgcalls.Binding

	activeSessions   map[int64]*VCSession
	activeSessionsMu sync.RWMutex
}

type VCSession struct {
	ChatID    int64
	FilePath  string
	IsVideo   bool
	IsPaused  bool
	IsMuted   bool
	StartTime time.Time
}

func NewCalls(client *tg.Client) *Calls {
	binding, err := ntgcalls.NewBinding()
	if err != nil {
		log.Fatal("Failed to create NTgCalls binding:", err)
	}

	return &Calls{
		client:         client,
		binding:        binding,
		activeSessions: make(map[int64]*VCSession),
	}
}

func (c *Calls) Start() error {
	log.Println(">> Booting NTgCalls client...")
	if err := c.binding.Start(); err != nil {
		return fmt.Errorf("failed to start NTgCalls: %w", err)
	}
	log.Println(">> NTgCalls client booted!")
	return nil
}

func (c *Calls) JoinVC(chatID int64, filePath string, video bool) error {
	groupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		return err
	}

	// 🔥 Resolve JoinAs properly (THIS FIXES NIL BUG)
	peer, err := c.client.ResolvePeer(chatID)
	if err != nil {
		return fmt.Errorf("failed to resolve peer: %w", err)
	}

	var joinAs tg.InputPeer
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		joinAs = p
	case *tg.InputPeerChat:
		joinAs = p
	default:
		return fmt.Errorf("unsupported peer type for JoinAs: %T", peer)
	}

	// 🔥 Create WebRTC params
	joinParams, err := c.binding.CreateCall(chatID)
	if err != nil {
		return fmt.Errorf("failed to create call params: %w", err)
	}

	log.Printf(">> Joining VC - chatID: %d, file: %s, video: %v", chatID, filePath, video)

	_, err = c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Muted:        false,
		VideoStopped: !video,
		Call:         groupCall,
		JoinAs:       joinAs, // ✅ IMPORTANT FIX
		Params:       &tg.DataJson{Data: joinParams},
	})
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to join group call: %w", err)
	}

	log.Println(">> Joined VC successfully")

	// Setup stream
	mediaDesc := ntgcalls.MediaDescription{
		Audio: &ntgcalls.AudioDescription{
			InputMode:     ntgcalls.InputModeFFmpeg,
			Input:         filePath,
			SampleRate:    48000,
			BitsPerSample: 16,
			ChannelCount:  2,
		},
	}

	if video {
		mediaDesc.Video = &ntgcalls.VideoDescription{
			InputMode: ntgcalls.InputModeFFmpeg,
			Input:     filePath,
			Width:     1280,
			Height:    720,
			Fps:       24,
		}
	}

	c.binding.SetStreamSources(chatID, ntgcalls.CaptureStream, mediaDesc)

	c.activeSessionsMu.Lock()
	c.activeSessions[chatID] = &VCSession{
		ChatID:    chatID,
		FilePath:  filePath,
		IsVideo:   video,
		StartTime: time.Now(),
	}
	c.activeSessionsMu.Unlock()

	return nil
}

func (c *Calls) LeaveVC(chatID int64) error {
	c.activeSessionsMu.Lock()
	delete(c.activeSessions, chatID)
	c.activeSessionsMu.Unlock()

	groupCall, err := c.GetInputGroupCall(chatID)
	if err == nil {
		_, _ = c.client.PhoneLeaveGroupCall(&tg.PhoneLeaveGroupCallParams{
			Call: groupCall,
		})
	}

	return c.binding.Stop(chatID)
}

func (c *Calls) PauseVC(chatID int64) error {
	c.activeSessionsMu.Lock()
	if s, ok := c.activeSessions[chatID]; ok {
		s.IsPaused = true
	}
	c.activeSessionsMu.Unlock()
	return c.binding.Pause(chatID)
}

func (c *Calls) ResumeVC(chatID int64) error {
	c.activeSessionsMu.Lock()
	if s, ok := c.activeSessions[chatID]; ok {
		s.IsPaused = false
	}
	c.activeSessionsMu.Unlock()
	return c.binding.Resume(chatID)
}

func (c *Calls) MuteVC(chatID int64) error   { return c.binding.Mute(chatID) }
func (c *Calls) UnmuteVC(chatID int64) error { return c.binding.Unmute(chatID) }
func (c *Calls) GetPing() int64              { return c.binding.GetPing() }

func (c *Calls) IsActive(chatID int64) bool {
	c.activeSessionsMu.RLock()
	defer c.activeSessionsMu.RUnlock()
	_, ok := c.activeSessions[chatID]
	return ok
}

func (c *Calls) GetInputGroupCall(chatID int64) (*tg.InputGroupCallObj, error) {
	peer, err := c.client.ResolvePeer(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		fullChannel, err := c.client.ChannelsGetFullChannel(&tg.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, err
		}

		fullChan := fullChannel.FullChat.(*tg.ChannelFull)
		if fullChan.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat! Start one first.")
		}

		return fullChan.Call.(*tg.InputGroupCallObj), nil

	case *tg.InputPeerChat:
		fullChat, err := c.client.MessagesGetFullChat(p.ChatID)
		if err != nil {
			return nil, err
		}

		chatFull := fullChat.FullChat.(*tg.ChatFullObj)
		if chatFull.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat! Start one first.")
		}

		return chatFull.Call.(*tg.InputGroupCallObj), nil

	default:
		return nil, fmt.Errorf("unsupported peer type: %T", peer)
	}
}

func (c *Calls) Stop() {
	c.binding.Stop(-1)
}
