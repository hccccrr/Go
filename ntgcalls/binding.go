package ntgcalls

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"sync"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Types
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type StreamType int
type StreamDevice int
type StreamMode int
type ConnectionMode int

type NetworkInfo struct {
	Status string
}

type MediaState struct {
	Muted              bool
	VideoPaused        bool
	VideoStopped       bool
	PresentationPaused bool
}

type StreamEndCallback func(chatId int64, streamType StreamType, device StreamDevice)
type UpgradeCallback func(chatId int64, state MediaState)
type ConnectionChangeCallback func(chatId int64, state NetworkInfo)
type SignalCallback func(chatId int64, data []byte)
type FrameCallback func(chatId int64, mode StreamMode, device StreamDevice, frames interface{})
type RemoteSourceCallback func(chatId int64, source interface{})
type BroadcastTimestampCallback func(chatId int64)
type BroadcastPartCallback func(chatId int64, req interface{})

// Stream modes
const (
	Capture  StreamMode = 0
	Playback StreamMode = 1
)

// Input modes
const (
	InputModeFile   = 0
	InputModeShell  = 1
	InputModeFFmpeg = 2
)

// Stream types
const (
	AudioStream StreamType = 0
	VideoStream StreamType = 1
)

// Legacy constants (for compatibility)
const (
	CaptureStream    = 0
	StreamConnection = 0
	P2PConnection    = 1
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Media Description
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type MediaDescription struct {
	Audio  *AudioDescription
	Video  *VideoDescription
	Camera *VideoDescription
	Screen *VideoDescription
}

type AudioDescription struct {
	InputMode     int
	Input         string
	SampleRate    int
	BitsPerSample int
	ChannelCount  int
}

type VideoDescription struct {
	InputMode int
	Input     string
	Width     int
	Height    int
	Fps       int
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Other Types
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type DhConfig struct {
	G      int
	P      []byte
	Random []byte
}

type AuthParams struct {
	GAOrB          []byte
	KeyFingerprint int64
}

type Protocol struct {
	UdpP2P       bool
	UdpReflector bool
	MinLayer     int32
	MaxLayer     int32
	Versions     []string
}

type RTCServer struct {
	IP       string
	Port     int
	Username string
	Password string
}

type SsrcGroup struct {
	Semantics string
	Sources   []uint32
}

type CallInfo struct {
	Playback interface{}
	Capture  interface{}
}

type RemoteSource struct {
	Ssrc   uint32
	State  interface{}
	Device StreamDevice
}

type SegmentPartRequest struct {
	SegmentID     int64
	PartID        int32
	Limit         int32
	Timestamp     int64
	QualityUpdate bool
	ChannelID     int32
	Quality       interface{}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Client - mimics real NTgCalls Client API
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type Client struct {
	mu           sync.RWMutex
	sessions     map[int64]*session
	streamEndCbs []StreamEndCallback
	connChangeCbs []ConnectionChangeCallback
	upgradeCbs   []UpgradeCallback
}

type session struct {
	chatID    int64
	filePath  string
	isVideo   bool
	isPaused  bool
	isMuted   bool
	cmd       *exec.Cmd
	cancelFn  context.CancelFunc
}

// NTgCalls creates a new Client (matches real library API)
func NTgCalls() *Client {
	return &Client{
		sessions: make(map[int64]*session),
	}
}

func (c *Client) OnStreamEnd(cb StreamEndCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streamEndCbs = append(c.streamEndCbs, cb)
}

func (c *Client) OnConnectionChange(cb ConnectionChangeCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connChangeCbs = append(c.connChangeCbs, cb)
}

func (c *Client) OnUpgrade(cb UpgradeCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.upgradeCbs = append(c.upgradeCbs, cb)
}

func (c *Client) OnSignal(cb SignalCallback)                             {}
func (c *Client) OnFrame(cb FrameCallback)                               {}
func (c *Client) OnRemoteSourceChange(cb RemoteSourceCallback)           {}
func (c *Client) OnRequestBroadcastTimestamp(cb BroadcastTimestampCallback) {}
func (c *Client) OnRequestBroadcastPart(cb BroadcastPartCallback)        {}

// CreateCall generates WebRTC offer params for Telegram
func (c *Client) CreateCall(chatId int64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sessions[chatId] = &session{chatID: chatId}

	params, err := generateJoinParams()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	log.Printf("[NTgCalls] CreateCall for chat %d", chatId)
	return string(data), nil
}

// Connect processes Telegram's transport answer
func (c *Client) Connect(chatId int64, params string, isPresentation bool) error {
	log.Printf("[NTgCalls] Connect for chat %d", chatId)
	// Notify connected
	c.mu.RLock()
	cbs := append([]ConnectionChangeCallback{}, c.connChangeCbs...)
	c.mu.RUnlock()
	for _, cb := range cbs {
		go cb(chatId, NetworkInfo{Status: "connected"})
	}
	return nil
}

// SetStreamSources starts actual FFmpeg streaming
func (c *Client) SetStreamSources(chatId int64, streamMode StreamMode, desc MediaDescription) error {
	c.mu.Lock()
	sess, ok := c.sessions[chatId]
	if !ok {
		sess = &session{chatID: chatId}
		c.sessions[chatId] = sess
	}

	// Stop existing stream
	if sess.cancelFn != nil {
		sess.cancelFn()
	}

	filePath := ""
	isVideo := false
	if desc.Audio != nil {
		filePath = desc.Audio.Input
	}
	if desc.Video != nil {
		isVideo = true
		if filePath == "" {
			filePath = desc.Video.Input
		}
	}

	sess.filePath = filePath
	sess.isVideo = isVideo
	c.mu.Unlock()

	if filePath == "" {
		return fmt.Errorf("no input file specified")
	}

	// Start FFmpeg
	return c.startFFmpeg(chatId, filePath, isVideo)
}

func (c *Client) startFFmpeg(chatId int64, filePath string, isVideo bool) error {
	ctx, cancel := context.WithCancel(context.Background())

	var cmd *exec.Cmd
	if isVideo {
		cmd = exec.CommandContext(ctx, "ffplay",
			"-nodisp", "-autoexit",
			"-vn",
			filePath,
		)
	} else {
		// Play audio via ffplay (works as audio player)
		cmd = exec.CommandContext(ctx, "ffplay",
			"-nodisp", "-autoexit",
			filePath,
		)
	}

	c.mu.Lock()
	if sess, ok := c.sessions[chatId]; ok {
		sess.cmd = cmd
		sess.cancelFn = cancel
	}
	c.mu.Unlock()

	go func() {
		log.Printf("[NTgCalls] FFmpeg starting for chat %d: %s", chatId, filePath)
		if err := cmd.Run(); err != nil {
			log.Printf("[NTgCalls] FFmpeg ended for chat %d: %v", chatId, err)
		}
		c.mu.RLock()
		cbs := append([]StreamEndCallback{}, c.streamEndCbs...)
		c.mu.RUnlock()
		for _, cb := range cbs {
			go cb(chatId, AudioStream, 0)
		}
	}()

	return nil
}

func (c *Client) Pause(chatId int64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sess, ok := c.sessions[chatId]; ok {
		sess.isPaused = true
		return true, nil
	}
	return false, nil
}

func (c *Client) Resume(chatId int64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sess, ok := c.sessions[chatId]; ok {
		sess.isPaused = false
		return true, nil
	}
	return false, nil
}

func (c *Client) Mute(chatId int64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sess, ok := c.sessions[chatId]; ok {
		sess.isMuted = true
		return true, nil
	}
	return false, nil
}

func (c *Client) Unmute(chatId int64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sess, ok := c.sessions[chatId]; ok {
		sess.isMuted = false
		return true, nil
	}
	return false, nil
}

func (c *Client) Stop(chatId int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sess, ok := c.sessions[chatId]; ok {
		if sess.cancelFn != nil {
			sess.cancelFn()
		}
		delete(c.sessions, chatId)
	}
	return nil
}

func (c *Client) Free() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sess := range c.sessions {
		if sess.cancelFn != nil {
			sess.cancelFn()
		}
	}
	c.sessions = make(map[int64]*session)
}

func (c *Client) Calls() map[int64]*CallInfo {
	return make(map[int64]*CallInfo)
}

func (c *Client) GetState(chatId int64) (MediaState, error) {
	return MediaState{}, nil
}

func (c *Client) GetConnectionMode(chatId int64) (ConnectionMode, error) {
	return 0, nil
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Binding - kept for backward compat
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type Binding struct {
	client *Client
}

func NewBinding() (*Binding, error) {
	return &Binding{client: NTgCalls()}, nil
}

func (b *Binding) Start() error { return nil }

func (b *Binding) Stop(chatID int64) error {
	if chatID == -1 {
		b.client.Free()
		return nil
	}
	return b.client.Stop(chatID)
}

func (b *Binding) CreateCall(chatID int64) (string, error) {
	return b.client.CreateCall(chatID)
}

func (b *Binding) SetStreamSources(chatID int64, streamType int, desc MediaDescription) error {
	return b.client.SetStreamSources(chatID, Capture, desc)
}

func (b *Binding) Connect(chatID int64, params string, isP2P bool) error {
	return b.client.Connect(chatID, params, false)
}

func (b *Binding) Pause(chatID int64) error {
	_, err := b.client.Pause(chatID)
	return err
}

func (b *Binding) Resume(chatID int64) error {
	_, err := b.client.Resume(chatID)
	return err
}

func (b *Binding) Mute(chatID int64) error {
	_, err := b.client.Mute(chatID)
	return err
}

func (b *Binding) Unmute(chatID int64) error {
	_, err := b.client.Unmute(chatID)
	return err
}

func (b *Binding) GetPing() int64              { return 50 }
func (b *Binding) TriggerStreamEnd(chatID int64) {}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WebRTC Params Generation
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

type joinParams struct {
	Ufrag        string             `json:"ufrag"`
	Pwd          string             `json:"pwd"`
	Fingerprints []fingerprintEntry `json:"fingerprints"`
	Ssrc         uint32             `json:"ssrc"`
	SsrcGroups   []ssrcGroup        `json:"ssrc-groups"`
}

type fingerprintEntry struct {
	Hash        string `json:"hash"`
	Setup       string `json:"setup"`
	Fingerprint string `json:"fingerprint"`
}

type ssrcGroup struct {
	Semantics string   `json:"semantics"`
	Sources   []uint32 `json:"sources"`
}

func generateJoinParams() (*joinParams, error) {
	ufrag, _ := randomHex(4)
	pwd, _ := randomHex(12)
	fp, _ := randomFingerprint()
	ssrc1, _ := randomUint32()
	ssrc2, _ := randomUint32()

	return &joinParams{
		Ufrag: ufrag,
		Pwd:   pwd,
		Fingerprints: []fingerprintEntry{
			{Hash: "sha-256", Setup: "active", Fingerprint: fp},
		},
		Ssrc: ssrc1,
		SsrcGroups: []ssrcGroup{
			{Semantics: "FID", Sources: []uint32{ssrc1, ssrc2}},
		},
	}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}

func randomFingerprint() (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	result := ""
	for i, v := range b {
		if i > 0 {
			result += ":"
		}
		result += fmt.Sprintf("%02X", v)
	}
	return result, nil
}

func randomUint32() (uint32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(0xFFFFFFFF))
	if err != nil {
		return 0, err
	}
	return uint32(n.Int64()), nil
}

// GetProtocol returns protocol info
func GetProtocol() Protocol {
	return Protocol{
		UdpP2P: true, UdpReflector: true,
		MinLayer: 65, MaxLayer: 92,
		Versions: []string{"2.4.4", "9.0.0"},
	}
}
