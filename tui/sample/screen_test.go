package sample

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// TestHomeScreen_Name tests that HomeScreen returns the correct identifier.
// This ensures screens are registered with the right name.
func TestHomeScreen_Name(t *testing.T) {
	screen := &HomeScreen{}
	assert.Equal(t, "home", screen.Name())
}

// TestHomeScreen_HandleEvent tests event handling for navigation keys.
// Verifies that 'd' triggers a switch command, while other keys are ignored.
// Uses table-driven tests for multiple scenarios.
func TestHomeScreen_HandleEvent(t *testing.T) {
	screen := &HomeScreen{}
	mockApp := newMockApp()

	tests := []struct {
		name     string
		event    *tcell.EventKey
		expected bool
		command  Event
	}{
		{"press d", tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), true, EventSwitchDeploy},
		{"press other", tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, cmd := screen.HandleEvent(tt.event, mockApp)
			assert.Equal(t, tt.expected, handled)
			assert.Equal(t, tt.command, cmd)
		})
	}
}

// TestHomeScreen_UpdateFooter tests that UpdateFooter calls the app with correct section text.
// Ensures the footer reflects the current screen's context.
func TestHomeScreen_UpdateFooter(t *testing.T) {
	screen := &HomeScreen{}
	mockApp := newMockApp()

	screen.UpdateFooter(mockApp)
	assert.Contains(t, mockApp.footerCalled, "Home Screen")
}

// TestDeployScreen_Name tests that DeployScreen returns the correct identifier.
func TestDeployScreen_Name(t *testing.T) {
	screen := &DeployScreen{}
	assert.Equal(t, "deploy", screen.Name())
}

// TestDeployScreen_HandleEvent tests event handling for escape key.
// Verifies ESC triggers a switch command to go back.
func TestDeployScreen_HandleEvent(t *testing.T) {
	screen := &DeployScreen{}
	mockApp := newMockApp()

	tests := []struct {
		name     string
		event    *tcell.EventKey
		expected bool
		command  Event
	}{
		{"press esc", tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), true, EventSwitchHome},
		{"press other", tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handled, cmd := screen.HandleEvent(tt.event, mockApp)
			assert.Equal(t, tt.expected, handled)
			assert.Equal(t, tt.command, cmd)
		})
	}
}

// TestDeployScreen_UpdateFooter tests that UpdateFooter calls the app with correct section text.
func TestDeployScreen_UpdateFooter(t *testing.T) {
	screen := &DeployScreen{}
	mockApp := newMockApp()

	screen.UpdateFooter(mockApp)
	assert.Contains(t, mockApp.footerCalled, "Deploy Screen")
}
