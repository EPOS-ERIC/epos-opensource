package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FilePicker is a component to select files or directories.
type FilePicker struct {
	app        *App
	view       *tview.TreeView
	root       *tview.TreeNode
	marked     map[string]bool
	onSelect   func([]string)
	onChanged  func(path string)
	lastSearch string
}

// newFilePicker creates a new file picker.
func newFilePicker(app *App, initialPath string, selectedPaths []string, onSelect func([]string)) *FilePicker {
	rootPath := "/"
	rootNode := tview.NewTreeNode(rootPath).SetReference(rootPath).SetColor(DefaultTheme.Primary)

	picker := &FilePicker{
		app:      app,
		view:     tview.NewTreeView().SetRoot(rootNode).SetCurrentNode(rootNode),
		root:     rootNode,
		marked:   make(map[string]bool),
		onSelect: onSelect,
	}
	picker.view.SetBorderPadding(0, 0, 3, 3).SetBackgroundColor(DefaultTheme.Background)

	for _, p := range selectedPaths {
		picker.marked[p] = true
	}

	picker.updateNodeText(rootNode)
	picker.addNodes(rootNode, rootPath)

	picker.expandTo(initialPath)
	for _, p := range selectedPaths {
		picker.expandTo(p)
	}

	// Set selection to current node from initial path
	pathElements := strings.Split(initialPath, string(os.PathSeparator))
	currentNode := rootNode
	for _, elem := range pathElements {
		if elem == "" {
			continue
		}
		for _, child := range currentNode.GetChildren() {
			ref := child.GetReference()
			if ref != nil && filepath.Base(ref.(string)) == elem {
				currentNode = child
				break
			}
		}
	}
	picker.view.SetCurrentNode(currentNode)

	picker.view.SetChangedFunc(func(node *tview.TreeNode) {
		if node != nil && node.GetReference() != nil {
			path := node.GetReference().(string)
			if picker.onChanged != nil {
				picker.onChanged(path)
			}
		}
	})

	return picker
}

// updateNodeText updates the visual representation of a node (checkbox and name).
func (f *FilePicker) updateNodeText(n *tview.TreeNode) {
	ref := n.GetReference()
	if ref == nil {
		return
	}
	path := ref.(string)
	name := filepath.Base(path)
	if path == "/" {
		name = "/"
	}

	pColor := DefaultTheme.Hex(DefaultTheme.Primary)
	sColor := DefaultTheme.Hex(DefaultTheme.Success)
	box := fmt.Sprintf("[%s]%s[-]", pColor, tview.Escape("[ ]"))
	if f.marked[path] {
		box = fmt.Sprintf("[%s]%s[-]", sColor, tview.Escape("[✓]"))
	}
	n.SetText(box + " " + name)
}

// expandTo expands the tree to the specified path.
func (f *FilePicker) expandTo(path string) {
	pathElements := strings.Split(path, string(os.PathSeparator))
	currentNode := f.root

	for _, elem := range pathElements {
		if elem == "" {
			continue
		}

		var foundNode *tview.TreeNode
		for _, child := range currentNode.GetChildren() {
			ref := child.GetReference()
			if ref != nil && filepath.Base(ref.(string)) == elem {
				foundNode = child
				break
			}
		}

		if foundNode != nil {
			fullPath := foundNode.GetReference().(string)
			f.addNodes(foundNode, fullPath)
			foundNode.SetExpanded(true)
			currentNode = foundNode
		} else {
			break
		}
	}
}

// toggleMark toggles the selection state of a node.
func (f *FilePicker) toggleMark(node *tview.TreeNode) {
	ref := node.GetReference()
	if ref == nil {
		return
	}
	path := ref.(string)
	f.marked[path] = !f.marked[path]
	f.updateNodeText(node)
}

// findParent searches for the parent of a node.
func (f *FilePicker) findParent(node *tview.TreeNode) *tview.TreeNode {
	if node == f.root {
		return nil
	}
	var parent *tview.TreeNode
	var search func(*tview.TreeNode) bool
	search = func(n *tview.TreeNode) bool {
		for _, child := range n.GetChildren() {
			if child == node {
				parent = n
				return true
			}
			if child.IsExpanded() {
				if search(child) {
					return true
				}
			}
		}
		return false
	}
	search(f.root)
	return parent
}

// submit returns the selected paths and closes the picker.
func (f *FilePicker) submit() {
	var result []string
	for p, m := range f.marked {
		if m {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		if node := f.view.GetCurrentNode(); node != nil {
			if ref := node.GetReference(); ref != nil {
				result = append(result, ref.(string))
			}
		}
	}
	f.app.pages.RemovePage("file-picker")
	f.app.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])
	if f.onSelect != nil {
		f.onSelect(result)
	}
}

// findAllMatches finds all nodes matching the search text.
func (f *FilePicker) findAllMatches(text string) []*tview.TreeNode {
	var matches []*tview.TreeNode
	if text == "" {
		return matches
	}
	text = strings.ToLower(text)

	var search func(n *tview.TreeNode)
	search = func(n *tview.TreeNode) {
		ref := n.GetReference()
		if ref != nil {
			name := filepath.Base(ref.(string))
			if strings.Contains(strings.ToLower(name), text) {
				if n != f.root || len(text) > 1 {
					matches = append(matches, n)
				}
			}
		}
		if n.IsExpanded() {
			for _, child := range n.GetChildren() {
				search(child)
			}
		}
	}
	search(f.root)
	return matches
}

// executeSearch performs a search and updates the UI.
func (f *FilePicker) executeSearch(text string, searchField *tview.InputField) {
	if text == "" {
		return
	}
	f.lastSearch = text
	matches := f.findAllMatches(text)

	if len(matches) > 0 {
		f.view.SetCurrentNode(matches[0])
		f.app.tview.SetFocus(f.view)
		// UX: update footer with match count
		f.app.UpdateFooter("[File Picker]", []string{fmt.Sprintf("match 1/%d", len(matches)), "n: next", "N: prev", "/: search"})
	} else {
		searchField.SetLabel("Not found: ")
		searchField.SetLabelColor(DefaultTheme.Error)
	}
}

// moveSelection moves the current node selection by a step (e.g., for mouse wheel).
func (f *FilePicker) moveSelection(step int) {
	var visibleNodes []*tview.TreeNode
	var traverse func(n *tview.TreeNode)
	traverse = func(n *tview.TreeNode) {
		visibleNodes = append(visibleNodes, n)
		if n.IsExpanded() {
			for _, child := range n.GetChildren() {
				traverse(child)
			}
		}
	}
	traverse(f.root)

	current := f.view.GetCurrentNode()
	index := -1
	for i, n := range visibleNodes {
		if n == current {
			index = i
			break
		}
	}

	if index != -1 {
		newIndex := index + step
		if newIndex >= 0 && newIndex < len(visibleNodes) {
			f.view.SetCurrentNode(visibleNodes[newIndex])
		}
	}
}

// setupInput configures keyboard navigation and actions.
func (f *FilePicker) setupInput(layout *tview.Flex, searchField *tview.InputField) {
	f.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		key := event.Key()
		rune := event.Rune()

		if key == tcell.KeyRune && rune == '/' {
			searchField.SetLabel("/")
			searchField.SetLabelColor(DefaultTheme.Secondary)
			layout.ResizeItem(searchField, 1, 0)
			f.app.tview.SetFocus(searchField)
			f.app.UpdateFooter("[Search]", []string{"enter: find", "esc: cancel"})
			return nil
		}

		if f.lastSearch != "" && (rune == 'n' || rune == 'N') {
			matches := f.findAllMatches(f.lastSearch)
			if len(matches) == 0 {
				return nil
			}
			currentNode := f.view.GetCurrentNode()
			currentIndex := -1
			for i, m := range matches {
				if m == currentNode {
					currentIndex = i
					break
				}
			}

			var nextIndex int
			if rune == 'N' {
				nextIndex = currentIndex - 1
				if nextIndex < 0 {
					nextIndex = len(matches) - 1
				}
			} else {
				nextIndex = currentIndex + 1
				if nextIndex >= len(matches) {
					nextIndex = 0
				}
			}

			if currentIndex == -1 {
				nextIndex = 0
			}

			f.view.SetCurrentNode(matches[nextIndex])
			// UX: update match count in footer
			f.app.UpdateFooter("[File Picker]", []string{fmt.Sprintf("match %d/%d", nextIndex+1, len(matches)), "n: next", "N: prev", "/: search"})
			return nil
		}

		switch {
		case key == tcell.KeyEsc:
			if f.lastSearch != "" || searchField.GetText() != "" {
				searchField.SetText("")
				searchField.SetLabel("/")
				searchField.SetLabelColor(DefaultTheme.Secondary)
				layout.ResizeItem(searchField, 0, 0)
				f.lastSearch = ""
				return nil
			}
			f.app.pages.RemovePage("file-picker")
			f.app.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])
			return nil

		case key == tcell.KeyEnter:
			f.submit()
			return nil

		case key == tcell.KeyRight:
			node := f.view.GetCurrentNode()
			if node != nil {
				if ref := node.GetReference(); ref != nil {
					path := ref.(string)
					info, err := os.Stat(path)
					if err == nil && info.IsDir() {
						if len(node.GetChildren()) == 0 {
							f.addNodes(node, path)
						}
						node.SetExpanded(true)
					}
				}
			}
			return nil

		case key == tcell.KeyLeft:
			node := f.view.GetCurrentNode()
			if node != nil {
				if node.IsExpanded() {
					node.SetExpanded(false)
				} else if parent := f.findParent(node); parent != nil {
					f.view.SetCurrentNode(parent)
				}
			}
			return nil

		case key == tcell.KeyRune && rune == ' ':
			if node := f.view.GetCurrentNode(); node != nil {
				f.toggleMark(node)
			}
			return nil
		}

		return event
	})

	searchField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			f.executeSearch(searchField.GetText(), searchField)
		case tcell.KeyEsc:
			searchField.SetText("")
			searchField.SetLabel("/")
			searchField.SetLabelColor(DefaultTheme.Secondary)
			layout.ResizeItem(searchField, 0, 0)
			f.lastSearch = ""
			f.app.tview.SetFocus(f.view)
			f.app.UpdateFooter("[File Picker]", KeyDescriptions["file-picker"])
		}
	})
}

// addNodes adds child nodes based on directory content.
func (f *FilePicker) addNodes(target *tview.TreeNode, path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		}
		if !files[i].IsDir() && files[j].IsDir() {
			return false
		}
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if len(file.Name()) > 0 && file.Name()[0] == '.' {
			continue
		}

		fullPath := filepath.Join(path, file.Name())
		node := tview.NewTreeNode("").SetReference(fullPath).SetSelectable(true)
		f.updateNodeText(node)

		if file.IsDir() {
			node.SetColor(DefaultTheme.Secondary)
		} else {
			node.SetColor(DefaultTheme.OnSurface)
		}
		node.SetSelectedTextStyle(tcell.StyleDefault.Foreground(DefaultTheme.Primary).Background(DefaultTheme.Secondary))

		target.AddChild(node)
	}
}

// showFilePicker displays the file picker modal.
func (a *App) showFilePicker(initialPaths []string, onSelect func([]string)) {
	absPath := a.resolveStartPath(initialPaths)
	picker := newFilePicker(a, absPath, initialPaths, onSelect)

	a.UpdateFooter("[File Picker]", KeyDescriptions["file-picker"])

	picker.view.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref == nil {
			return
		}
		path := ref.(string)

		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			if len(node.GetChildren()) == 0 {
				picker.addNodes(node, path)
			}
			node.SetExpanded(!node.IsExpanded())
		} else {
			picker.toggleMark(node)
		}
	})

	btnSubmit := tview.NewButton("Submit").SetSelectedFunc(func() { picker.submit() })
	ApplyButtonStyle(btnSubmit)

	btnCancel := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.pages.RemovePage("file-picker")
		a.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])
	})
	btnCancel.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Surface).Foreground(DefaultTheme.OnSurface))
	btnCancel.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Error).Foreground(DefaultTheme.OnError))

	buttonBar := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(btnSubmit, 10, 0, false).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(btnCancel, 10, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	searchField := tview.NewInputField().
		SetLabel("/").
		SetFieldWidth(0).
		SetLabelColor(DefaultTheme.Secondary).
		SetFieldBackgroundColor(DefaultTheme.Background).
		SetFieldTextColor(DefaultTheme.OnSurface)

	pathBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	pathBar.SetBackgroundColor(DefaultTheme.Surface)
	pathBar.SetText(" [::b]Path:[-] " + absPath)

	picker.onChanged = func(path string) {
		pathBar.SetText(" [::b]Path:[-] " + path)
	}

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pathBar, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(picker.view, 0, 1, true).
		AddItem(searchField, 0, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(buttonBar, 1, 0, false)

	layout.SetBorder(true).
		SetBorderColor(DefaultTheme.Primary).
		SetTitle(" [::b]Select Files ").
		SetTitleColor(DefaultTheme.Secondary)

	// UX: Selection follows scroll wheel
	picker.view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case tview.MouseScrollUp:
			picker.moveSelection(-1)
		case tview.MouseScrollDown:
			picker.moveSelection(1)
		}
		return action, event
	})

	picker.setupInput(layout, searchField)

	pickerLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(layout, 0, 1, true)

	a.pages.AddPage("file-picker", CenterPrimitive(pickerLayout, 2, 4), true, true)
	a.tview.SetFocus(picker.view)
}

// resolveStartPath determines the initial directory for the picker.
func (a *App) resolveStartPath(initialPaths []string) string {
	startPath := ""
	if len(initialPaths) > 0 && initialPaths[0] != "" {
		info, err := os.Stat(initialPaths[0])
		if err == nil {
			if info.IsDir() {
				startPath = initialPaths[0]
			} else {
				startPath = filepath.Dir(initialPaths[0])
			}
		}
	}

	if startPath == "" {
		cwd, err := os.Getwd()
		if err == nil {
			startPath = cwd
		} else {
			startPath = "/"
		}
	}

	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "/"
	}
	return absPath
}
