package helpers

// TEXTS contains all text templates
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

func (t TEXTS) StartPM() string {
	return `╭─────────────────────╮
│  **👋 Welcome!**
╰─────────────────────╯

**Hey** %s**!**

I'm **%s**, an advanced music bot!

**✨ Features:**
• High-quality audio streaming
• Video playback support
• Queue management
• Favorites system

Add me to your group and enjoy music!

**🔗 Powered By:** @%s`
}

func (t TEXTS) StartGC() string {
	return `**🎵 Music Bot Online!**

Ready to play awesome music?
Use ` + "`/help`" + ` to see commands!`
}

func (t TEXTS) HelpPM() string {
	return `╭─────────────────────╮
│  **⚙️ Help Menu**
╰─────────────────────╯

**📌 Information:**
• Commands are categorized by user type
• Use buttons below to navigate

**🔗 Powered By:** %s`
}

func (t TEXTS) HelpGC() string {
	return `**❓ Need Help?**

Get the complete help menu in your PM.
Click the button below!`
}

func (t TEXTS) PingReply() string {
	return `╭─────────────────────╮
│  **🏓 Pong!**
╰─────────────────────╯

**⚡ Speed:** ` + "`%d ms`" + `
**⏰ Uptime:** ` + "`%s`" + `
**🎙️ VC Ping:** ` + "`%s ms`" + ``
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

func (t TEXTS) HelpAdmin() string {
	return "**👑 Admin Commands**\n\n/pause, /resume, /skip, /loop"
}

func (t TEXTS) HelpUser() string {
	return "**👥 User Commands**\n\n/play, /queue, /current"
}

func (t TEXTS) HelpSudo() string {
	return "**⭐ Sudo Commands**\n\n/stats, /gban, /restart"
}

func (t TEXTS) HelpOwners() string {
	return "**🔱 Owner Commands**\n\n/eval, /exec, /addsudo"
}
