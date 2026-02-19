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
	client *tg.Client
	ntg    *ntg.Client

	activeSessions   map[int64]*VCSession
	activeSessionsMu sync.RWMutex
}

type VCSession struct {
	ChatID    int64
	FilePath  string
	IsVideo   bool
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

func NewCalls(client *tg.Client) *Calls {
	return &Calls{
		client:         client,
		ntg:            ntg.NTgCalls(),
		activeSessions: make(map[int64]*VCSession),
	}
}

func (c *Calls) Start() error {
	log.Println(">> Booting NTgCalls client...")

	c.ntg.OnStreamEnd(func(chatId int64, streamType ntg.StreamType, device ntg.StreamDevice) {
		log.Printf(">> Stream ended for chat %d", chatId)
		c.LeaveVC(chatId)
	})

	c.ntg.OnConnectionChange(func(chatId int64, state ntg.NetworkInfo) {
		log.Printf(">> Connection changed for chat %d", chatId)
	})

	log.Println(">> NTgCalls client booted!")
	return nil
}

func (c *Calls) getSelfPeer() (tg.InputPeer, error) {
	me, err := c.client.GetMe()
	if err != nil {
		return nil, err
	}
	return c.client.ResolvePeer(me.ID)
}

func (c *Calls) JoinVC(chatID int64, filePath string, video bool) error {
	groupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		return err
	}

	// 1️⃣ Create WebRTC offer
	offer, err := c.ntg.CreateCall(chatID)
	if err != nil {
		return fmt.Errorf("CreateCall failed: %w", err)
	}

	joinAs, err := c.getSelfPeer()
	if err != nil {
		return fmt.Errorf("getSelfPeer failed: %w", err)
	}

	log.Printf(">> Joining VC - chatID: %d, file: %s", chatID, filePath)

	// 2️⃣ Join Telegram group call
	result, err := c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Call:         groupCall,
		JoinAs:       joinAs,
		Muted:        false,
		VideoStopped: !video,
		Params:       &tg.DataJson{Data: offer},
	})
	if err != nil {
		c.ntg.Stop(chatID)
		return fmt.Errorf("PhoneJoinGroupCall failed: %w", err)
	}

	// 3️⃣ Extract transport params
	var answer string
	if updates, ok := result.(*tg.UpdatesObj); ok {
		for _, upd := range updates.Updates {
			if conn, ok := upd.(*tg.UpdateGroupCallConnection); ok {
				answer = conn.Params.Data
				break
			}
		}
	}

	if answer == "" {
		return fmt.Errorf("transport params missing from Telegram response")
	}

	log.Printf(">> Got transport answer from Telegram")

	// 4️⃣ Connect NTgCalls with transport answer
	if err := c.ntg.Connect(chatID, answer, false); err != nil {
		return fmt.Errorf("Connect failed: %w", err)
	}

	// 5️⃣ Set stream sources - REAL MediaDescription structure
	// Microphone = audio input, Camera = video input
	media := ntg.MediaDescription{
		Microphone: &ntg.AudioDescription{
			MediaSource:  ntg.MediaSourceFFmpeg,
			Input:        filePath,
			SampleRate:   48000,
			ChannelCount: 2,
		},
	}

	if video {
		media.Camera = &ntg.VideoDescription{
			MediaSource: ntg.MediaSourceFFmpeg,
			Input:       filePath,
			Width:       1280,
			Height:      720,
			Fps:         24,
		}
	}

	if err := c.ntg.SetStreamSources(chatID, ntg.CaptureStream, media); err != nil {
		return fmt.Errorf("SetStreamSources failed: %w", err)
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

	log.Println(">> ✅ Streaming started successfully!")
	return nil
}

func (c *Calls) LeaveVC(chatID int64) error {
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
	_, err := c.ntg.Pause(chatID)
	return err
}

func (c *Calls) ResumeVC(chatID int64) error {
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

// GetInputGroupCall returns *tg.InputGroupCallObj for a chat
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
			return nil, fmt.Errorf("ChannelsGetFullChannel failed: %w", err)
		}
		fullChan, ok := fullChannel.FullChat.(*tg.ChannelFull)
		if !ok {
			return nil, fmt.Errorf("unexpected FullChat type: %T", fullChannel.FullChat)
		}
		if fullChan.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat! Start VC from group settings first.")
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
			return nil, fmt.Errorf("MessagesGetFullChat failed: %w", err)
		}
		chatFull, ok := fullChat.FullChat.(*tg.ChatFullObj)
		if !ok {
			return nil, fmt.Errorf("unexpected FullChat type: %T", fullChat.FullChat)
		}
		if chatFull.Call == nil {
			return nil, fmt.Errorf("❌ No active Voice Chat! Start VC from group settings first.")
		}
		callObj, ok := chatFull.Call.(*tg.InputGroupCallObj)
		if !ok {
			return nil, fmt.Errorf("unexpected Call type: %T", chatFull.Call)
		}
		log.Printf(">> Found group call: ID=%d", callObj.ID)
		return callObj, nil

	default:
		return nil, fmt.Errorf("unsupported peer type: %T", peer)
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
