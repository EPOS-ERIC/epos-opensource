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
	browseBtn *tview.Button
}

// showPopulateForm displays the dynamic populate form.
func (a *App) showPopulateForm() {
	a.PushFocus()
	envName, isDocker := a.envList.GetSelected()

	if envName == "" {
		return
	}

	a.UpdateFooter(GetFooterText(PopulateFormKey), PopulateFormKey)

	// Initial state with one empty path
	state := &populateState{
		paths:    []string{""},
		examples: false,
	}

	// Layout
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

		for i, path := range state.paths {
			idx := i
			input := NewStyledInputField(fmt.Sprintf("Path %d ", i+1), path).
				SetChangedFunc(func(text string) {
					if idx < len(state.paths) {
						state.paths[idx] = text
					}
				})

			size := 1
			if i == 0 {
				input.SetBorderPadding(1, 0, 1, 1)
				size = 2
			} else {
				input.SetBorderPadding(0, 0, 1, 1)
			}
			formFlex.AddItem(input, size, 0, true).
				AddItem(tview.NewBox(), 1, 0, false)

			state.inputs = append(state.inputs, input)
		}

		addPathBtn := NewStyledInactiveButton("Add Path", func() {
			state.paths = append(state.paths, "")
			rebuildUI()
			if len(state.inputs) > 0 {
				a.tview.SetFocus(state.inputs[len(state.inputs)-1])
			}
		})

		browseBtn := NewStyledInactiveButton("Browse Files", func() {
			var currentPaths []string
			for _, p := range state.paths {
				if strings.TrimSpace(p) != "" {
					currentPaths = append(currentPaths, p)
				}
			}

			a.showFilePicker(currentPaths, func(selectedPaths []string) {
				state.paths = selectedPaths
				if len(state.paths) == 0 {
					state.paths = []string{""}
				}
				rebuildUI()
				if state.browseBtn != nil {
					a.tview.SetFocus(state.browseBtn)
				}
			})
		})
		state.browseBtn = browseBtn

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

		populateBtn := NewStyledButton("Populate", func() {
			a.handlePopulate(envName, state, isDocker)
		})

		cancelBtn := NewStyledButton("Cancel", func() {
			a.ResetToHome(ResetOptions{
				PageNames:    []string{"populate"},
				RefreshFiles: true,
				RestoreFocus: true,
			})
		})

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
		}
		allFocusable = append(allFocusable, browseBtn, addPathBtn, checkbox, populateBtn, cancelBtn)

		formFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				a.ResetToHome(ResetOptions{
					PageNames:    []string{"populate"},
					RefreshFiles: true,
					RestoreFocus: true,
				})
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

	a.pages.AddPage("populate", CenterPrimitive(formFlex, 1, 2), true, true)
	a.currentPage = "populate"
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

	a.pages.RemovePage("populate")
	a.showPopulateProgress(envName, validPaths, state.examples, isDocker)
}

// showPopulateProgress displays the populate progress with live output.
func (a *App) showPopulateProgress(envName string, paths []string, examples bool, isDocker bool) {
	a.RunBackgroundTask(TaskOptions{
		Operation: "Populate",
		EnvName:   envName,
		IsDocker:  isDocker,
		Task: func() (string, error) {
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
			return "", err
		},
	})
}
