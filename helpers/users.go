package helpers

import "fmt"

// UserModel handles user profile and statistics
type UserModel struct {
	profile string
	stats   string
}

// NewUserModel creates a new UserModel instance
func NewUserModel() *UserModel {
	return &UserModel{
		profile: TextTemplates.Profile(),
		stats:   TextTemplates.Stats(),
	}
}

// ProfileContext contains user profile data
type ProfileContext struct {
	Mention     string
	ID          int64
	UserType    string
	SongsPlayed int
	JoinDate    string
}

// StatsContext contains bot statistics data
type StatsContext struct {
	Users   int
	Chats   int
	Gbans   int
	Blocked int
	Songs   int
	Active  int
	Core    int
	CPU     string
	Disk    string
	RAM     string
	Uptime  string
	Mention string
}

// GetProfileText generates formatted profile text
func (um *UserModel) GetProfileText(ctx ProfileContext, mention string) string {
	return fmt.Sprintf(
		um.profile,
		um.GetUserLevelSymbol(ctx.SongsPlayed),
		ctx.Mention,
		ctx.ID,
		ctx.UserType,
		um.GetUserLevel(ctx.SongsPlayed),
		ctx.SongsPlayed,
		ctx.JoinDate,
		mention,
	)
}

// GetUserLevel returns the user level based on songs played
func (um *UserModel) GetUserLevel(songsPlayed int) string {
	switch {
	case songsPlayed < 50:
		return "Novice"
	case songsPlayed < 100:
		return "Beginner"
	case songsPlayed < 200:
		return "Intermediate"
	case songsPlayed < 400:
		return "Advanced"
	case songsPlayed < 800:
		return "Expert"
	default:
		return "Master"
	}
}

// GetUserLevelSymbol returns star symbols based on user level
func (um *UserModel) GetUserLevelSymbol(songsPlayed int) string {
	switch {
	case songsPlayed < 50:
		return "☆☆☆☆☆"
	case songsPlayed < 100:
		return "★☆☆☆☆"
	case songsPlayed < 200:
		return "★★☆☆☆"
	case songsPlayed < 400:
		return "★★★☆☆"
	case songsPlayed < 800:
		return "★★★★☆"
	default:
		return "★★★★★"
	}
}

// GetStatsText generates formatted statistics text
func (um *UserModel) GetStatsText(ctx StatsContext) string {
	return fmt.Sprintf(
		um.stats,
		ctx.Users,
		ctx.Chats,
		ctx.Gbans,
		ctx.Blocked,
		ctx.Songs,
		ctx.Active,
		ctx.Core,
		ctx.CPU,
		ctx.Disk,
		ctx.RAM,
		ctx.Uptime,
		ctx.Mention,
	)
}

// MusicUser is the global instance
var MusicUser = NewUserModel()
