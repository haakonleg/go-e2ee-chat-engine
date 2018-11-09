package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type GUIConfig struct {
	DefaultServerText     string
	ChatRoomsPollInterval int
	CreateUserHandler     func(server string, username string)
	LoginUserHandler      func(server string, username string)
	CreateRoomHandler     func(name string)
	JoinChatHandler       func(name string)
}

type GUI struct {
	chatRoomsPollInterval int
	app                   *tview.Application
	pages                 *tview.Pages
	loginGUI              *LoginGUI
	roomsGUI              *RoomsGUI
	chatGUI               *ChatGUI
}

func NewGUI(config *GUIConfig) *GUI {
	g := &GUI{
		chatRoomsPollInterval: config.ChatRoomsPollInterval,
		app:                   tview.NewApplication()}

	g.loginGUI = &LoginGUI{
		gui:               g,
		DefaultServerText: config.DefaultServerText,
		CreateUserHandler: config.CreateUserHandler,
		LoginUserHandler:  config.LoginUserHandler}
	g.loginGUI.Create()

	g.roomsGUI = &RoomsGUI{
		gui:               g,
		CreateRoomHandler: config.CreateRoomHandler,
		JoinChatHandler:   config.JoinChatHandler}
	g.roomsGUI.Create()

	g.chatGUI = &ChatGUI{
		gui: g}
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
func (g *GUI) ShowDialog(message string) {
	modal := tview.NewModal()
	modal.SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.pages.RemovePage("error")
		}).
		SetBackgroundColor(tcell.ColorDarkRed)

	g.pages.AddPage("error", modal, true, true)
	g.app.SetFocus(modal)
}

// ShowChatRoomGUI switches to the chat rooms interface
func (g *GUI) ShowChatRoomGUI(client *Client) {
	g.pages.SwitchToPage("rooms")
	g.app.SetInputCapture(g.roomsGUI.KeyHandler)

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
		OnChatInfo:     g.chatGUI.onChatInfo,
		OnChatMessage:  g.chatGUI.onChatMessage,
		OnUserJoined:   g.chatGUI.onUserJoined,
		OnUserLeft:     g.chatGUI.onUserLeft,
		Socket:         client.sock,
		PrivateKey:     client.privateKey,
		AuthKey:        client.authKey}

	// Set handlers for chat gui
	g.chatGUI.SendChatMessageHandler = client.chatSession.SendChatMessage
	g.chatGUI.LeaveChatHandler = client.chatSession.LeaveChat

	go client.chatSession.StartChatSession()
}
