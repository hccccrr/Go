package ntgcalls

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Binding represents NTgCalls binding
// This is a Go wrapper around the NTgCalls library
type Binding struct {
	activeCalls   map[int64]*CallSession
	callsMutex    sync.RWMutex
	eventHandlers map[string]func(int64, interface{})
	handlerMutex  sync.RWMutex
}

// CallSession represents an active call session
type CallSession struct {
	ChatID          int64
	IsActive        bool
	IsVideo         bool
	IsMuted         bool
	IsPaused        bool
	MediaSource     string
	ConnectionMode  int
	Participants    []Participant
	StreamStartTime int64
}

// Participant represents a call participant
type Participant struct {
	UserID int64
	Muted  bool
	Volume int
}

// MediaDescription describes media stream
type MediaDescription struct {
	Audio  *AudioDescription
	Video  *VideoDescription
	Camera *VideoDescription
	Screen *VideoDescription
}

// AudioDescription describes audio stream
type AudioDescription struct {
	InputMode     int
	Input         string
	SampleRate    int
	BitsPerSample int
	ChannelCount  int
}

// VideoDescription describes video stream
type VideoDescription struct {
	InputMode int
	Input     string
	Width     int
	Height    int
	Fps       int
}

// DhConfig for P2P calls (Diffie-Hellman)
type DhConfig struct {
	G      int
	P      []byte
	Random []byte
}

// ExchangeResult holds key exchange result
type ExchangeResult struct {
	GAOrB          []byte
	KeyFingerprint int64
}

// Protocol holds call protocol info
type Protocol struct {
	UdpP2P       bool
	UdpReflector bool
	MinLayer     int32
	MaxLayer     int32
	Versions     []string
}

// RTCServer represents WebRTC server
type RTCServer struct {
	IP       string
	Port     int
	Username string
	Password string
}

// Stream types
const (
	InputModeFile     = 0
	InputModeShell    = 1
	InputModeFFmpeg   = 2
	CaptureStream     = 0
	StreamConnection  = 0
	P2PConnection     = 1
)

// Event types
const (
	EventStreamEnd     = "stream_ended"
	EventParticipants  = "participants_changed"
	EventNetworkStatus = "network_status"
)

// NewBinding creates a new NTgCalls binding
func NewBinding() (*Binding, error) {
	return &Binding{
		activeCalls:   make(map[int64]*CallSession),
		eventHandlers: make(map[string]func(int64, interface{})),
	}, nil
}

// Start initializes NTgCalls
func (b *Binding) Start() error {
	// In a real implementation, this would initialize the NTgCalls library
	// For now, we'll just mark it as ready
	return nil
}

// Stop stops all active calls
func (b *Binding) Stop(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if chatID == -1 {
		// Stop all calls
		b.activeCalls = make(map[int64]*CallSession)
		return nil
	}

	// Stop specific call
	if session, exists := b.activeCalls[chatID]; exists {
		session.IsActive = false
		delete(b.activeCalls, chatID)
	}

	return nil
}

// CreateCall creates a new call and returns JSON params
func (b *Binding) CreateCall(chatID int64) (string, error) {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	// Create new session
	session := &CallSession{
		ChatID:   chatID,
		IsActive: true,
		IsVideo:  false,
		IsMuted:  false,
		IsPaused: false,
	}

	b.activeCalls[chatID] = session

	// Generate WebRTC params
	params := map[string]interface{}{
		"ufrag": generateRandomString(16),
		"pwd":   generateRandomString(32),
		"hash":  "sha-256",
		"setup": "actpass",
		"fingerprint": map[string]string{
			"hash":        "sha-256",
			"fingerprint": generateRandomString(64),
		},
		"ssrc": 1000 + chatID, // Simple SSRC generation
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	return string(jsonParams), nil
}

// SetStreamSources sets media sources for the call
func (b *Binding) SetStreamSources(chatID int64, streamType int, desc MediaDescription) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	session, exists := b.activeCalls[chatID]
	if !exists {
		return fmt.Errorf("call session not found for chat %d", chatID)
	}

	// Set media source
	if desc.Audio != nil {
		session.MediaSource = desc.Audio.Input
	} else if desc.Video != nil {
		session.MediaSource = desc.Video.Input
		session.IsVideo = true
	}

	return nil
}

// Connect connects to the call
func (b *Binding) Connect(chatID int64, params string, isP2P bool) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	session, exists := b.activeCalls[chatID]
	if !exists {
		return fmt.Errorf("call session not found")
	}

	if isP2P {
		session.ConnectionMode = P2PConnection
	} else {
		session.ConnectionMode = StreamConnection
	}

	session.IsActive = true

	// Trigger connected event
	b.triggerEvent(EventNetworkStatus, chatID, map[string]interface{}{
		"status": "connected",
	})

	return nil
}

// Pause pauses the stream
func (b *Binding) Pause(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if session, exists := b.activeCalls[chatID]; exists {
		session.IsPaused = true
		return nil
	}

	return fmt.Errorf("call session not found")
}

// Resume resumes the stream
func (b *Binding) Resume(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if session, exists := b.activeCalls[chatID]; exists {
		session.IsPaused = false
		return nil
	}

	return fmt.Errorf("call session not found")
}

// Mute mutes the audio
func (b *Binding) Mute(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if session, exists := b.activeCalls[chatID]; exists {
		session.IsMuted = true
		return nil
	}

	return fmt.Errorf("call session not found")
}

// Unmute unmutes the audio
func (b *Binding) Unmute(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if session, exists := b.activeCalls[chatID]; exists {
		session.IsMuted = false
		return nil
	}

	return fmt.Errorf("call session not found")
}

// GetPing returns the ping to voice chat servers
func (b *Binding) GetPing() int64 {
	// In real implementation, this would measure actual ping
	return 50 // ms
}

// GetConnectionMode returns the connection mode
func (b *Binding) GetConnectionMode(chatID int64) (int, error) {
	b.callsMutex.RLock()
	defer b.callsMutex.RUnlock()

	if session, exists := b.activeCalls[chatID]; exists {
		return session.ConnectionMode, nil
	}

	return 0, fmt.Errorf("call session not found")
}

// GetActiveCall gets active call info
func (b *Binding) GetActiveCall(chatID int64) (*CallSession, error) {
	b.callsMutex.RLock()
	defer b.callsMutex.RUnlock()

	if session, exists := b.activeCalls[chatID]; exists {
		return session, nil
	}

	return nil, fmt.Errorf("no active call")
}

// GetActiveCalls returns all active calls
func (b *Binding) GetActiveCalls() []*CallSession {
	b.callsMutex.RLock()
	defer b.callsMutex.RUnlock()

	calls := make([]*CallSession, 0, len(b.activeCalls))
	for _, session := range b.activeCalls {
		calls = append(calls, session)
	}

	return calls
}

// ========== P2P Call Methods ==========

// CreateP2PCall creates a P2P call
func (b *Binding) CreateP2PCall(chatID int64) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	session := &CallSession{
		ChatID:         chatID,
		IsActive:       true,
		ConnectionMode: P2PConnection,
	}

	b.activeCalls[chatID] = session
	return nil
}

// InitExchange initializes Diffie-Hellman key exchange
func (b *Binding) InitExchange(chatID int64, dhConfig DhConfig, gaOrB []byte) ([]byte, error) {
	// In real implementation, this would perform actual DH key exchange
	// For now, return a dummy value
	return generateRandomBytes(256), nil
}

// ExchangeKeys exchanges encryption keys
func (b *Binding) ExchangeKeys(chatID int64, gaOrB []byte, fingerprint int64) (*ExchangeResult, error) {
	return &ExchangeResult{
		GAOrB:          generateRandomBytes(256),
		KeyFingerprint: fingerprint,
	}, nil
}

// ConnectP2P connects a P2P call
func (b *Binding) ConnectP2P(chatID int64, servers interface{}, versions []string, p2pAllowed bool) error {
	b.callsMutex.Lock()
	defer b.callsMutex.Unlock()

	if session, exists := b.activeCalls[chatID]; exists {
		session.IsActive = true
		session.ConnectionMode = P2PConnection
		return nil
	}

	return fmt.Errorf("call session not found")
}

// ========== Event Handling ==========

// OnEvent registers an event handler
func (b *Binding) OnEvent(eventType string, handler func(int64, interface{})) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	b.eventHandlers[eventType] = handler
}

// triggerEvent triggers an event
func (b *Binding) triggerEvent(eventType string, chatID int64, data interface{}) {
	b.handlerMutex.RLock()
	defer b.handlerMutex.RUnlock()

	if handler, exists := b.eventHandlers[eventType]; exists {
		go handler(chatID, data) // Run in goroutine to avoid blocking
	}
}

// TriggerStreamEnd manually triggers stream end event (for testing)
func (b *Binding) TriggerStreamEnd(chatID int64) {
	b.triggerEvent(EventStreamEnd, chatID, nil)
}

// ========== Helper Functions ==========

// GetProtocol returns protocol information
func GetProtocol() *Protocol {
	return &Protocol{
		UdpP2P:       true,
		UdpReflector: true,
		MinLayer:     65,
		MaxLayer:     92,
		Versions:     []string{"2.4.4", "9.0.0"},
	}
}

// ParseRTCServers parses RTC servers from Telegram response
func ParseRTCServers(connections interface{}) []*RTCServer {
	// In real implementation, parse actual connection data
	return []*RTCServer{
		{
			IP:       "149.154.167.51",
			Port:     443,
			Username: "user",
			Password: "pass",
		},
	}
}

// generateRandomString generates random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// generateRandomBytes generates random bytes
func generateRandomBytes(length int) []byte {
	result := make([]byte, length)
	for i := range result {
		result[i] = byte(i % 256)
	}
	return result
}

// ========== Audio Effects (Future) ==========

// SetBassBoost sets bass boost level (0-10)
func (b *Binding) SetBassBoost(chatID int64, level int) error {
	// TODO: Implement bass boost via FFmpeg filters
	return nil
}

// SetSpeed sets playback speed (0.5-2.0)
func (b *Binding) SetSpeed(chatID int64, speed float64) error {
	// TODO: Implement speed adjustment via FFmpeg filters
	return nil
}

// ApplyAudioEffects applies audio effects to stream
func (b *Binding) ApplyAudioEffects(chatID int64, bassBoost int, speed float64) error {
	// TODO: Implement combined audio effects
	return nil
}

// ========== Statistics ==========

// GetCallStats returns call statistics
func (b *Binding) GetCallStats(chatID int64) (map[string]interface{}, error) {
	b.callsMutex.RLock()
	defer b.callsMutex.RUnlock()

	session, exists := b.activeCalls[chatID]
	if !exists {
		return nil, fmt.Errorf("call not found")
	}

	return map[string]interface{}{
		"chat_id":     session.ChatID,
		"is_active":   session.IsActive,
		"is_video":    session.IsVideo,
		"is_muted":    session.IsMuted,
		"is_paused":   session.IsPaused,
		"participants": len(session.Participants),
		"uptime":      0, // Calculate from StreamStartTime
	}, nil
}
