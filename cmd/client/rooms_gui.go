package main

import (
	"strconv"
	"time"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type RoomsGUI struct {
	CreateRoomHandler func(name string)
	JoinChatHandler   func(name string)
	ChatRoomsUpdater  *time.Ticker
	ServerAddress     string

	gui           *GUI
	layout        *tview.Pages
	roomList      *tview.List
	createRoomBtn *tview.Button
	joinRoomBtn   *tview.Button
	serverStatus  *tview.TextView
}

// Create initializes the widgets in the chat rooms GUI
func (gui *RoomsGUI) Create() {
	gui.roomList = tview.NewList()
	gui.roomList.SetSelectedFunc(func(index int, text, secText string, scut rune) {
		gui.JoinChatHandler(text)
	}).
		SetBorder(true).
		SetTitle("Chat Rooms").
		SetTitleAlign(tview.AlignLeft)

	gui.createRoomBtn = tview.NewButton("Create Room (C)")
	gui.joinRoomBtn = tview.NewButton("Join Room (J)")

	gui.serverStatus = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Connected to: ws://blahblah:1234\tConnected Users: 9000")

	grid := tview.NewGrid()
	grid.SetRows(1, 0, 1).
		SetColumns(20, 2, 20, 0).
		AddItem(gui.serverStatus, 0, 0, 1, 4, 0, 0, false).
		AddItem(gui.roomList, 1, 0, 1, 4, 0, 0, true).
		AddItem(gui.createRoomBtn, 2, 0, 1, 1, 0, 0, false).
		AddItem(gui.joinRoomBtn, 2, 2, 1, 1, 0, 0, false)

	gui.layout = tview.NewPages().
		AddPage("main", grid, true, true)
}

// Creates a new popup window with an input field and shows it, the text entered by the user
// will be passed to the callback function "doneFunc" if enter is pressed
func (gui *RoomsGUI) getInput(title, label string, doneFunc func(text string)) string {
	input := tview.NewInputField()

	handler := func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			doneFunc(input.GetText())
			gui.layout.RemovePage("popup")
		case tcell.KeyEscape:
			gui.layout.RemovePage("popup")
		}
	}

	input.SetLabel(label).
		SetDoneFunc(handler)

	box := tview.NewBox().
		SetBorder(true).
		SetTitle(title)

	popup := tview.NewGrid().
		SetRows(0, 1, 1, 1, 0).
		SetColumns(0, 1, 38, 1, 0).
		AddItem(box, 1, 1, 3, 3, 0, 0, false).
		AddItem(input, 2, 2, 1, 1, 0, 0, true)

	gui.layout.AddPage("popup", popup, true, true)
	return ""
}

func (gui *RoomsGUI) addChatRoom(room *websock.Room) {
	// Check if the room already exists in the list
	hasChatRoom := func(name string) bool {
		for i := 0; i < gui.roomList.GetItemCount(); i++ {
			if name, _ := gui.roomList.GetItemText(i); name == room.Name {
				return true
			}
		}
		return false
	}

	// Add the chat room if it is not in the list
	if !hasChatRoom(room.Name) {
		gui.roomList.AddItem(room.Name, "Online users: "+strconv.Itoa(room.OnlineUsers), 0, nil)
	}
}

// This function runs in a separate goroutine and updates the chat rooms list on a regular interval
func (gui *RoomsGUI) updateChatRooms(client *Client) {
	update := func() {
		chatRooms, err := client.getChatRooms()

		gui.gui.app.QueueUpdate(func() {
			if err != nil {
				gui.gui.ShowDialog(err.Error())
				gui.gui.app.Draw()
				return
			}

			// Update
			gui.serverStatus.SetText("Connected to: " + gui.ServerAddress + "\tConnected Users: " + strconv.Itoa(chatRooms.TotalConnected))

			// Add every chat room to the list
			for _, room := range chatRooms.Rooms {
				gui.addChatRoom(&room)
			}
			gui.gui.app.Draw()
		})
	}

	update()
	// Update the chat rooms on every timer fire
	for range gui.ChatRoomsUpdater.C {
		update()
	}
}

// KeyHandler is the keyboard input handler for the chat rooms GUI
func (gui *RoomsGUI) KeyHandler(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Rune() == 'c' {
		gui.getInput("New Room", "Name ", gui.CreateRoomHandler)
	}
	return ev
}
