package main

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type LoginGUI struct {
	app               *tview.Application
	layout            *tview.Grid
	statusText        *tview.TextView
	focusableElements []tview.Primitive
	focusedIndex      int
}

func (gui *LoginGUI) Show() {
	if err := gui.app.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewLoginGUI() *LoginGUI {
	gui := new(LoginGUI)

	serverInput := tview.NewInputField().
		SetLabel("Server   ").
		SetFieldWidth(60)

	usernameInput := tview.NewInputField().
		SetLabel("Username ").
		SetFieldWidth(60)

	createBtn := tview.NewButton("Create User")
	loginBtn := tview.NewButton("Log In")

	gui.statusText = tview.NewTextView().
		SetTextColor(tcell.ColorLightBlue).
		SetTextAlign(tview.AlignCenter)
	gui.statusText.SetText("Welcome. Create a new user or log in using your private key file.")

	gui.layout = tview.NewGrid()
	gui.layout.SetRows(0, 1, 1, 1, 1, 0, 2).
		SetColumns(0, 30, 5, 30, 0).
		AddItem(serverInput, 1, 1, 1, 3, 0, 0, true).
		AddItem(usernameInput, 2, 1, 1, 3, 0, 0, false).
		AddItem(createBtn, 4, 1, 1, 1, 0, 0, false).
		AddItem(loginBtn, 4, 3, 1, 1, 0, 0, false).
		AddItem(gui.statusText, 6, 0, 1, 5, 0, 0, false).
		SetBorder(true).
		SetTitle("Chat Client")

	gui.focusableElements = []tview.Primitive{
		serverInput, usernameInput,
		createBtn, loginBtn}

	gui.app = tview.NewApplication().
		SetRoot(gui.layout, true).
		SetFocus(gui.layout).
		SetInputCapture(gui.setNextFocus)

	return gui
}

func (gui *LoginGUI) setNextFocus(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyTab {
		gui.app.SetFocus(gui.focusableElements[gui.focusedIndex])

		gui.focusedIndex++
		if gui.focusedIndex == len(gui.focusableElements) {
			gui.focusedIndex = 0
		}
		return ev
	}
	return ev
}
