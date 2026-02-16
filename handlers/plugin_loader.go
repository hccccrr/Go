package handlers

import (
	"log"
	
	"shizumusic/core"
)

// PluginRegistration holds plugin registration info
type PluginRegistration struct {
	Name     string
	Register func(client *core.Client, db *core.Database)
}

// Global plugin registry
var pluginRegistry []PluginRegistration

// RegisterPlugin registers a new plugin
func RegisterPlugin(name string, register func(client *core.Client, db *core.Database)) {
	pluginRegistry = append(pluginRegistry, PluginRegistration{
		Name:     name,
		Register: register,
	})
}

// LoadAllPlugins loads all registered plugins
func LoadAllPlugins(client *core.Client, db *core.Database) {
	log.Println(">> Loading plugins...")
	
	loadedCount := 0
	for _, plugin := range pluginRegistry {
		log.Printf("   ✅ Loading: %s", plugin.Name)
		plugin.Register(client, db)
		loadedCount++
	}
	
	log.Printf("✅ Loaded %d plugins successfully", loadedCount)
}
