package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database holds MongoDB connection and local caches
type Database struct {
	client *mongo.Client
	db     *mongo.Database

	// Collections
	authchats    *mongo.Collection
	authusers    *mongo.Collection
	autoend      *mongo.Collection
	blockedUsers *mongo.Collection
	chats        *mongo.Collection
	favorites    *mongo.Collection
	gbanDB       *mongo.Collection
	songsDB      *mongo.Collection
	sudoUsers    *mongo.Collection
	users        *mongo.Collection

	// Local caches (in-memory)
	activeVC      []ActiveVC
	activeVCMutex sync.RWMutex
	inactive      map[int64]time.Time
	inactiveMutex sync.RWMutex
	loop          map[int64]int
	loopMutex     sync.RWMutex
	watcher       map[int64]map[string]bool
	watcherMutex  sync.RWMutex
	audioEffects  map[int64]AudioEffects
	effectsMutex  sync.RWMutex
}

// ActiveVC represents an active voice chat
type ActiveVC struct {
	ChatID   int64     `bson:"chat_id"`
	JoinTime time.Time `bson:"join_time"`
	VCType   string    `bson:"vc_type"`
}

// AudioEffects stores audio processing settings
type AudioEffects struct {
	BassBoost int     `bson:"bass_boost"`
	Speed     float64 `bson:"speed"`
}

// User represents a user document
type User struct {
	UserID             int64     `bson:"user_id"`
	UserName           string    `bson:"user_name"`
	JoinDate           string    `bson:"join_date"`
	SongsPlayed        int       `bson:"songs_played"`
	MessagesCount      int       `bson:"messages_count"`
	LastMsgTime        []string  `bson:"last_msg_time"`
	SpamCooldownUntil  *string   `bson:"spam_cooldown_until,omitempty"`
}

// NewDatabase creates a new database connection
func NewDatabase(uri string) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println(">> Database connection successful!")

	db := client.Database("ShizuMusicDB")

	return &Database{
		client:       client,
		db:           db,
		authchats:    db.Collection("authchats"),
		authusers:    db.Collection("authusers"),
		autoend:      db.Collection("autoend"),
		blockedUsers: db.Collection("blocked_users"),
		chats:        db.Collection("chats"),
		favorites:    db.Collection("favorites"),
		gbanDB:       db.Collection("gban_db"),
		songsDB:      db.Collection("songsdb"),
		sudoUsers:    db.Collection("sudousers"),
		users:        db.Collection("users"),
		activeVC:     []ActiveVC{{ChatID: 0, JoinTime: time.Now(), VCType: "voice"}},
		inactive:     make(map[int64]time.Time),
		loop:         make(map[int64]int),
		watcher:      make(map[int64]map[string]bool),
		audioEffects: make(map[int64]AudioEffects),
	}, nil
}

// Connect pings the database to verify connection
func (d *Database) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	log.Println(">> Database connection successful!")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.client.Disconnect(ctx); err != nil {
		return err
	}

	log.Println(">> Database connection closed!")
	return nil
}

// ========== USER OPERATIONS ==========

// AddUser adds a new user
func (d *Database) AddUser(userID int64, userName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := User{
		UserID:        userID,
		UserName:      userName,
		JoinDate:      time.Now().Format("02-01-2006 15:04"),
		SongsPlayed:   0,
		MessagesCount: 0,
		LastMsgTime:   []string{},
	}

	_, err := d.users.InsertOne(ctx, user)
	return err
}

// IsUserExist checks if user exists
func (d *Database) IsUserExist(userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := d.users.CountDocuments(ctx, bson.M{"user_id": userID})
	return count > 0, err
}

// GetUser gets user data
func (d *Database) GetUser(userID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := d.users.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// UpdateUser updates user field
func (d *Database) UpdateUser(userID int64, key string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// For counter fields, increment instead of replace
	if key == "songs_played" || key == "messages_count" {
		user, _ := d.GetUser(userID)
		if user != nil {
			if key == "songs_played" {
				value = user.SongsPlayed + value.(int)
			} else {
				value = user.MessagesCount + value.(int)
			}
		}
	}

	_, err := d.users.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{key: value}},
	)
	return err
}

// TotalUsersCount returns total user count
func (d *Database) TotalUsersCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.users.CountDocuments(ctx, bson.M{})
}

// ========== ACTIVE VC OPERATIONS (Local) ==========

// AddActiveVC adds active voice chat
func (d *Database) AddActiveVC(chatID int64, vcType string) error {
	d.activeVCMutex.Lock()
	defer d.activeVCMutex.Unlock()

	// Check if already exists
	for _, vc := range d.activeVC {
		if vc.ChatID == chatID {
			return nil
		}
	}

	d.activeVC = append(d.activeVC, ActiveVC{
		ChatID:   chatID,
		JoinTime: time.Now(),
		VCType:   vcType,
	})
	return nil
}

// IsActiveVC checks if VC is active
func (d *Database) IsActiveVC(chatID int64) (bool, error) {
	d.activeVCMutex.RLock()
	defer d.activeVCMutex.RUnlock()

	for _, vc := range d.activeVC {
		if vc.ChatID == chatID {
			return true, nil
		}
	}
	return false, nil
}

// RemoveActiveVC removes active VC
func (d *Database) RemoveActiveVC(chatID int64) error {
	d.activeVCMutex.Lock()
	defer d.activeVCMutex.Unlock()

	for i, vc := range d.activeVC {
		if vc.ChatID == chatID {
			d.activeVC = append(d.activeVC[:i], d.activeVC[i+1:]...)
			break
		}
	}
	return nil
}

// GetActiveVC gets all active VCs
func (d *Database) GetActiveVC() []ActiveVC {
	d.activeVCMutex.RLock()
	defer d.activeVCMutex.RUnlock()

	return append([]ActiveVC{}, d.activeVC...)
}

// ========== LOOP OPERATIONS (Local) ==========

// SetLoop sets loop count
func (d *Database) SetLoop(chatID int64, count int) error {
	d.loopMutex.Lock()
	defer d.loopMutex.Unlock()

	d.loop[chatID] = count
	return nil
}

// GetLoop gets loop count
func (d *Database) GetLoop(chatID int64) (int, error) {
	d.loopMutex.RLock()
	defer d.loopMutex.RUnlock()

	return d.loop[chatID], nil
}

// ========== AUDIO EFFECTS (Local) ==========

// SetAudioEffects sets audio effects
func (d *Database) SetAudioEffects(chatID int64, bassBoost int, speed float64) error {
	d.effectsMutex.Lock()
	defer d.effectsMutex.Unlock()

	d.audioEffects[chatID] = AudioEffects{
		BassBoost: bassBoost,
		Speed:     speed,
	}
	return nil
}

// GetAudioEffects gets audio effects
func (d *Database) GetAudioEffects(chatID int64) AudioEffects {
	d.effectsMutex.RLock()
	defer d.effectsMutex.RUnlock()

	if effects, ok := d.audioEffects[chatID]; ok {
		return effects
	}
	return AudioEffects{BassBoost: 0, Speed: 1.0}
}

// ========== SUDO USERS ==========

// GetSudoUsers gets sudo users list
func (d *Database) GetSudoUsers() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		UserIDs []int64 `bson:"user_ids"`
	}

	err := d.sudoUsers.FindOne(ctx, bson.M{"sudo": "sudo"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []int64{}, nil
	}
	if err != nil {
		return nil, err
	}

	return result.UserIDs, nil
}

// AddSudo adds sudo user
func (d *Database) AddSudo(userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, _ := d.GetSudoUsers()
	users = append(users, userID)

	_, err := d.sudoUsers.UpdateOne(
		ctx,
		bson.M{"sudo": "sudo"},
		bson.M{"$set": bson.M{"user_ids": users}},
		options.Update().SetUpsert(true),
	)
	return err
}

// ========== BLOCKED/GBANNED USERS ==========

// GetBlockedUsers gets blocked users
func (d *Database) GetBlockedUsers() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		UserIDs []int64 `bson:"user_ids"`
	}

	err := d.blockedUsers.FindOne(ctx, bson.M{"blocked": "blocked"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []int64{}, nil
	}
	return result.UserIDs, err
}

// GetGbannedUsers gets globally banned users
func (d *Database) GetGbannedUsers() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		UserIDs []int64 `bson:"user_ids"`
	}

	err := d.gbanDB.FindOne(ctx, bson.M{"gbanned": "gbanned"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []int64{}, nil
	}
	return result.UserIDs, err
}

// ========== AUTHCHATS ==========

// IsAuthchat checks if chat allows all users
func (d *Database) IsAuthchat(chatID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		ChatIDs []int64 `bson:"chat_ids"`
	}

	err := d.authchats.FindOne(ctx, bson.M{"authchats": "authchats"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, id := range result.ChatIDs {
		if id == chatID {
			return true, nil
		}
	}
	return false, nil
}

// ========== AUTOEND ==========

// GetAutoend checks if autoend is enabled
func (d *Database) GetAutoend() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Status string `bson:"status"`
	}

	err := d.autoend.FindOne(ctx, bson.M{"autoend": "autoend"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return true, nil // Default enabled
	}
	if err != nil {
		return true, err
	}

	return result.Status == "on", nil
}

// UpdateSongsCount increments songs count
func (d *Database) UpdateSongsCount(count int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	current, _ := d.TotalSongsCount()
	newCount := current + count

	_, err := d.songsDB.UpdateOne(
		ctx,
		bson.M{"songs": "songs"},
		bson.M{"$set": bson.M{"count": newCount}},
		options.Update().SetUpsert(true),
	)
	return err
}

// TotalSongsCount gets total songs played
func (d *Database) TotalSongsCount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Count int `bson:"count"`
	}

	err := d.songsDB.FindOne(ctx, bson.M{"songs": "songs"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	return result.Count, err
}
