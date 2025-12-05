# TUI Sample Package

This package provides a simple sample implementation of a Terminal User Interface (TUI) using the `tview` library. It demonstrates patterns for building extensible, testable TUIs by separating concerns between the app (global state and UI orchestration) and screens (modular UI components). The code follows repo conventions: config integration (via `config/`), simple error handling, and debug logging with the `log` package. Use this as a guide to develop the rest of the TUI consistently.

## Architecture Overview

The TUI follows a modular architecture:

- **App**: Manages global state, UI primitives (e.g., pages, frame), and navigation.
- **Screens**: Individual UI components (e.g., home, deploy) that handle their own events and updates.
- **Events**: Screens process key events and return `Event` constants (e.g., `EventSwitchDeploy`) for the app to execute.

This separation ensures screens are reusable and the app remains focused on orchestration. Interfaces enable testability by hiding `tview` dependencies.

## Key Components

- **`App`**: The main struct in `app.go`. Registers screens, switches between them, updates the footer, and runs the TUI loop. Why? Centralizes shared logic for consistency and ease of extension.
- **`Screen` Interface**: Defined in `screen.go`. Methods include `Name()` (identifier), `Init()` (setup UI), `HandleEvent()` (process keys), and `UpdateFooter()` (set context). Why? Standardizes screen behavior and allows polymorphism.
- **`AppInterface`**: A subset of `App` methods (e.g., `SwitchTo`, `UpdateFooter`). Why? Enables dependency injection for screens, making them testable with mocks (no real `tview` needed).

Example: `HomeScreen` in `home.go` handles the 'd' key to navigate:

```go
func (h *HomeScreen) HandleEvent(event *tcell.EventKey, app AppInterface) (bool, Event) {
    if event.Rune() == 'd' {
        return true, EventSwitchDeploy
    }
    return false, ""
}
```

## Patterns and Best Practices

- **Dependency Injection**: Pass `AppInterface` to screens for app interactions. Why? Decouples screens from the full app, improving testability.
- **Events**: Return `Event` constants (e.g., `EventSwitchDeploy`) from `HandleEvent`. Why? Improves type safety and maintainability over raw strings, while keeping event logic in screens and actions in the app.
- **Footer Updates**: Screens call `app.UpdateFooter(sectionText, keys)` with relevant info. Why? Provides context-specific help without hardcoding in the app.
- **Type Assertions**: Use `if realApp, ok := app.(*App)` for `tview`-specific code. Why? Allows UI setup in real apps but skips in tests (mocks don't match).
- **Testing**: Use mocks (e.g., `mockApp` in tests) and table-driven tests. Assert behaviors like returned commands. Why? Ensures reliability without UI dependencies.
- **Configuration**: Set up for future integration via `config/model.go` and `default.yaml`. Pass `*config.Config` to `NewApp` (nil for tests). Why? Allows customization without hardcoding.
- **Logging**: Uses `log.Printf` for debug info (e.g., unhandled keys, switches). Why? Matches repo's simple logging for debugging.

## Adding New Screens

1. Create a new struct (e.g., `MyScreen`) implementing the `Screen` interface.
2. In `Init`, set up UI primitives (use type assertion for `tview`).
3. Handle events in `HandleEvent` (return commands for navigation).
4. Update footer in `UpdateFooter`.
5. Register the screen in the app (e.g., `app.RegisterScreen(&MyScreen{})`).

Example: A basic "Settings" screen:

```go
type SettingsScreen struct{}

func (s *SettingsScreen) Name() string { return "settings" }

func (s *SettingsScreen) Init(app AppInterface) {
    // Setup UI (e.g., text view)
    if realApp, ok := app.(*App); ok {
        // Add to pages
    }
}

func (s *SettingsScreen) HandleEvent(event *tcell.EventKey, app AppInterface) (bool, Event) {
    if event.Key() == tcell.KeyEsc {
        return true, EventSwitchHome
    }
    return false, ""
}

func (s *SettingsScreen) UpdateFooter(app AppInterface) {
    app.UpdateFooter("Settings", []string{"esc: back"})
}
```

## Testing Guidelines

Follow the patterns in `*_test.go`: Use `mockApp` for isolation, table-driven tests for events, and assert key behaviors (e.g., commands returned). Why? Keeps tests fast and focused on logic.

This package is a foundation—extend it by adding screens that follow these patterns for a cohesive TUI.</content>
<parameter name="filePath">/Users/marco/epos-cli/tui/sample/README.md

