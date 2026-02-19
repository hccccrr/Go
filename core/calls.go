package core

import (
	"fmt"
	"log"
	"sync"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	ntg "shizumusic/ntgcalls"
)

type Calls struct {
	client     *tg.Client
	ntg        *ntg.Client
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
	DhConfig       ntg.DhConfig
	GAorB          []byte
	KeyFingerprint int64
	IsOutgoing     bool
	PhoneCall      *tg.PhoneCallObj
	WaitData       chan error
}

type PendingConnection struct {
	Payload string
}

func NewCalls(client *tg.Client) *Calls {
	return &Calls{
		client:             client,
		ntg:                ntg.NTgCalls(),
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

	c.ntg.OnStreamEnd(func(chatId int64, streamType ntg.StreamType, device ntg.StreamDevice) {
		log.Printf(">> Stream ended for chat %d", chatId)
	})

	c.ntg.OnConnectionChange(func(chatId int64, state ntg.NetworkInfo) {
		log.Printf(">> Connection changed for chat %d: %s", chatId, state.Status)
	})

	log.Println(">> NTgCalls client booted!")
	return nil
}

func (c *Calls) getSelfPeer() (tg.InputPeer, error) {
	me, err := c.client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get self: %w", err)
	}
	peer, err := c.client.ResolvePeer(me.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve self peer: %w", err)
	}
	return peer, nil
}

func (c *Calls) JoinVC(chatID int64, filePath string, video bool) error {
	groupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		return err
	}

	// Step 1: Get WebRTC offer from NTgCalls
	joinParams, err := c.ntg.CreateCall(chatID)
	if err != nil {
		return fmt.Errorf("failed to create call: %w", err)
	}

	// Step 2: Get self peer
	joinAs, err := c.getSelfPeer()
	if err != nil {
		return fmt.Errorf("failed to resolve joinAs: %w", err)
	}

	log.Printf(">> Joining VC - chatID: %d, file: %s, video: %v", chatID, filePath, video)

	// Step 3: Join via Telegram
	callRes, err := c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Muted:        false,
		VideoStopped: !video,
		Call:         groupCall,
		JoinAs:       joinAs,
		Params:       &tg.DataJson{Data: joinParams},
	})
	if err != nil {
		c.ntg.Stop(chatID)
		return fmt.Errorf("failed to join group call: %w", err)
	}
	log.Printf(">> Joined VC, result: %T", callRes)

	// Step 4: Extract transport params
	transportParams := "{}"
	if updates, ok := callRes.(*tg.UpdatesObj); ok {
		for _, u := range updates.Updates {
			if connUpdate, ok := u.(*tg.UpdateGroupCallConnection); ok {
				transportParams = connUpdate.Params.Data
				log.Printf(">> Got transport params")
				break
			}
		}
	}

	// Step 5: Connect NTgCalls with Telegram's answer
	if err := c.ntg.Connect(chatID, transportParams, false); err != nil {
		log.Printf(">> Warning: connect failed: %v", err)
	}

	// Step 6: Start streaming
	mediaDesc := ntg.MediaDescription{
		Audio: &ntg.AudioDescription{
			InputMode:     ntg.InputModeFile,
			Input:         filePath,
			SampleRate:    48000,
			BitsPerSample: 16,
			ChannelCount:  2,
		},
	}
	if video {
		mediaDesc.Video = &ntg.VideoDescription{
			InputMode: ntg.InputModeFile,
			Input:     filePath,
			Width:     1280,
			Height:    720,
			Fps:       24,
		}
	}

	if err := c.ntg.SetStreamSources(chatID, ntg.Capture, mediaDesc); err != nil {
		log.Printf(">> Warning: set stream sources failed: %v", err)
	}

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

	groupCall, err := c.GetInputGroupCall(chatID)
	if err == nil {
		c.client.PhoneLeaveGroupCall(tg.InputGroupCall(groupCall), 0)
	}

	return c.ntg.Stop(chatID)
}

func (c *Calls) PauseVC(chatID int64) error {
	c.activeSessionsMu.Lock()
	if s, ok := c.activeSessions[chatID]; ok {
		s.IsPaused = true
	}
	c.activeSessionsMu.Unlock()
	_, err := c.ntg.Pause(chatID)
	return err
}

func (c *Calls) ResumeVC(chatID int64) error {
	c.activeSessionsMu.Lock()
	if s, ok := c.activeSessions[chatID]; ok {
		s.IsPaused = false
	}
	c.activeSessionsMu.Unlock()
	_, err := c.ntg.Resume(chatID)
	return err
}

func (c *Calls) MuteVC(chatID int64) error {
	_, err := c.ntg.Mute(chatID)
	return err
}

func (c *Calls) UnmuteVC(chatID int64) error {
	_, err := c.ntg.Unmute(chatID)
	return err
}

func (c *Calls) GetPing() int64 { return 50 }

func (c *Calls) IsActive(chatID int64) bool {
	c.activeSessionsMu.RLock()
	defer c.activeSessionsMu.RUnlock()
	_, ok := c.activeSessions[chatID]
	return ok
}

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
		log.Printf(">> Found group call: ID=%d", callObj.ID)
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
		log.Printf(">> Found group call: ID=%d", callObj.ID)
		return callObj, nil

	default:
		return nil, fmt.Errorf("unsupported peer type: %T (chatID: %d)", peer, chatID)
	}
}

func (c *Calls) Stop() {
	c.activeSessionsMu.RLock()
	ids := make([]int64, 0, len(c.activeSessions))
	for id := range c.activeSessions {
		ids = append(ids, id)
	}
	c.activeSessionsMu.RUnlock()
	for _, id := range ids {
		c.ntg.Stop(id)
	}
}
