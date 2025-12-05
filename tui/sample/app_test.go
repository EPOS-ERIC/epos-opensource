package sample

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockApp implements AppInterface for testing screens.
// It records method calls to verify interactions without real tview components.
type mockApp struct {
	currentScreen Screen
	screens       map[string]Screen
	switchCalled  []string // tracks SwitchTo calls for verification
	footerCalled  []string // tracks UpdateFooter calls for verification
}

func newMockApp() *mockApp {
	return &mockApp{
		screens:      make(map[string]Screen),
		switchCalled: []string{},
		footerCalled: []string{},
	}
}

func (m *mockApp) RegisterScreen(screen Screen) {
	m.screens[screen.Name()] = screen
}

func (m *mockApp) SwitchTo(screenName string) {
	m.switchCalled = append(m.switchCalled, screenName)
	if screen, exists := m.screens[screenName]; exists {
		m.currentScreen = screen
	}
}

func (m *mockApp) UpdateFooter(sectionText string, keys []string) {
	m.footerCalled = append(m.footerCalled, sectionText)
}

// TestNewApp tests that NewApp initializes all fields correctly.
// This ensures the app is ready for screen registration and UI setup.
func TestNewApp(t *testing.T) {
	// Test that NewApp initializes all fields correctly.
	// This ensures the app is ready for screen registration and UI setup.
	app := NewApp(nil)
	screen := &HomeScreen{}
	app.RegisterScreen(screen)
	assert.Contains(t, app.screens, "home")
	assert.Equal(t, screen, app.screens["home"])
}

// TestSwitchTo tests switching between registered screens.
// Ensures currentScreen updates correctly and handles invalid names gracefully.
func TestSwitchTo(t *testing.T) {
	app := NewApp(nil)
	home := &HomeScreen{}
	deploy := &DeployScreen{}
	app.RegisterScreen(home)
	app.RegisterScreen(deploy)

	app.SwitchTo("home")
	assert.Equal(t, home, app.currentScreen)

	app.SwitchTo("deploy")
	assert.Equal(t, deploy, app.currentScreen)

	app.SwitchTo("nonexistent")
	assert.Equal(t, deploy, app.currentScreen) // unchanged
}
