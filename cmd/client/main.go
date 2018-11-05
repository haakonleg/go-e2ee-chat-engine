package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell"

	"github.com/rivo/tview"
	"golang.org/x/net/websocket"
)

const (
	privKeyFile           = "privatekey.pem"
	chatRoomsPollInterval = 5
)

type Client struct {
	sock    *websocket.Conn
	authKey []byte

	app      *tview.Application
	pages    *tview.Pages
	loginGUI *LoginGUI
	roomsGUI *RoomsGUI
}

// Connect connects to the websocket server
func (c *Client) Connect(server string) bool {
	if c.sock == nil {
		ws, err := websocket.Dial(server, "", "http://")
		if err != nil {
			c.ShowDialog("Error connecting to server")
			return false
		}
		c.sock = ws
	}
	return true
}

// ShowDialog shows a message dialog to the user
func (c *Client) ShowDialog(message string) {
	modal := tview.NewModal()
	modal.SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			c.pages.RemovePage("error")
		}).
		SetBackgroundColor(tcell.ColorDarkRed)

	c.pages.AddPage("error", modal, true, true)
	c.app.SetFocus(modal)
}

// ShowChatRoomGUI switches to the chat rooms interface
func (c *Client) ShowChatRoomGUI() {
	c.pages.SwitchToPage("rooms")
	c.app.SetInputCapture(c.roomsGUI.KeyHandler)

	// Start updater for the chat room list
	c.roomsGUI.ChatRoomsUpdater = time.NewTicker(chatRoomsPollInterval * time.Second)
	go c.roomsGUI.updateChatRooms(c)
}

func main() {
	c := &Client{}

	// Initialize GUI
	c.app = tview.NewApplication()

	c.loginGUI = &LoginGUI{
		DefaultServerText: "ws://localhost:5000",
		CreateUserHandler: c.createUserHandler,
		LoginUserHandler:  c.loginUserHandler}
	c.loginGUI.Create(c.app)

	c.roomsGUI = &RoomsGUI{
		CreateNewRoomHandler: c.createNewChatRoomHandler}
	c.roomsGUI.Create()

	c.pages = tview.NewPages().
		AddPage("login", c.loginGUI.layout, true, true).
		AddPage("rooms", c.roomsGUI.layout, true, false)

	c.app.SetRoot(c.pages, true).
		SetFocus(c.pages).
		SetInputCapture(c.loginGUI.KeyHandler)

	// Enter GUI event loop
	if err := c.app.Run(); err != nil {
		log.Fatal(err)
	}
}
