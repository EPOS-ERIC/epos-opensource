# TUI Package

Terminal user interface for managing EPOS environments using [tview](https://github.com/rivo/tview).

## Structure

```
tui/
├── app.go      # Core App struct, Run(), shared utilities
├── theme.go    # Colors, styles, key descriptions
├── home.go     # Home screen with environment lists
├── deploy.go   # Docker deploy form and progress
├── help.go     # Help modal
├── writer.go   # Output capture for command output
└── README.md
```

## Adding a New Screen

Each screen follows this pattern:

```go
// screen_name.go

// 1. Data struct (if the screen has form data)
type screenData struct {
    field1 string
    field2 bool
}

// 2. Show function - creates UI and adds page
func (a *App) showScreenName() {
    // Build primitives
    layout := tview.NewFlex()...

    // Add input capture for THIS screen's keys
    layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyEsc {
            a.returnFromScreenName()
            return nil
        }
        return event
    })

    // Add page and switch
    a.pages.AddAndSwitchToPage("screenName", layout, true)
    a.UpdateFooter("[Screen Title]", []string{"esc: back", "enter: confirm"})
}

// 3. Return function - cleanup and go back
func (a *App) returnFromScreenName() {
    a.pages.RemovePage("screenName")
    a.pages.SwitchToPage("home")
    // Restore focus
    a.tview.SetFocus(a.docker)
}

// 4. Action handlers
func (a *App) handleScreenAction(data *screenData) {
    // Validate, then do work
}
```

## Key APIs

| Function                       | Purpose                       |
| ------------------------------ | ----------------------------- |
| `a.UpdateFooter(title, keys)`  | Set footer text and shortcuts |
| `a.ShowError(message)`         | Show error modal              |
| `a.pages.AddAndSwitchToPage()` | Add and show a screen         |
| `a.pages.RemovePage()`         | Remove a screen               |
| `a.tview.SetFocus()`           | Set keyboard focus            |
| `CenterPrimitive(p, w, h)`     | Center a primitive on screen  |
