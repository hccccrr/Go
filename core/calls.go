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
	client    *tg.Client
	binding   *ntgcalls.Binding
	audience  map[int64]int
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
	// Create media description
	mediaDesc := ntgcalls.MediaDescription{
		Audio: &ntgcalls.AudioDescription{
			InputMode:   ntgcalls.InputModeFile,
			Input:       filePath,
			SampleRate:  48000,
			BitsPerSample: 16,
			ChannelCount: 2,
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

	// Connect to call
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

// connectCall connects to a call (group or P2P)
func (c *Calls) connectCall(chatID int64, mediaDesc ntgcalls.MediaDescription, jsonParams string) error {
	// Create wait channel
	c.waitConnectMutex.Lock()
	waitChan := make(chan error)
	c.waitConnect[chatID] = waitChan
	c.waitConnectMutex.Unlock()

	defer func() {
		c.waitConnectMutex.Lock()
		delete(c.waitConnect, chatID)
		c.waitConnectMutex.Unlock()
	}()

	if chatID >= 0 {
		// P2P call (private chat)
		return c.handleP2PCall(chatID, mediaDesc)
	}

	// Group call
	return c.handleGroupCall(chatID, mediaDesc, jsonParams, waitChan)
}

// handleGroupCall handles group voice chat connection
func (c *Calls) handleGroupCall(chatID int64, mediaDesc ntgcalls.MediaDescription, jsonParams string, waitChan chan error) error {
	var err error
	
	// Create call
	jsonParams, err = c.binding.CreateCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return err
	}

	// Set stream sources
	if err := c.binding.SetStreamSources(chatID, ntgcalls.CaptureStream, mediaDesc); err != nil {
		c.binding.Stop(chatID)
		return err
	}

	// Get input group call
	inputGroupCall, err := c.GetInputGroupCall(chatID)
	if err != nil {
		c.binding.Stop(chatID)
		return err
	}

	// Type assert to proper InputGroupCall type
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
		Call:         groupCall, // Use type-asserted value
		Params: &tg.DataJson{ // Changed from DataJSON to DataJson
			Data: jsonParams,
		},
	})
	if err != nil {
		c.binding.Stop(chatID)
		return err
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

	// Connect NTgCalls
	if err := c.binding.Connect(chatID, resultParams, false); err != nil {
		return err
	}

	// Wait for connection or timeout
	select {
	case err := <-waitChan:
		return err
	case <-time.After(20 * time.Second):
		return fmt.Errorf("connection timeout: no response from ntgcalls")
	}
}

// handleP2PCall handles P2P (private) voice chat
func (c *Calls) handleP2PCall(chatID int64, mediaDesc ntgcalls.MediaDescription) error {
	// P2P call implementation (simplified)
	// Full implementation would include DH key exchange, etc.
	return fmt.Errorf("P2P calls not yet implemented")
}

// GetInputGroupCall gets input group call for chat
func (c *Calls) GetInputGroupCall(chatID int64) (interface{}, error) {
	// Use ChannelsGetFullChannel instead of GetFullChat
	// First, get the channel/chat
	chat, err := c.client.GetChat(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	// Convert to input peer
	inputPeer := chat.InputPeer()
	
	// For channels/supergroups, use ChannelsGetFullChannel
	if inputChannel, ok := inputPeer.(*tg.InputPeerChannel); ok {
		fullChannel, err := c.client.ChannelsGetFullChannel(&tg.InputChannelObj{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get full channel: %w", err)
		}

		// Extract call from FullChannel
		if fullChan, ok := fullChannel.FullChat.(*tg.ChannelFullObj); ok {
			if fullChan.Call == nil {
				return nil, fmt.Errorf("no active group call in chat")
			}
			return fullChan.Call, nil
		}
	}

	// For regular groups, use MessagesGetFullChat
	if inputChat, ok := inputPeer.(*tg.InputPeerChat); ok {
		fullChat, err := c.client.MessagesGetFullChat(inputChat.ChatID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full chat: %w", err)
		}

		if fullChat.FullChat == nil {
			return nil, fmt.Errorf("no full chat data")
		}

		if chatFull, ok := fullChat.FullChat.(*tg.ChatFullObj); ok {
			if chatFull.Call == nil {
				return nil, fmt.Errorf("no active group call in chat")
			}
			return chatFull.Call, nil
		}
	}

	return nil, fmt.Errorf("unsupported chat type")
}

// Stop stops NTgCalls
func (c *Calls) Stop() {
	c.binding.Stop(-1) // Stop all
}
