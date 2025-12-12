package tui

import (
	"fmt"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// populateState holds the state of the populate form.
type populateState struct {
	paths    []string
	examples bool
	inputs   []*tview.InputField
	buttons  []*tview.Button
}

// showPopulateForm displays the dynamic populate form.
func (a *App) showPopulateForm() {
	a.previousFocus = a.tview.GetFocus()
	envName := a.SelectedDockerEnv()
	if envName == "" {
		return
	}

	a.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])

	// Initial state with one empty path
	state := &populateState{
		paths:    []string{""},
		examples: true,
	}

	// Container for the dynamic form
	formFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Wrapper to center and style
	container := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formFlex, 0, 1, true)

	container.SetBorder(true).
		SetBorderColor(DefaultTheme.Primary).
		SetTitle(fmt.Sprintf(" [::b]Populate: %s ", envName)).
		SetTitleColor(DefaultTheme.Secondary)

	// Function to rebuild the UI based on state
	var rebuildUI func()
	rebuildUI = func() {
		formFlex.Clear()
		state.inputs = nil
		state.buttons = nil

		// 1. Path Rows
		for i, path := range state.paths {
			idx := i // Capture loop variable via closure

			input := tview.NewInputField().
				SetLabel(fmt.Sprintf("Path %d ", i+1)).
				SetText(path).
				SetFieldWidth(40).
				SetChangedFunc(func(text string) {
					if idx < len(state.paths) {
						state.paths[idx] = text
					}
				})
			input.SetFieldBackgroundColor(DefaultTheme.Surface).
				SetFieldTextColor(DefaultTheme.Secondary).
				SetLabelColor(DefaultTheme.Secondary)

			browseBtn := tview.NewButton("Browse").SetSelectedFunc(func() {
				a.showFilePicker(state.paths[idx], func(selectedPaths []string) {
					if len(selectedPaths) == 0 {
						return
					}
					// Update current input with first path
					state.paths[idx] = selectedPaths[0]

					// Add remaining paths as new rows
					if len(selectedPaths) > 1 {
						state.paths = append(state.paths, selectedPaths[1:]...)
					}

					rebuildUI()

					// Restore focus to the input of the row we just modified
					if idx < len(state.inputs) {
						a.tview.SetFocus(state.inputs[idx])
					}
				})
			})
			browseBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
			browseBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

			row := tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(input, 0, 1, true).
				AddItem(tview.NewBox(), 1, 0, false).
				AddItem(browseBtn, 10, 0, false)

			formFlex.AddItem(row, 1, 0, true).
				AddItem(tview.NewBox(), 1, 0, false)

			state.inputs = append(state.inputs, input)
			state.buttons = append(state.buttons, browseBtn) // Only browse buttons here
		}

		// 2. Add Path Button
		addPathBtn := tview.NewButton("Add Path").SetSelectedFunc(func() {
			state.paths = append(state.paths, "")
			rebuildUI()
			// Focus the new input
			if len(state.inputs) > 0 {
				a.tview.SetFocus(state.inputs[len(state.inputs)-1])
			}
		})
		addPathBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Surface).Foreground(DefaultTheme.Secondary))

		// 3. Examples Checkbox
		checkbox := tview.NewCheckbox().
			SetLabel("Populate Examples ").
			SetChecked(state.examples).
			SetChangedFunc(func(checked bool) {
				state.examples = checked
			})
		checkbox.SetLabelColor(DefaultTheme.Secondary).
			SetFieldBackgroundColor(DefaultTheme.Surface).
			SetFieldTextColor(DefaultTheme.Secondary)

		// 4. Action Buttons
		populateBtn := tview.NewButton("Populate").SetSelectedFunc(func() {
			a.handlePopulate(envName, state)
		})
		populateBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		populateBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

		cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
			a.returnFromPopulate()
		})
		cancelBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		cancelBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

		// Layout for controls
		controls := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(addPathBtn, 1, 0, false).
			AddItem(tview.NewBox(), 1, 0, false).
			AddItem(checkbox, 1, 0, false).
			AddItem(tview.NewBox(), 1, 0, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(populateBtn, 0, 1, false).
				AddItem(tview.NewBox(), 1, 0, false).
				AddItem(cancelBtn, 0, 1, false), 1, 0, false)

		formFlex.AddItem(controls, 0, 1, false)

		// Custom Input Capture for Focus Cycling
		var allFocusable []tview.Primitive
		for i := range state.inputs {
			allFocusable = append(allFocusable, state.inputs[i])
			if i < len(state.buttons) { // state.buttons now only contains browse buttons
				allFocusable = append(allFocusable, state.buttons[i])
			}
		}
		allFocusable = append(allFocusable, addPathBtn, checkbox, populateBtn, cancelBtn)

		container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				a.returnFromPopulate()
				return nil
			}
			if event.Key() == tcell.KeyTab {
				current := a.tview.GetFocus()
				for i, p := range allFocusable {
					if p == current {
						next := i + 1
						if next >= len(allFocusable) {
							next = 0
						}
						a.tview.SetFocus(allFocusable[next])
						return nil
					}
				}
				// If nothing focused (shouldn't happen), focus first
				if len(allFocusable) > 0 {
					a.tview.SetFocus(allFocusable[0])
				}
				return nil
			}
			if event.Key() == tcell.KeyBacktab {
				current := a.tview.GetFocus()
				for i, p := range allFocusable {
					if p == current {
						prev := i - 1
						if prev < 0 {
							prev = len(allFocusable) - 1
						}
						a.tview.SetFocus(allFocusable[prev])
						return nil
					}
				}
			}
			return event
		})
	}

	rebuildUI()

	// Center the layout manually (fixed size 60x20)
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(container, 20, 1, true).
		AddItem(nil, 0, 1, false)

	outerLayout := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, 60, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("populate-confirm", outerLayout, true, true)
	if len(state.inputs) > 0 {
		a.tview.SetFocus(state.inputs[0])
	}
}

// handlePopulate validates the form and starts population.
func (a *App) handlePopulate(envName string, state *populateState) {
	// Parse paths
	var validPaths []string
	for _, p := range state.paths {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			validPaths = append(validPaths, trimmed)
		}
	}

	// Validation: At least one path or examples must be selected
	if len(validPaths) == 0 && !state.examples {
		a.ShowError("You must provide at least one path OR enable examples.")
		return
	}

	a.pages.RemovePage("populate-confirm")
	a.showPopulateProgress(envName, validPaths, state.examples)
}

// showPopulateProgress displays the populate progress with live output.
func (a *App) showPopulateProgress(envName string, paths []string, examples bool) {
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { a.tview.Draw() })
	outputView.SetBorder(true).
		SetTitle(fmt.Sprintf(" [::b]Populating: %s ", envName)).
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
		SetBorderPadding(0, 0, 2, 2)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(DefaultTheme.SecondaryTag("") + "Populating... Please wait" + "[-]")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)
	layout.SetBackgroundColor(DefaultTheme.Background)

	// Connect output writer
	a.outputWriter.ClearBuffer()
	a.outputWriter.SetView(a.tview, outputView)

	a.pages.AddAndSwitchToPage("populate-progress", layout, true)
	a.UpdateFooter("[Populating]", KeyDescriptions["populating"])

	// Run populate in background
	go func() {
		_, err := dockercore.Populate(dockercore.PopulateOpts{
			Name:             envName,
			TTLDirs:          paths,
			PopulateExamples: examples,
			Parallel:         1, // Default to 1 as per TUI simplicity
		})

		a.tview.QueueUpdateDraw(func() {
			if err != nil {
				statusBar.SetText(fmt.Sprintf("%sPopulate failed: %v[-]", DefaultTheme.ErrorTag(""), err))
			} else {
				statusBar.SetText(DefaultTheme.SuccessTag("") + "Environment populated successfully!" + "[-]")
			}

			layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
					a.outputWriter.ClearView()
					a.returnFromPopulate()
					return nil
				}
				return event
			})
			a.UpdateFooter("[Populate Complete]", KeyDescriptions["populate-complete"])
		})
	}()
}

// returnFromPopulate cleans up and returns to the home screen.
func (a *App) returnFromPopulate() {
	a.pages.RemovePage("populate-confirm")
	a.pages.RemovePage("populate-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()

	if a.previousFocus != nil {
		a.tview.SetFocus(a.previousFocus)
	}
}
