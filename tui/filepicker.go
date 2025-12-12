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
	view   *tview.TreeView
	root   *tview.TreeNode
	marked map[string]bool
}

// newFilePicker creates a new file picker.
func newFilePicker(initialPath string) *FilePicker {
	rootPath := "/"

	rootNode := tview.NewTreeNode(rootPath).
		SetReference(rootPath).
		SetColor(DefaultTheme.Primary)

	picker := &FilePicker{
		view:   tview.NewTreeView().SetRoot(rootNode).SetCurrentNode(rootNode),
		root:   rootNode,
		marked: make(map[string]bool),
	}

	// Helper to format text
	updateText := func(n *tview.TreeNode, path string) {
		name := filepath.Base(path)
		if path == "/" {
			name = "/"
		}
		pColor := DefaultTheme.Hex(DefaultTheme.Primary)
		sColor := DefaultTheme.Hex(DefaultTheme.Success)
		box := fmt.Sprintf("[%s][ ][-]", pColor)
		if picker.marked[path] {
			box = fmt.Sprintf("[%s][x][-]", sColor)
		}
		n.SetText(box + " " + name)
	}

	updateText(rootNode, rootPath)

	picker.addNodes(rootNode, rootPath)

	// Auto-expand to initialPath
	pathElements := strings.Split(initialPath, string(os.PathSeparator))
	currentNode := rootNode

	for _, elem := range pathElements {
		if elem == "" {
			continue
		}

		// Find child matching this element (ignoring prefix in comparison)
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
			picker.addNodes(foundNode, fullPath)
			foundNode.SetExpanded(true)
			currentNode = foundNode
		} else {
			break
		}
	}
	picker.view.SetCurrentNode(currentNode)

	picker.view.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)
		children := node.GetChildren()
		if len(children) == 0 {
			picker.addNodes(node, path)
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	picker.view.SetBackgroundColor(DefaultTheme.Background)

	return picker
}

// addNodes adds child nodes to the given node based on the directory content.
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
		node := tview.NewTreeNode("").
			SetReference(fullPath).
			SetSelectable(true)

		// Set Text with Checkbox
		name := file.Name()
		pColor := DefaultTheme.Hex(DefaultTheme.Primary)
		sColor := DefaultTheme.Hex(DefaultTheme.Success)
		box := fmt.Sprintf("[%s][ ][-]", pColor)
		if f.marked[fullPath] {
			box = fmt.Sprintf("[%s][x][-]", sColor)
		}
		node.SetText(box + " " + name)

		if file.IsDir() {
			node.SetColor(DefaultTheme.Secondary)
			node.SetSelectedTextStyle(tcell.StyleDefault.Foreground(DefaultTheme.Primary).Background(DefaultTheme.Secondary))
		} else {
			node.SetColor(DefaultTheme.OnSurface)
			node.SetSelectedTextStyle(tcell.StyleDefault.Foreground(DefaultTheme.Primary).Background(DefaultTheme.Secondary))
		}

		target.AddChild(node)
	}
}

// showFilePicker displays the file picker modal.
func (a *App) showFilePicker(startPath string, onSelect func([]string)) {
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
		absPath = "/"
	}

	picker := newFilePicker(absPath)

	// Update footer for file picker
	a.UpdateFooter("[File Picker]", KeyDescriptions["file-picker"])

	// Restore footer on exit (helper)
	restoreFooter := func() {
		a.UpdateFooter("[Populate Environment]", KeyDescriptions["populate-form"])
	}

	// Internal helper to update node text
	updateNodeVisual := func(n *tview.TreeNode) {
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
		if picker.marked[path] {
			box = fmt.Sprintf("[%s]%s[-]", sColor, tview.Escape("[✓]"))
		}
		n.SetText(box + " " + name)
	}

	// Submit handler
	submit := func() {
		var result []string
		for p, m := range picker.marked {
			if m {
				result = append(result, p)
			}
		}
		// If nothing marked, maybe return current node?
		if len(result) == 0 {
			if node := picker.view.GetCurrentNode(); node != nil {
				if ref := node.GetReference(); ref != nil {
					result = append(result, ref.(string))
				}
			}
		}
		a.pages.RemovePage("file-picker")
		restoreFooter()
		onSelect(result)
	}

	// Start Helper to handle selection (Enter)
	picker.view.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}
		path := reference.(string)

		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			if len(node.GetChildren()) == 0 {
				picker.addNodes(node, path)
			}
			node.SetExpanded(!node.IsExpanded())
		} else {
			// File: Toggle Mark
			picker.marked[path] = !picker.marked[path]
			updateNodeVisual(node)
		}
	})

	btnSubmit := tview.NewButton("Submit").SetSelectedFunc(func() {
		submit()
	})
	btnSubmit.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Primary).Foreground(DefaultTheme.OnPrimary))
	btnSubmit.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Secondary).Foreground(DefaultTheme.Primary))

	btnCancel := tview.NewButton("Cancel").SetSelectedFunc(func() {
		a.pages.RemovePage("file-picker")
		restoreFooter()
	})
	btnCancel.SetStyle(tcell.StyleDefault.Background(DefaultTheme.Surface).Foreground(DefaultTheme.OnSurface))
	btnCancel.SetActivatedStyle(tcell.StyleDefault.Background(DefaultTheme.Error).Foreground(DefaultTheme.OnError))

	buttonBar := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(btnSubmit, 10, 0, false).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(btnCancel, 10, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	// Search Frame
	searchField := tview.NewInputField().
		SetLabel("/").
		SetFieldWidth(0).
		SetLabelColor(DefaultTheme.Secondary).
		SetFieldBackgroundColor(DefaultTheme.Background).
		SetFieldTextColor(DefaultTheme.OnSurface)

	// Layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(picker.view, 0, 1, true).
		AddItem(searchField, 0, 0, false). // Hidden initially
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(buttonBar, 1, 0, false)
	layout.SetBorder(true).
		SetBorderColor(DefaultTheme.Primary).
		SetTitle(" [::b]Select Files ").
		SetTitleColor(DefaultTheme.Secondary).
		SetBorderColor(DefaultTheme.Primary).
		SetBorderPadding(0, 0, 1, 1)

	var lastSearch string

	// Helper to find all matches
	findAllMatches := func(text string) []*tview.TreeNode {
		var matches []*tview.TreeNode
		if text == "" {
			return matches
		}
		text = strings.ToLower(text)

		var search func(n *tview.TreeNode)
		search = func(n *tview.TreeNode) {
			// Use GetReference for search since GetText has prefix now
			ref := n.GetReference()
			if ref != nil {
				name := filepath.Base(ref.(string))
				if strings.Contains(strings.ToLower(name), text) {
					if n != picker.root || len(text) > 1 {
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
		search(picker.root)
		return matches
	}

	// Search Logic
	executeSearch := func(text string) {
		if text == "" {
			return
		}
		lastSearch = text
		matches := findAllMatches(text)

		if len(matches) > 0 {
			picker.view.SetCurrentNode(matches[0])
			// Keep Search UI, but change focus
			a.tview.SetFocus(picker.view)
			a.UpdateFooter("[File Picker]", KeyDescriptions["file-picker"])
		} else {
			searchField.SetLabel("Not found: ")
			searchField.SetLabelColor(DefaultTheme.Error)
		}
	}

	searchField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			executeSearch(searchField.GetText())
		case tcell.KeyEsc:
			// Cancel search
			searchField.SetText("")
			searchField.SetLabel("/")
			searchField.SetLabelColor(DefaultTheme.Secondary)
			layout.ResizeItem(searchField, 0, 0)
			lastSearch = ""
			a.tview.SetFocus(picker.view)
			a.UpdateFooter("[File Picker]", KeyDescriptions["file-picker"])
		}
	})

	// Add input capture for navigation and selection
	picker.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			// Open Search
			searchField.SetLabel("/")
			searchField.SetLabelColor(DefaultTheme.Secondary)
			layout.ResizeItem(searchField, 1, 0)
			a.tview.SetFocus(searchField)
			a.UpdateFooter("[Search]", []string{"enter: find", "esc: cancel"})
			return nil
		}

		// Search Navigation
		if lastSearch != "" && (event.Rune() == 'n' || event.Rune() == 'N') {
			matches := findAllMatches(lastSearch)
			if len(matches) == 0 {
				return nil
			}
			currentNode := picker.view.GetCurrentNode()
			currentIndex := -1
			for i, m := range matches {
				if m == currentNode {
					currentIndex = i
					break
				}
			}

			var nextIndex int
			if event.Rune() == 'N' { // Previous
				nextIndex = currentIndex - 1
				if nextIndex < 0 {
					nextIndex = len(matches) - 1
				}
			} else { // Next
				nextIndex = currentIndex + 1
				if nextIndex >= len(matches) {
					nextIndex = 0
				}
			}

			// If current node wasn't in matches (user moved), default to first match
			if currentIndex == -1 {
				nextIndex = 0
			}

			picker.view.SetCurrentNode(matches[nextIndex])
			return nil
		}

		switch {
		case event.Key() == tcell.KeyEsc:
			if lastSearch != "" || searchField.GetText() != "" {
				// Clear search first
				searchField.SetText("")
				searchField.SetLabel("/")
				searchField.SetLabelColor(DefaultTheme.Secondary)
				layout.ResizeItem(searchField, 0, 0)
				lastSearch = ""
				return nil
			}
			a.pages.RemovePage("file-picker")
			restoreFooter()
			return nil

		case event.Key() == tcell.KeyEnter:
			submit()
			return nil

		case event.Key() == tcell.KeyRight:
			node := picker.view.GetCurrentNode()
			if node != nil {
				ref := node.GetReference()
				if ref == nil {
					return nil
				}
				path := ref.(string)
				info, err := os.Stat(path)
				if err == nil && info.IsDir() {
					if len(node.GetChildren()) == 0 {
						picker.addNodes(node, path)
					}
					node.SetExpanded(true)
				}
			}
			return nil

		case event.Key() == tcell.KeyLeft:
			node := picker.view.GetCurrentNode()
			if node != nil {
				if node.IsExpanded() {
					node.SetExpanded(false)
				} else {
					// Navigate to parent
					// Since TreeNode doesn't expose GetParent, we search from root.
					if node == picker.root {
						return nil // Cannot go up from root
					}
					var parent *tview.TreeNode
					var findParent func(*tview.TreeNode) bool
					findParent = func(n *tview.TreeNode) bool {
						for _, child := range n.GetChildren() {
							if child == node {
								parent = n
								return true
							}
							if child.IsExpanded() { // Only search expanded nodes as parents
								if findParent(child) {
									return true
								}
							}
						}
						return false
					}
					findParent(picker.root)
					if parent != nil {
						picker.view.SetCurrentNode(parent)
					}
				}
			}
			return nil

		case event.Key() == tcell.KeyRune && event.Rune() == ' ':
			node := picker.view.GetCurrentNode()
			if node != nil {
				ref := node.GetReference()
				if ref != nil {
					// Toggle Mark
					path := ref.(string)
					picker.marked[path] = !picker.marked[path]
					updateNodeVisual(node)
				}
			}
			return nil
		}

		return event
	})

	a.pages.AddPage("file-picker", CenterPrimitive(layout, 2, 4), true, true)
	a.tview.SetFocus(picker.view)
}
