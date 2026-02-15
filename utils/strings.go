package utils

// TEXTS contains all text templates for HellMusic V3
type TEXTS struct{}

var TextTemplates = TEXTS{}

// Song & Video Information
func (t TEXTS) AboutSong() string {
	return `╭─────────────────────╮
│  **🎵 Song Information**
╰─────────────────────╯

**📝 Title:** ` + "`%s`" + `
**📺 Channel:** ` + "`%s`" + `
**📅 Published:** ` + "`%s`" + `
**👁️ Views:** ` + "`%s`" + `
**⏱️ Duration:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

func (t TEXTS) AboutUser() string {
	return `╭─────────────────────╮
│  **👤 Top User Info**
╰─────────────────────╯

**👤 Name:** %s
**🆔 User ID:** ` + "`%d`" + `
**⭐ Level:** ` + "`%s`" + `
**🎵 Songs Played:** ` + "`%d`" + `
**📅 Member Since:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

func (t TEXTS) SongCaption() string {
	return `╭─────────────────────╮
│  **🎵 Download Info**
╰─────────────────────╯

**📝 Title:** [%s](%s)
**👁️ Views:** ` + "`%s`" + `
**⏱️ Duration:** ` + "`%s`" + `
**👤 Requested By:** %s

**🔗 Powered By:** %s`
}

// Playback Status
func (t TEXTS) Playing() string {
	return `╭─────────────────────╮
│  **🎵 Now Playing**
╰─────────────────────╯

**🔗 Stream:** %s

**📝 Song:** ` + "`%s`" + `
**⏱️ Duration:** ` + "`%s`" + `
**👤 Requested By:** %s`
}

func (t TEXTS) Queue() string {
	return `╭─────────────────────╮
│  **📋 Added to Queue**
╰─────────────────────╯

**🔢 Position:** ` + "`#%d`" + `
**📝 Song:** ` + "`%s`" + `
**⏱️ Duration:** ` + "`%s`" + `
**👤 Queued By:** %s`
}

// User Profile
func (t TEXTS) Profile() string {
	return `╭─────────────────────╮
│  %s
│  **👤 User Profile**
╰─────────────────────╯

**👤 Name:** %s
**🆔 User ID:** ` + "`%d`" + `
**📱 Type:** ` + "`%s`" + `
**⭐ Level:** ` + "`%s`" + `
**🎵 Songs Played:** ` + "`%d`" + `
**📅 Member Since:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

// Statistics
func (t TEXTS) Stats() string {
	return `╭─────────────────────╮
│  **📊 Bot Statistics**
╰─────────────────────╯

**📊 Server Stats:**
├ **👥 Total Users:** ` + "`%d`" + `
├ **💬 Total Chats:** ` + "`%d`" + `
├ **🚫 Gbans:** ` + "`%d`" + `
├ **🔒 Blocked:** ` + "`%d`" + `
├ **🎵 Songs Played:** ` + "`%d`" + `
└ **🎙️ Active VC:** ` + "`%d`" + `

**💻 System Stats:**
├ **🖥️ CPU Cores:** ` + "`%d`" + `
├ **⚡ CPU Usage:** ` + "`%s`" + `
├ **💾 Disk Usage:** ` + "`%s`" + `
├ **🎯 RAM Usage:** ` + "`%s`" + `
└ **⏰ Uptime:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

func (t TEXTS) System() string {
	return `╭─────────────────────╮
│  **💻 System Info**
╰─────────────────────╯

**🖥️ CPU Cores:** ` + "`%d`" + `
**⚡ CPU Usage:** ` + "`%s`" + `
**💾 Disk Usage:** ` + "`%s`" + `
**🎯 RAM Usage:** ` + "`%s`" + `
**⏰ Uptime:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

func (t TEXTS) PingReply() string {
	return `╭─────────────────────╮
│  **🏓 Pong!**
╰─────────────────────╯

**⚡ Speed:** ` + "`%s ms`" + `
**⏰ Uptime:** ` + "`%s`" + `
**🎙️ VC Ping:** ` + "`%s ms`" + ``
}

// Startup & Source
func (t TEXTS) Booted() string {
	return `╭─────────────────────╮
│  **#START**
│  **🎵 %s is Alive!**
╰─────────────────────╯

**📦 Version Info:**
├ **🎵 HellMusic:** ` + "`%s`" + `
├ **🐍 Python:** ` + "`%s`" + `
├ **📡 Telethon:** ` + "`%s`" + `
└ **📞 PyTgCalls:** ` + "`%s`" + `

**🔗 Powered By:** %s`
}

func (t TEXTS) Source() string {
	return `╭─────────────────────╮
│  **📦 Source Code**
╰─────────────────────╯

**📌 Note:**
• The source code is available on GitHub
• All projects under The-HellBot are open-source
• Free to use and modify to your needs
• Anyone selling this code is a scammer

**⭐ Support Us:**
• Star the repository if you like it
• Contact us for help with the code

**🔗 Powered By:** %s`
}

// Help Texts
func (t TEXTS) HelpAdmin() string {
	return `╭─────────────────────╮
│  **👑 Admin Commands**
╰─────────────────────╯

**🔐 Authorization:**
• ` + "`/auth`" + ` - Authorize user
• ` + "`/unauth`" + ` - Unauthorize user
• ` + "`/authlist`" + ` - List authorized users
• ` + "`/authchat`" + ` - Enable for all users

**🎵 Playback Control:**
• ` + "`/mute`" + ` - Mute the stream
• ` + "`/unmute`" + ` - Unmute the stream
• ` + "`/pause`" + ` - Pause playback
• ` + "`/resume`" + ` - Resume playback
• ` + "`/stop`" + ` ` + "`/end`" + ` - Stop playback
• ` + "`/skip`" + ` - Skip current track
• ` + "`/replay`" + ` - Replay from start

**⚙️ Advanced:**
• ` + "`/loop [0-10]`" + ` - Loop track (0 to disable)
• ` + "`/seek [seconds]`" + ` - Seek position
• ` + "`/clean`" + ` - Clear queue when bugged`
}

func (t TEXTS) HelpUser() string {
	return `╭─────────────────────╮
│  **👥 User Commands**
╰─────────────────────╯

**🎵 Play Music:**
• ` + "`/play`" + ` - Play audio track
• ` + "`/vplay`" + ` - Play video track
• ` + "`/fplay`" + ` - Force play audio
• ` + "`/fvplay`" + ` - Force play video

**❤️ Favorites:**
• ` + "`/favs`" + ` ` + "`/myfavs`" + ` - Show favorites
• ` + "`/delfavs`" + ` - Delete favorites

**ℹ️ Information:**
• ` + "`/current`" + ` ` + "`/playing`" + ` - Now playing
• ` + "`/queue`" + ` ` + "`/q`" + ` - View queue
• ` + "`/song`" + ` - Download song
• ` + "`/lyrics`" + ` - Get lyrics
• ` + "`/profile`" + ` ` + "`/me`" + ` - Your profile`
}

func (t TEXTS) HelpSudo() string {
	return `╭─────────────────────╮
│  **⭐ Sudo Commands**
╰─────────────────────╯

**📊 Management:**
• ` + "`/active`" + ` - Active voice chats
• ` + "`/autoend`" + ` - Auto-end toggle
• ` + "`/stats`" + ` - Full statistics
• ` + "`/logs`" + ` - Get bot logs

**🚫 Moderation:**
• ` + "`/block`" + ` ` + "`/unblock`" + ` - Block user
• ` + "`/blocklist`" + ` - Blocked users
• ` + "`/gban`" + ` ` + "`/ungban`" + ` - Global ban
• ` + "`/gbanlist`" + ` - Gbanned users

**⚙️ System:**
• ` + "`/restart`" + ` - Restart bot
• ` + "`/sudolist`" + ` - Sudo users`
}

func (t TEXTS) HelpOthers() string {
	return `╭─────────────────────╮
│  **📚 Other Commands**
╰─────────────────────╯

**ℹ️ General:**
• ` + "`/start`" + ` - Check if alive
• ` + "`/ping`" + ` - Check ping
• ` + "`/help`" + ` - Show help menu
• ` + "`/sysinfo`" + ` - System info
• ` + "`/leaderboard`" + ` - Top users`
}

func (t TEXTS) HelpOwners() string {
	return `╭─────────────────────╮
│  **🔱 Owner Commands**
╰─────────────────────╯

**💻 Execution:**
• ` + "`/eval`" + ` ` + "`/run`" + ` - Python script
• ` + "`/exec`" + ` ` + "`/sh`" + ` - Bash script

**⚙️ Config:**
• ` + "`/getvar`" + ` - Get config var

**👑 Sudo Management:**
• ` + "`/addsudo`" + ` - Add sudo user
• ` + "`/rmsudo`" + ` - Remove sudo user`
}

func (t TEXTS) HelpGC() string {
	return `**❓ Need Help?**

Get the complete help menu in your PM.
Click the button below to get started!`
}

func (t TEXTS) HelpPM() string {
	return `╭─────────────────────╮
│  **⚙️ Help Menu**
╰─────────────────────╯

**📌 Information:**
• Commands are categorized by user type
• Use buttons below to navigate
• Contact us if you need assistance

**🔗 Powered By:** %s`
}

// Start Messages
func (t TEXTS) StartGC() string {
	return `**🎵 HellMusic is Online!**

Ready to play some awesome music?
Use ` + "`/help`" + ` to see all commands!`
}

func (t TEXTS) StartPM() string {
	return `╭─────────────────────╮
│  **👋 Welcome!**
╰─────────────────────╯

**Hey** %s**!**

I'm **%s**, an advanced music bot that can play music in Voice Chats with high quality streaming!

**✨ Features:**
• High-quality audio streaming
• Video playback support
• Queue management
• Favorites system
• Advanced controls

Add me to your group and enjoy unlimited music!

**🔗 Powered By:** @%s`
}

// Miscellaneous
const Performer = "HellMusic V3"

func (t TEXTS) ErrorGeneric() string {
	return `**❌ An Error Occurred**

` + "```%s```" + `

Please try again later or contact support.`
}

func (t TEXTS) ErrorNoVC() string {
	return `**❌ No Active Voice Chat**

Please start a voice chat first!`
}

func (t TEXTS) ErrorNoPermission() string {
	return `**❌ Insufficient Permissions**

You don't have permission to use this command.`
}

func (t TEXTS) SuccessGeneric() string {
	return `**✅ Success**

%s`
}

func (t TEXTS) Loading() string {
	return "**⏳ Processing...**\n\nPlease wait..."
}

func (t TEXTS) Searching() string {
	return "**🔍 Searching...**\n\n`%s`"
}

func (t TEXTS) Downloading() string {
	return "**📥 Downloading...**\n\n`%s`"
}

func (t TEXTS) Processing() string {
	return "**⚙️ Processing...**\n\n`%s`"
}
