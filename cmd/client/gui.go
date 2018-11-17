package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// GUIConfig contains configuration parameters for the GUI.
// The functions defined here are callbacks, which will be called when some UI action happens.
type GUIConfig struct {
	DefaultServerText     string
	ChatRoomsPollInterval int
	CreateUserHandler     func(server string, username string)
	LoginUserHandler      func(server string, username string)
	CreateRoomHandler     func(name, password string, isHidden bool)
	JoinChatHandler       func(name, password string)
}

// GUI contains the widgets/state of the user interface
type GUI struct {
	chatRoomsPollInterval int
	app                   *tview.Application
	pages                 *tview.Pages
	loginGUI              *LoginGUI
	roomsGUI              *RoomsGUI
	chatGUI               *ChatGUI
}

// NewGUI creates a new instance of the GUI using a GUIConfig object
func NewGUI(config *GUIConfig) *GUI {
	g := &GUI{
		chatRoomsPollInterval: config.ChatRoomsPollInterval,
		app:                   tview.NewApplication()}

	g.loginGUI = &LoginGUI{
		GUI:               g,
		DefaultServerText: config.DefaultServerText,
		CreateUserHandler: config.CreateUserHandler,
		LoginUserHandler:  config.LoginUserHandler}
	g.loginGUI.Create()

	g.roomsGUI = &RoomsGUI{
		GUI:               g,
		CreateRoomHandler: config.CreateRoomHandler,
		JoinChatHandler:   config.JoinChatHandler}
	g.roomsGUI.Create()

	g.chatGUI = &ChatGUI{GUI: g}
	g.chatGUI.Create()

	g.pages = tview.NewPages().
		AddPage("login", g.loginGUI.layout, true, true).
		AddPage("rooms", g.roomsGUI.layout, true, false).
		AddPage("chat", g.chatGUI.layout, true, false)

	g.app.SetRoot(g.pages, true).
		SetFocus(g.pages).
		SetInputCapture(g.loginGUI.KeyHandler)

	return g
}

// ShowDialog shows a message dialog to the user
func (g *GUI) ShowDialog(message string, onDismiss func()) {
	modal := tview.NewModal()
	modal.SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.pages.RemovePage("error")
		}).
		SetBackgroundColor(tcell.ColorDarkRed)

	if onDismiss != nil {
		modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Ok" {
				onDismiss()
			}
		})
	}

	g.pages.AddPage("error", modal, true, true)
	g.app.SetFocus(modal)
}

// ShowChatRoomGUI switches to the chat rooms interface
func (g *GUI) ShowChatRoomGUI(client *Client) {
	g.pages.SwitchToPage("rooms")
	g.app.SetInputCapture(g.roomsGUI.KeyHandler)

	g.roomsGUI.ServerAddress = g.loginGUI.serverInput.GetText()

	// Start updater for the chat room list
	g.roomsGUI.ChatRoomsUpdater = time.NewTicker(time.Duration(g.chatRoomsPollInterval) * time.Second)
	go g.roomsGUI.updateChatRooms(client)
}

// ShowChatGUI switches to the chat interface
func (g *GUI) ShowChatGUI(client *Client) {
	g.roomsGUI.ChatRoomsUpdater.Stop()

	g.pages.SwitchToPage("chat")
	g.app.SetInputCapture(g.chatGUI.KeyHandler)

	// Start chat session
	client.chatSession = &ChatSession{
		DisconnectFunc: func() { g.ShowChatRoomGUI(client) },
		OnChatInfo:     g.chatGUI.OnChatInfo,
		OnChatMessage:  g.chatGUI.OnChatMessage,
		OnUserJoined:   g.chatGUI.OnUserJoined,
		OnUserLeft:     g.chatGUI.OnUserLeft,
		Reader:         client.wsReader,
		Socket:         client.ws,
		PrivateKey:     client.privateKey,
		AuthKey:        client.authKey}

	// Set handlers for chat gui
	g.chatGUI.SendChatMessageHandler = client.chatSession.SendChatMessage
	g.chatGUI.LeaveChatHandler = client.chatSession.LeaveChat

	go client.chatSession.StartChatSession()
}
