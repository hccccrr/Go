package core

import (
	"log"
	"strconv"
	"strings"

	"shizumusic/config"
)

// UsersData manages user permissions and setup
type UsersData struct {
	// Developer IDs
	Devs []int64
}

// NewUsersData creates new users data instance
func NewUsersData() *UsersData {
	return &UsersData{
		Devs: []int64{
			8244881089, // Vivan
			7616808278, // Bad
		},
	}
}

// SetupGodUsers initializes owner/god users
func (u *UsersData) SetupGodUsers() {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("👑 Setting up owners...")

	if config.Cfg.OwnerID == 0 {
		log.Println("⚠️  No owner ID configured")
		return
	}

	// Parse owner ID(s) from config
	ownerIDStr := strconv.FormatInt(config.Cfg.OwnerID, 10)
	godUsers := strings.Fields(ownerIDStr)

	for _, userStr := range godUsers {
		userID, err := strconv.ParseInt(userStr, 10, 64)
		if err != nil {
			continue
		}

		config.Cfg.GodMutex.Lock()
		config.Cfg.GodUsers[userID] = true
		config.Cfg.GodMutex.Unlock()

		log.Printf("✅ Added owner: %d", userID)
	}

	log.Println("✅ Owners setup complete!")
}

// SetupSudoUsers initializes sudo users
func (u *UsersData) SetupSudoUsers(db *Database) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("⭐ Setting up sudo users...")

	// Add developers
	for _, devID := range u.Devs {
		config.Cfg.SudoMutex.Lock()
		config.Cfg.SudoUsers[devID] = true
		config.Cfg.SudoMutex.Unlock()
	}

	// Get sudo users from database
	dbUsers, err := db.GetSudoUsers()
	if err != nil {
		log.Printf("Warning: Failed to get sudo users from DB: %v", err)
		dbUsers = []int64{}
	}

	// Add developers to database if not present
	for _, devID := range u.Devs {
		found := false
		for _, dbUser := range dbUsers {
			if dbUser == devID {
				found = true
				break
			}
		}

		if !found {
			if err := db.AddSudo(devID); err != nil {
				log.Printf("Warning: Failed to add developer %d: %v", devID, err)
			} else {
				log.Printf("✅ Added developer: %d", devID)
			}
		}
	}

	// Add owner as sudo
	if config.Cfg.OwnerID != 0 {
		config.Cfg.SudoMutex.Lock()
		config.Cfg.SudoUsers[config.Cfg.OwnerID] = true
		config.Cfg.SudoMutex.Unlock()

		// Check if owner in database
		found := false
		for _, dbUser := range dbUsers {
			if dbUser == config.Cfg.OwnerID {
				found = true
				break
			}
		}

		if !found {
			if err := db.AddSudo(config.Cfg.OwnerID); err != nil {
				log.Printf("Warning: Failed to add owner as sudo: %v", err)
			} else {
				log.Printf("✅ Added owner as sudo: %d", config.Cfg.OwnerID)
			}
		}
	}

	// Add all database sudo users to config
	for _, userID := range dbUsers {
		config.Cfg.SudoMutex.Lock()
		config.Cfg.SudoUsers[userID] = true
		config.Cfg.SudoMutex.Unlock()
	}

	config.Cfg.SudoMutex.RLock()
	totalSudo := len(config.Cfg.SudoUsers)
	config.Cfg.SudoMutex.RUnlock()

	log.Printf("✅ Total sudo users: %d", totalSudo)
}

// SetupBannedUsers initializes banned users
func (u *UsersData) SetupBannedUsers(db *Database) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚫 Setting up banned users...")

	// Get blocked users from database
	blockedUsers, err := db.GetBlockedUsers()
	if err != nil {
		log.Printf("Warning: Failed to get blocked users: %v", err)
		blockedUsers = []int64{}
	}

	for _, userID := range blockedUsers {
		config.Cfg.BannedMutex.Lock()
		config.Cfg.BannedUsers[userID] = true
		config.Cfg.BannedMutex.Unlock()
	}

	// Get globally banned users
	gbannedUsers, err := db.GetGbannedUsers()
	if err != nil {
		log.Printf("Warning: Failed to get gbanned users: %v", err)
		gbannedUsers = []int64{}
	}

	for _, userID := range gbannedUsers {
		config.Cfg.BannedMutex.Lock()
		config.Cfg.BannedUsers[userID] = true
		config.Cfg.BannedMutex.Unlock()
	}

	config.Cfg.BannedMutex.RLock()
	totalBanned := len(config.Cfg.BannedUsers)
	config.Cfg.BannedMutex.RUnlock()

	log.Printf("✅ Total banned users: %d", totalBanned)
}

// Setup initializes all user data
func (u *UsersData) Setup(db *Database) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("👥 Initializing user data...")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	u.SetupGodUsers()
	u.SetupSudoUsers(db)
	u.SetupBannedUsers(db)

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ User data initialized successfully!")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// Global user data instance
var UserData = NewUsersData()
