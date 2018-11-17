package main

import (
	"log"
	"strconv"
	"time"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	newRoomPopup  = "createRoomPopup"
	joinRoomPopup = "joinRoomPopup"
	passwordPopup = "passwordPopup"
	labelName     = "Name "
	labelPassword = "Password "
)

// RoomsGUI contains the widgets/state of the chat rooms view
type RoomsGUI struct {
	*GUI
	CreateRoomHandler func(name, password string, isHidden bool)
	JoinChatHandler   func(name, password string)
	ChatRoomsUpdater  *time.Ticker
	ServerAddress     string

	layout        *tview.Pages
	roomList      *tview.List
	createRoomBtn *tview.Button
	joinRoomBtn   *tview.Button
	serverStatus  *tview.TextView
	chatRooms     map[string]*websock.Room
}

// Create initializes the widgets in the chat rooms GUI
func (gui *RoomsGUI) Create() {
	gui.chatRooms = make(map[string]*websock.Room, 0)

	gui.roomList = tview.NewList()
	gui.roomList.
		SetSelectedFunc(gui.onChatRoomSelected).
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

// Called when a chat room is selected in the list
func (gui *RoomsGUI) onChatRoomSelected(index int, name, secText string, scut rune) {
	chatRoom := gui.chatRooms[name]
	log.Println(chatRoom)

	if chatRoom.HasPassword {
		gui.passwordPopup(func(password string) {
			gui.JoinChatHandler(name, password)
		})
	} else {
		gui.JoinChatHandler(name, "")
	}
}

// validateForm validates the chat room name and password input by the user when
// trying to create a new chat room or joining a chat room
func (gui *RoomsGUI) validateForm(form *tview.Form) (string, string, bool) {
	name := form.GetFormItemByLabel(labelName).(*tview.InputField).GetText()
	password := form.GetFormItemByLabel(labelPassword).(*tview.InputField).GetText()

	// Validate username and password
	if len(name) < 3 {
		gui.ShowDialog("Name must be 3 characters or longer", nil)
		return "", "", false
	}
	if len(password) != 0 && len(password) < 6 {
		gui.ShowDialog("Password must be 6 characters or longer", nil)
		return "", "", false
	}

	return name, password, true
}

// onNewRoom is called when the user presses the "Create" button in the new room form
func (gui *RoomsGUI) onNewRoom(roomForm *tview.Form) {
	isHidden := roomForm.GetFormItem(1).(*tview.Checkbox).IsChecked()
	name, password, ok := gui.validateForm(roomForm)

	if ok {
		gui.layout.RemovePage(newRoomPopup)
		gui.CreateRoomHandler(name, password, isHidden)
	}
}

// newRoomPopup creates a popup window containing a form for creating a new chat room
// if the user presses the "Create" button, onNewRoom is called
func (gui *RoomsGUI) newRoomPopup() {
	form := tview.NewForm()
	form.AddInputField(labelName, "", 30, nil, nil).
		AddCheckbox("Hidden", false, nil).
		AddInputField(labelPassword, "", 60, nil, nil).
		AddButton("Create", func() { gui.onNewRoom(form) }).
		SetCancelFunc(func() { gui.layout.RemovePage(newRoomPopup) })
	form.GetFormItem(2).(*tview.InputField).
		SetPlaceholder("Blank for no password").
		SetPlaceholderTextColor(tcell.ColorWhite).
		SetMaskCharacter('*')

	box := tview.NewBox().SetBorder(true).SetTitle("New Room")
	popup := tview.NewGrid().
		SetRows(0, 1, 9, 1, 0).
		SetColumns(0, 1, 40, 1, 0).
		AddItem(box, 1, 1, 3, 3, 0, 0, false).
		AddItem(form, 2, 2, 1, 1, 0, 0, true)

	gui.layout.AddPage(newRoomPopup, popup, true, true)
}

// onJoinRoom is called when the user presses "Join" in the join room form
func (gui *RoomsGUI) onJoinRoom(joinForm *tview.Form) {
	name, password, ok := gui.validateForm(joinForm)
	log.Println(name)
	log.Println(password)

	if ok {
		gui.layout.RemovePage(joinRoomPopup)
		gui.JoinChatHandler(name, password)
	}
}

// joinRoomPopup creates a popup window containing a form for joining a chat room
// if the user presses the "Join" button, onJoinRoom is called
func (gui *RoomsGUI) joinRoomPopup() {
	form := tview.NewForm()
	form.AddInputField(labelName, "", 30, nil, nil).
		AddInputField(labelPassword, "", 60, nil, nil).
		AddButton("Join", func() { gui.onJoinRoom(form) }).
		SetCancelFunc(func() { gui.layout.RemovePage(joinRoomPopup) })
	form.GetFormItem(1).(*tview.InputField).
		SetPlaceholder("Blank for no password").
		SetPlaceholderTextColor(tcell.ColorWhite).
		SetMaskCharacter('*')

	box := tview.NewBox().SetBorder(true).SetTitle("Join Room")
	popup := tview.NewGrid().
		SetRows(0, 1, 7, 1, 0).
		SetColumns(0, 1, 40, 1, 0).
		AddItem(box, 1, 1, 3, 3, 0, 0, false).
		AddItem(form, 2, 2, 1, 1, 0, 0, true)

	gui.layout.AddPage(joinRoomPopup, popup, true, true)
}

// passwordPopup creates a popup window containing an input field for
// the user to enter a chat room password
func (gui *RoomsGUI) passwordPopup(onDone func(password string)) {
	pwdInput := tview.NewInputField()

	handler := func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			password := pwdInput.GetText()
			if len(password) < 6 {
				gui.ShowDialog("Password must be 6 characters or longer", nil)
				return
			}
			onDone(password)
			fallthrough
		case tcell.KeyEsc:
			gui.layout.RemovePage(passwordPopup)
		}
	}

	pwdInput.SetFieldWidth(60).
		SetDoneFunc(handler).
		SetMaskCharacter('*')

	box := tview.NewBox().SetBorder(true).SetTitle("Password")
	popup := tview.NewGrid().
		SetRows(0, 1, 1, 1, 0).
		SetColumns(0, 1, 40, 1, 0).
		AddItem(box, 1, 1, 3, 3, 0, 0, false).
		AddItem(pwdInput, 2, 2, 1, 1, 0, 0, true)

	gui.layout.AddPage(passwordPopup, popup, true, true)
}

// addChatRoom adds the given chat room to the map of chat rooms, and adds it
// to the list if it is not already added
func (gui *RoomsGUI) addChatRoom(room *websock.Room) {
	// Add the chat room if it is not in the list
	if _, hasRoom := gui.chatRooms[room.Name]; !hasRoom {
		gui.chatRooms[room.Name] = room
		gui.roomList.AddItem(room.Name,
			"[Online users: "+strconv.Itoa(room.OnlineUsers)+"] [Password: "+strconv.FormatBool(room.HasPassword)+"]",
			0, nil)
	}
}

// This function runs in a separate goroutine and updates the chat rooms list on a regular interval
func (gui *RoomsGUI) updateChatRooms(client *Client) {
	update := func() {
		chatRooms, err := client.getChatRooms()
		log.Println(chatRooms)

		gui.app.QueueUpdate(func() {
			if err != nil {
				gui.ShowDialog(err.Error(), nil)
				gui.app.Draw()
				return
			}

			// Update status
			gui.serverStatus.SetText("Connected to: " + gui.ServerAddress + "\tConnected Users: " + strconv.Itoa(chatRooms.TotalConnected))

			// Add every chat room to the list
			for i := range chatRooms.Rooms {
				gui.addChatRoom(&chatRooms.Rooms[i])
			}
			gui.app.Draw()
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
	// Check if there is a popup window currently displayed
	hasPopup := func() bool {
		return gui.layout.HasPage(newRoomPopup) ||
			gui.layout.HasPage(joinRoomPopup) ||
			gui.layout.HasPage(passwordPopup)
	}

	if !hasPopup() {
		switch ev.Rune() {
		case 'c':
			gui.newRoomPopup()
		case 'j':
			gui.joinRoomPopup()
		}
	}
	return ev
}
