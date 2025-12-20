package tui

import (
	"fmt"
	"strings"

	"github.com/epos-eu/epos-opensource/cmd/docker/dockercore"
	"github.com/epos-eu/epos-opensource/cmd/k8s/k8score"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// populateState holds the state of the populate form.
type populateState struct {
	paths     []string
	examples  bool
	inputs    []*tview.InputField
	buttons   []*tview.Button // Legacy, unused for per-row
	browseBtn *tview.Button
}

// showPopulateForm displays the dynamic populate form.
func (a *App) showPopulateForm() {
	a.previousFocus = a.tview.GetFocus()
	isDocker := a.currentEnv == a.dockerFlex
	envName := ""
	if isDocker {
		envName = a.SelectedDockerEnv()
	} else {
		envName = a.SelectedK8sEnv()
	}

	if envName == "" {
		return
	}

	a.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])

	// Initial state with one empty path
	state := &populateState{
		paths:    []string{""},
		examples: false,
	}

	// Container for the dynamic form with border and styling
	formFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	formFlex.SetBorder(true).
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
				SetFieldWidth(0).
				SetChangedFunc(func(text string) {
					if idx < len(state.paths) {
						state.paths[idx] = text
					}
				})
			input.SetFieldBackgroundColor(DefaultTheme.Surface).
				SetFieldTextColor(DefaultTheme.Secondary)
			if i == 0 {
				input.SetLabelColor(DefaultTheme.Secondary).SetBorderPadding(1, 0, 1, 1)
			} else {
				input.SetLabelColor(DefaultTheme.Secondary).SetBorderPadding(0, 0, 1, 1)
			}

			// Row with just input
			row := tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(input, 0, 1, true)

			size := 1
			if i == 0 {
				size = 2
			}
			formFlex.AddItem(row, size, 0, true).
				AddItem(tview.NewBox(), 1, 0, false)

			state.inputs = append(state.inputs, input)
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
		addPathBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

		// 3. Browse Files Button
		browseBtn := tview.NewButton("Browse Files").SetSelectedFunc(func() {
			// Gather non-empty paths to pre-select, or allow empty path list
			var currentPaths []string
			for _, p := range state.paths {
				if strings.TrimSpace(p) != "" {
					currentPaths = append(currentPaths, p)
				}
			}

			a.showFilePicker(currentPaths, func(selectedPaths []string) {
				// We replace state.paths with selection.
				// If selection is empty, maybe keep one empty row?
				state.paths = selectedPaths
				if len(state.paths) == 0 {
					state.paths = []string{""}
				}
				rebuildUI()

				// Restore focus to the browse button
				if state.browseBtn != nil {
					a.tview.SetFocus(state.browseBtn)
				}
			})
		})
		browseBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Surface).Foreground(DefaultTheme.Secondary))
		browseBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))
		state.browseBtn = browseBtn

		// 3. Examples Checkbox
		checkbox := tview.NewCheckbox().
			SetLabel("Populate Examples ").
			SetChecked(state.examples).
			SetChangedFunc(func(checked bool) {
				state.examples = checked
			})
		checkbox.SetLabelColor(DefaultTheme.Secondary).
			SetFieldBackgroundColor(DefaultTheme.Surface).
			SetFieldTextColor(DefaultTheme.Secondary).
			SetBorderPadding(0, 0, 1, 1)

		// 4. Action Buttons
		populateBtn := tview.NewButton("Populate").SetSelectedFunc(func() {
			a.handlePopulate(envName, state, isDocker)
		})
		populateBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		populateBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

		cancelBtn := tview.NewButton("Cancel").SetSelectedFunc(func() {
			a.returnFromPopulate()
		})
		cancelBtn.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
		cancelBtn.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

		controls := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox(), 0, 1, false).
				AddItem(browseBtn, 16, 1, false).
				AddItem(tview.NewBox(), 0, 1, false).
				AddItem(addPathBtn, 12, 1, false).
				AddItem(tview.NewBox(), 0, 1, false), 1, 0, false).
			AddItem(tview.NewBox(), 1, 0, false).
			AddItem(checkbox, 1, 0, false).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox(), 0, 1, false).
				AddItem(populateBtn, 12, 1, false).
				AddItem(tview.NewBox(), 2, 0, false).
				AddItem(cancelBtn, 10, 1, false).
				AddItem(tview.NewBox(), 0, 1, false), 1, 0, false)

		formFlex.AddItem(controls, 0, 1, false)

		// Custom Input Capture for Focus Cycling
		var allFocusable []tview.Primitive
		for i := range state.inputs {
			allFocusable = append(allFocusable, state.inputs[i])
			if i < len(state.buttons) {
				allFocusable = append(allFocusable, state.buttons[i])
			}
		}
		allFocusable = append(allFocusable, browseBtn, addPathBtn, checkbox, populateBtn, cancelBtn)

		formFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

	a.pages.AddPage("populate-confirm", CenterPrimitive(formFlex, 1, 2), true, true)
	if len(state.inputs) > 0 {
		a.tview.SetFocus(state.inputs[0])
	}
}

// handlePopulate validates the form and starts population.
func (a *App) handlePopulate(envName string, state *populateState, isDocker bool) {
	// Parse paths
	var validPaths []string
	for _, p := range state.paths {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			validPaths = append(validPaths, trimmed)
		}
	}

	a.pages.RemovePage("populate-confirm")
	a.showPopulateProgress(envName, validPaths, state.examples, isDocker)
}

// showPopulateProgress displays the populate progress with live output.
func (a *App) showPopulateProgress(envName string, paths []string, examples bool, isDocker bool) {
	progress := NewOperationProgress(a, "Populate", envName)
	progress.Start()

	// Run populate in background
	go func() {
		var err error
		if isDocker {
			_, err = dockercore.Populate(dockercore.PopulateOpts{
				Name:             envName,
				TTLDirs:          paths,
				PopulateExamples: examples,
				Parallel:         1,
			})
		} else {
			_, err = k8score.Populate(k8score.PopulateOpts{
				Name:             envName,
				TTLDirs:          paths,
				PopulateExamples: examples,
				Parallel:         1,
			})
		}

		if err != nil {
			progress.Complete(false, err.Error())
		} else {
			progress.Complete(true, "Environment populated successfully!")
		}
	}()
}

// returnFromPopulate cleans up and returns to the home screen.
func (a *App) returnFromPopulate() {
	a.pages.RemovePage("populate-confirm")
	a.pages.RemovePage("populate-progress")
	a.pages.SwitchToPage("home")
	a.refreshLists()
	a.refreshIngestedFiles()

	if a.previousFocus != nil {
		a.tview.SetFocus(a.previousFocus)
	}
	if a.detailsShown {
		key := DetailsK8sKey
		if a.currentEnv == a.dockerFlex {
			key = DetailsDockerKey
		}
		a.UpdateFooter("[Environment Details]", KeyDescriptions[key])
	} else {
		if a.currentEnv == a.dockerFlex {
			a.UpdateFooter("[Docker Environments]", KeyDescriptions["docker"])
		} else {
			a.UpdateFooter("[K8s Environments]", KeyDescriptions["k8s"])
		}
	}
}
