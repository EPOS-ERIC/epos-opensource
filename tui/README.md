# TUI Package

This package implements the Terminal User Interface for the EPOS CLI using [tview](https://github.com/rivo/tview).

## Design Pattern: Central Coordinator

The package is organized around the `App` struct (`app.go`), which acts as a central coordinator. Unlike a traditional MVC where components might talk to each other directly, here components like `EnvList` and `DetailsPanel` primarily interact with the `App` to trigger state changes, navigation, or background tasks.

**Why?** This keeps individual components decoupled and ensures that global concerns—like which page is visible, what is currently focused, and UI thread safety—are handled in one place.

## Core Mechanisms

### Focus & Navigation Stack
Focus management is the most complex part of a multi-page TUI. We use a manual stack (`focusStack`) to handle this:
- **Pushing Focus**: When opening a modal or form, the current primitive is pushed onto the stack.
- **Restoring Focus**: `ResetToHome` pops the stack to return the user exactly where they were (e.g., a specific item in a list) after a workflow completes or is cancelled.

### State Synchronization
The TUI does not rely on local state for environment data. Instead:
- A background ticker in `App` periodically triggers `Refresh()` on components.
- Components pull the latest data directly from the `db` package.
- This ensures the UI stays consistent even if the environment state changes via external CLI commands or background processes.

### Thread-Safe I/O Capture
Long-running operations (Deploy, Update, etc.) execute in separate goroutines to keep the UI responsive.
- **Output Capture**: The `OutputWriter` redirects stdout/stderr from core commands to a buffer.
- **UI Updates**: All background tasks must use `a.tview.QueueUpdateDraw` when updating UI components to ensure thread safety during the `tview` event loop.

## Standardized Workflow Patterns

Most operations follow a strict **"Trigger -> Confirmation/Form -> Progress -> Result"** flow.

- **Confirmations**: `ShowConfirmation` provides a consistent visual language for destructive actions.
- **Progress Runner**: `RunBackgroundTask` encapsulates the boiler-plate of starting a goroutine, switching to the `OperationProgress` view, and showing the final success/error overlay.

## Styling
Consistency is enforced through **`components.go`** (factory functions) and **`theme.go`** (color tokens). Avoid direct primitive initialization or hardcoded colors; use the `NewStyled*` factories to ensure the application maintains a cohesive look and feel.
