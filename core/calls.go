package core

import (
	"fmt"
	"log"
	"sync"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/ntgcalls"
)

// Calls handles voice chat operations
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
	mediaDesc := ntgcalls.MediaDescription{
		Audio: &ntgcalls.AudioDescription{
			InputMode:     ntgcalls.InputModeFile,
			Input:         filePath,
			SampleRate:    48000,
			BitsPerSample: 16,
			ChannelCount:  2,
		},
	}

	if video {
		mediaDesc.Video = &ntgcalls.VideoDescription{
			InputMode: ntgcalls.InputModeFile,
			Input:     filePath,
			Width:     1280,
			Height:    720,
			Fps:       24,
		}
	}

	return c.connectCall(chatID, mediaDesc, "")
}

func (c *Calls) LeaveVC(chatID int64) error {
	c.audienceMu.Lock()
	delete(c.audience, chatID)
	c.audienceMu.Unlock()
	return c.binding.Stop(chatID)
}

func (c *Calls) PauseVC(chatID int64) error  { return c.binding.Pause(chatID) }
func (c *Calls) ResumeVC(chatID int64) error { return c.binding.Resume(chatID) }
func (c *Calls) MuteVC(chatID int64) error   { return c.binding.Mute(chatID) }
func (c *Calls) UnmuteVC(chatID int64) error { return c.binding.Unmute(chatID) }
func (c *Calls) GetPing() int64              { return c.binding.GetPing() }

func (c *Calls) connectCall(chatID int64, mediaDesc ntgcalls.MediaDescription, jsonParams string) error {
	c.waitConnectMutex.Lock()
	waitChan := make(chan error)
	c.waitConnect[chatID] = waitChan
	c.waitConnectMutex.Unlock()

	defer func() {
		c.waitConnectMutex.Lock()
		delete(c.waitConnect, chatID)
		c.waitConnectMutex.Unlock()
	}()

	return c.handleGroupCall(chatID, mediaDesc, jsonParams, waitChan)
}

func (c *Calls) handleGroupCall(chatID int64, mediaDesc ntgcalls.MediaDescription, jsonParams string, waitChan chan error) error {
	var err error

	jsonParams, err = c.binding.CreateCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to create call: %w", err)
	}

	if err := c.binding.SetStreamSources(chatID, ntgcalls.CaptureStream, mediaDesc); err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to set stream sources: %w", err)
	}

	inputGroupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to get group call: %w", err)
	}

	groupCall, ok := inputGroupCall.(tg.InputGroupCall)
	if !ok {
		c.binding.Stop(chatID)
		return fmt.Errorf("invalid group call type")
	}

	resultParams := `{"transport": null}`
	callRes, err := c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Muted:        false,
		VideoStopped: mediaDesc.Video == nil,
		Call:         groupCall,
		Params:       &tg.DataJson{Data: jsonParams},
	})
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to join group call: %w", err)
	}

	if updates, ok := callRes.(*tg.UpdatesObj); ok {
		for _, u := range updates.Updates {
			if connUpdate, ok := u.(*tg.UpdateGroupCallConnection); ok {
				resultParams = connUpdate.Params.Data
				break
			}
		}
	}

	if err := c.binding.Connect(chatID, resultParams, false); err != nil {
		return fmt.Errorf("ntgcalls connect failed: %w", err)
	}

	select {
	case err := <-waitChan:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("connection timeout after 20s")
	}
}

// GetInputGroupCall gets input group call for chat
// Gogram returns raw peer IDs which can be:
//   - Negative (standard Telegram format): -100XXXXXXXXX for supergroups
//   - Positive large number: raw channel/group ID without negation
func (c *Calls) GetInputGroupCall(chatID int64) (interface{}, error) {
	// Normalize: gogram sometimes gives positive raw IDs
	// Try resolving peer directly - works for both positive and negative IDs
	peer, err := c.client.ResolvePeer(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer: %w", err)
	}

	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		// Supergroup / Channel
		fullChannel, err := c.client.ChannelsGetFullChannel(&tg.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get full channel: %w", err)
		}

		if fullChan, ok := fullChannel.FullChat.(*tg.ChannelFull); ok {
			if fullChan.Call == nil {
				return nil, fmt.Errorf("❌ No active Voice Chat found!\nPlease start a Voice Chat from group settings first.")
			}
			return fullChan.Call, nil
		}
		return nil, fmt.Errorf("could not parse channel info")

	case *tg.InputPeerChat:
		// Regular group
		fullChat, err := c.client.MessagesGetFullChat(p.ChatID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full chat: %w", err)
		}

		if chatFull, ok := fullChat.FullChat.(*tg.ChatFullObj); ok {
			if chatFull.Call == nil {
				return nil, fmt.Errorf("❌ No active Voice Chat found!\nPlease start a Voice Chat from group settings first.")
			}
			return chatFull.Call, nil
		}
		return nil, fmt.Errorf("could not parse chat info")

	default:
		return nil, fmt.Errorf("unsupported peer type: %T (chatID: %d)", peer, chatID)
	}
}

func (c *Calls) Stop() {
	c.binding.Stop(-1)
}
