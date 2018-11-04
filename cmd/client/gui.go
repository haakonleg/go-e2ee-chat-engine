package main

import (
	"log"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type UserHandlerFunc func(gui *LoginGUI, server string, username string) error

type LoginGUI struct {
	DefaultServerText string
	CreateUserHandler UserHandlerFunc
	LoginUserHandler  UserHandlerFunc

	app *tview.Application

	layout        *tview.Grid
	serverInput   *tview.InputField
	usernameInput *tview.InputField
	createBtn     *tview.Button
	loginBtn      *tview.Button
	statusText    *tview.TextView

	focusableElements []tview.Primitive
	focusedIndex      int
}

// Show starts the event loop for the GUI
func (gui *LoginGUI) Show() {
	if err := gui.app.Run(); err != nil {
		log.Fatal(err)
	}
}

// NewLoginGUI creates a new instance of the login GUI
func (gui *LoginGUI) Create() {
	gui.serverInput = tview.NewInputField().
		SetLabel("Server   ").
		SetFieldWidth(60).
		SetText(gui.DefaultServerText)

	gui.usernameInput = tview.NewInputField().
		SetLabel("Username ").
		SetFieldWidth(60)

	gui.createBtn = tview.NewButton("Create User")
	gui.loginBtn = tview.NewButton("Log In")

	gui.statusText = tview.NewTextView().
		SetTextColor(tcell.ColorLightBlue).
		SetTextAlign(tview.AlignCenter)
	gui.statusText.SetText("Welcome. Create a new user or log in using your private key file.")

	gui.layout = tview.NewGrid()
	gui.layout.SetRows(0, 1, 1, 1, 1, 0, 2).
		SetColumns(0, 30, 5, 30, 0).
		AddItem(gui.serverInput, 1, 1, 1, 3, 0, 0, false).
		AddItem(gui.usernameInput, 2, 1, 1, 3, 0, 0, true).
		AddItem(gui.createBtn, 4, 1, 1, 1, 0, 0, false).
		AddItem(gui.loginBtn, 4, 3, 1, 1, 0, 0, false).
		AddItem(gui.statusText, 6, 0, 1, 5, 0, 0, false).
		SetBorder(true).
		SetTitle("Chat Client")

	gui.focusableElements = []tview.Primitive{
		gui.serverInput, gui.usernameInput,
		gui.createBtn, gui.loginBtn}
	gui.focusedIndex = 1

	gui.app = tview.NewApplication().
		SetRoot(gui.layout, true).
		SetFocus(gui.layout).
		SetInputCapture(gui.keyHandler)
}

func (gui *LoginGUI) keyHandler(ev *tcell.EventKey) *tcell.EventKey {
	// Change focus to next element if tab was pressed
	if ev.Key() == tcell.KeyTab {
		gui.focusedIndex++
		if gui.focusedIndex == len(gui.focusableElements) {
			gui.focusedIndex = 0
		}

		gui.app.SetFocus(gui.focusableElements[gui.focusedIndex])

	} else if ev.Key() == tcell.KeyEnter {
		var handler UserHandlerFunc

		// Check if a button was pressed, and call its handler
		switch gui.app.GetFocus() {
		case gui.createBtn:
			handler = gui.CreateUserHandler
		case gui.loginBtn:
			handler = gui.LoginUserHandler
		}

		// If the handler returned an error display it in the status text
		if err := handler(gui, gui.serverInput.GetText(), gui.usernameInput.GetText()); err != nil {
			gui.statusText.SetText("Error: " + err.Error())
		}
	}
	return ev
}
