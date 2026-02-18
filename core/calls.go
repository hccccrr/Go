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

	// P2P call configs
	p2pConfigs      map[int64]*P2PConfig
	p2pConfigsMutex sync.RWMutex

	// Input calls
	inputCalls      map[int64]interface{}
	inputCallsMutex sync.RWMutex

	// Wait channels for connections
	waitConnect      map[int64]chan error
	waitConnectMutex sync.Mutex

	// Pending connections
	pendingConnections      map[int64]*PendingConnection
	pendingConnectionsMutex sync.Mutex
}

// P2PConfig holds P2P call configuration
type P2PConfig struct {
	DhConfig       ntgcalls.DhConfig
	GAorB          []byte
	KeyFingerprint int64
	IsOutgoing     bool
	PhoneCall      *tg.PhoneCallObj
	WaitData       chan error
}

// PendingConnection holds pending connection data
type PendingConnection struct {
	MediaDescription ntgcalls.MediaDescription
	Payload          string
}

// NewCalls creates a new calls instance
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

// Start initializes NTgCalls
func (c *Calls) Start() error {
	log.Println(">> Booting NTgCalls client...")

	if err := c.binding.Start(); err != nil {
		return fmt.Errorf("failed to start NTgCalls: %w", err)
	}

	log.Println(">> NTgCalls client booted!")
	return nil
}

// JoinVC joins a voice chat
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

// LeaveVC leaves a voice chat
func (c *Calls) LeaveVC(chatID int64) error {
	c.audienceMu.Lock()
	delete(c.audience, chatID)
	c.audienceMu.Unlock()

	return c.binding.Stop(chatID)
}

// PauseVC pauses voice chat
func (c *Calls) PauseVC(chatID int64) error {
	return c.binding.Pause(chatID)
}

// ResumeVC resumes voice chat
func (c *Calls) ResumeVC(chatID int64) error {
	return c.binding.Resume(chatID)
}

// MuteVC mutes voice chat
func (c *Calls) MuteVC(chatID int64) error {
	return c.binding.Mute(chatID)
}

// UnmuteVC unmutes voice chat
func (c *Calls) UnmuteVC(chatID int64) error {
	return c.binding.Unmute(chatID)
}

// GetPing returns NTgCalls ping
func (c *Calls) GetPing() int64 {
	return c.binding.GetPing()
}

// connectCall connects to a voice chat
// Music bot sirf group/supergroup chats me kaam karta hai - P2P nahi
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

	// Always route to group call handler
	// Music bots only work in groups/supergroups (negative chatIDs)
	return c.handleGroupCall(chatID, mediaDesc, jsonParams, waitChan)
}

// handleGroupCall handles group voice chat connection
func (c *Calls) handleGroupCall(chatID int64, mediaDesc ntgcalls.MediaDescription, jsonParams string, waitChan chan error) error {
	var err error

	// Create call
	jsonParams, err = c.binding.CreateCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to create call: %w", err)
	}

	// Set stream sources
	if err := c.binding.SetStreamSources(chatID, ntgcalls.CaptureStream, mediaDesc); err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to set stream sources: %w", err)
	}

	// Get input group call
	inputGroupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to get group call: %w", err)
	}

	// Type assert to InputGroupCall
	groupCall, ok := inputGroupCall.(tg.InputGroupCall)
	if !ok {
		c.binding.Stop(chatID)
		return fmt.Errorf("invalid group call type")
	}

	// Join group call via Telegram
	resultParams := `{"transport": null}`
	callRes, err := c.client.PhoneJoinGroupCall(&tg.PhoneJoinGroupCallParams{
		Muted:        false,
		VideoStopped: mediaDesc.Video == nil,
		Call:         groupCall,
		Params: &tg.DataJson{
			Data: jsonParams,
		},
	})
	if err != nil {
		c.binding.Stop(chatID)
		return fmt.Errorf("failed to join group call: %w", err)
	}

	// Extract connection params from response
	if updates, ok := callRes.(*tg.UpdatesObj); ok {
		for _, u := range updates.Updates {
			if connUpdate, ok := u.(*tg.UpdateGroupCallConnection); ok {
				resultParams = connUpdate.Params.Data
				break
			}
		}
	}

	// Connect NTgCalls with params
	if err := c.binding.Connect(chatID, resultParams, false); err != nil {
		return fmt.Errorf("ntgcalls connect failed: %w", err)
	}

	// Wait for connection or timeout
	select {
	case err := <-waitChan:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("connection timeout after 20s")
	}
}

// GetInputGroupCall gets input group call for chat
func (c *Calls) GetInputGroupCall(chatID int64) (interface{}, error) {
	// Supergroup/Channel: chatID < -1000000000000
	if chatID < -1000000000000 {
		channelID := -(chatID + 1000000000000)

		peer, err := c.client.ResolvePeer(chatID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve peer: %w", err)
		}

		var accessHash int64
		if inputPeer, ok := peer.(*tg.InputPeerChannel); ok {
			accessHash = inputPeer.AccessHash
		} else {
			return nil, fmt.Errorf("peer is not a channel")
		}

		fullChannel, err := c.client.ChannelsGetFullChannel(&tg.InputChannelObj{
			ChannelID:  channelID,
			AccessHash: accessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get full channel: %w", err)
		}

		if fullChan, ok := fullChannel.FullChat.(*tg.ChannelFull); ok {
			if fullChan.Call == nil {
				return nil, fmt.Errorf("no active voice chat in this group.\nStart a Voice Chat first from group settings.")
			}
			return fullChan.Call, nil
		}

		return nil, fmt.Errorf("could not parse channel info")
	}

	// Regular group: chatID between -1000000000000 and 0
	if chatID < 0 {
		groupID := -chatID

		fullChat, err := c.client.MessagesGetFullChat(groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full chat: %w", err)
		}

		if fullChat.FullChat == nil {
			return nil, fmt.Errorf("no full chat data")
		}

		if chatFull, ok := fullChat.FullChat.(*tg.ChatFullObj); ok {
			if chatFull.Call == nil {
				return nil, fmt.Errorf("no active voice chat in this group.\nStart a Voice Chat first from group settings.")
			}
			return chatFull.Call, nil
		}

		return nil, fmt.Errorf("could not parse chat info")
	}

	return nil, fmt.Errorf("invalid chat ID: %d (music bot only works in groups)", chatID)
}

// Stop stops all active calls
func (c *Calls) Stop() {
	c.binding.Stop(-1)
}
