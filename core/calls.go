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
		return err
	}

	joinAs, err := c.getSelfPeer()
	if err != nil {
		return err
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
		return err
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
		return fmt.Errorf("transport params missing")
	}

	// 4️⃣ Connect NTgCalls
	if err := c.ntg.Connect(chatID, answer, false); err != nil {
		return err
	}

	// 5️⃣ Start streaming (REAL FIX)
	media := ntg.MediaDescription{
		Audio: &ntg.AudioDescription{
			InputMode:     ntg.InputModeFFmpeg,
			Input:         filePath,
			SampleRate:    48000,
			BitsPerSample: 16,
			ChannelCount:  2,
		},
	}

	if video {
		media.Video = &ntg.VideoDescription{
			InputMode: ntg.InputModeFFmpeg,
			Input:     filePath,
			Width:     1280,
			Height:    720,
			Fps:       24,
		}
	}

	if err := c.ntg.SetStreamSources(chatID, ntg.CaptureStream, media); err != nil {
		return err
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

	log.Println(">> Streaming started successfully")

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

func (c *Calls) IsActive(chatID int64) bool {
	c.activeSessionsMu.RLock()
	defer c.activeSessionsMu.RUnlock()
	_, ok := c.activeSessions[chatID]
	return ok
}

func (c *Calls) Stop() {
	c.activeSessionsMu.RLock()
	defer c.activeSessionsMu.RUnlock()

	for id := range c.activeSessions {
		c.ntg.Stop(id)
	}
}
