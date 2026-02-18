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
	client     *tg.Client
	binding    *ntgcalls.Binding
	audience   map[int64]int
	audienceMu sync.RWMutex

	p2pConfigs      map[int64]*P2PConfig
	p2pConfigsMutex sync.RWMutex

	inputCalls      map[int64]interface{}
	inputCallsMutex sync.RWMutex

	waitConnect      map[int64]chan error
	waitConnectMutex sync.Mutex

	pendingConnections      map[int64]*PendingConnection
	pendingConnectionsMutex sync.Mutex

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

type P2PConfig struct {
	DhConfig       ntgcalls.DhConfig
	GAorB          []byte
	KeyFingerprint int64
	IsOutgoing     bool
	PhoneCall      *tg.PhoneCallObj
	WaitData       chan error
}

type PendingConnection struct {
	MediaDescription ntgcalls.MediaDescription
	Payload          string
}

func NewCalls(client *tg.Client) *Calls {
	binding, err := ntgcalls.NewBinding()
	if err != nil {
		log.Fatal("Failed to create NTgCalls binding:", err)
	}
	return &Calls{
		client:             client,
		binding:            binding,
		audience:           make(map[int64]int),
		p2pConfigs:         make(map[int64]*P2PConfig),
		inputCalls:         make(map[int64]interface{}),
		waitConnect:        make(map[int64]chan error),
		pendingConnections: make(map[int64]*PendingConnection),
		activeSessions:     make(map[int64]*VCSession),
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

	// Get real WebRTC join params from binding
	joinParams, err := c.binding.CreateCall(chatID)
	if err != nil {
		return fmt.Errorf("failed to create call params: %w", err)
	}

	log.Printf(">> Joining VC - chatID: %d, file: %s, video: %v", chatID, filePath, video)
	log.Printf(">> Join params: %s", joinParams)

	callRes, err := c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Muted:        false,
		VideoStopped: !video,
		Call:         groupCall,
		Params:       &tg.DataJson{Data: joinParams},
	})
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to join group call: %w", err)
	}

	log.Printf(">> Joined VC successfully, result type: %T", callRes)

	// Set stream sources
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

	// Track session
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
	c.audienceMu.Lock()
	delete(c.audience, chatID)
	c.audienceMu.Unlock()

	c.activeSessionsMu.Lock()
	delete(c.activeSessions, chatID)
	c.activeSessionsMu.Unlock()

	// Leave Telegram group call
	groupCall, err := c.GetInputGroupCall(chatID)
	if err == nil {
		c.client.PhoneLeaveGroupCall(tg.InputGroupCall(groupCall), 0)
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

// GetInputGroupCall returns the *tg.InputGroupCallObj for a chat
func (c *Calls) GetInputGroupCall(chatID int64) (*tg.InputGroupCallObj, error) {
	peer, err := c.client.ResolvePeer(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer (chatID: %d): %w", chatID, err)
	}

	log.Printf(">> Resolved peer type: %T for chatID: %d", peer, chatID)

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		fullChannel, err := c.client.ChannelsGetFullChannel(&tg.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get full channel: %w", err)
		}

		fullChan, ok := fullChannel.FullChat.(*tg.ChannelFull)
		if !ok {
			return nil, fmt.Errorf("unexpected FullChat type: %T", fullChannel.FullChat)
		}
		if fullChan.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat!\nStart a Voice Chat from group settings first.")
		}

		callObj, ok := fullChan.Call.(*tg.InputGroupCallObj)
		if !ok {
			return nil, fmt.Errorf("unexpected Call type: %T", fullChan.Call)
		}

		log.Printf(">> Found group call: ID=%d, AccessHash=%d", callObj.ID, callObj.AccessHash)
		return callObj, nil

	case *tg.InputPeerChat:
		fullChat, err := c.client.MessagesGetFullChat(p.ChatID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full chat: %w", err)
		}

		chatFull, ok := fullChat.FullChat.(*tg.ChatFullObj)
		if !ok {
			return nil, fmt.Errorf("unexpected FullChat type: %T", fullChat.FullChat)
		}
		if chatFull.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat!\nStart a Voice Chat from group settings first.")
		}

		callObj, ok := chatFull.Call.(*tg.InputGroupCallObj)
		if !ok {
			return nil, fmt.Errorf("unexpected Call type: %T", chatFull.Call)
		}

		log.Printf(">> Found group call: ID=%d, AccessHash=%d", callObj.ID, callObj.AccessHash)
		return callObj, nil

	default:
		return nil, fmt.Errorf("unsupported peer type: %T (chatID: %d)", peer, chatID)
	}
}

func (c *Calls) Stop() {
	c.binding.Stop(-1)
}
