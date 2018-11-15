package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"github.com/rivo/tview"
)

type ChatGUI struct {
	*GUI
	SendChatMessageHandler func(message string)
	LeaveChatHandler       func()

	layout   *tview.Grid
	userList *tview.TextView
	msgView  *tview.TextView
	msgInput *tview.InputField
}

// Create initializes the widgets in the chat GUI
func (gui *ChatGUI) Create() {
	gui.userList = tview.NewTextView()
	gui.userList.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Users")

	gui.msgView = tview.NewTextView()
	gui.msgView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Chat")

	sendBtn := tview.NewButton("(Enter) Send")
	exitBtn := tview.NewButton("(Esc) Leave")

	gui.layout = tview.NewGrid()
	gui.layout.SetRows(0, 3, 1).
		SetColumns(20, 1, 20, 0, 30).
		AddItem(gui.msgView, 0, 0, 1, 4, 0, 0, false).
		AddItem(gui.userList, 0, 4, 2, 1, 0, 0, false).
		AddItem(sendBtn, 2, 0, 1, 1, 0, 0, false).
		AddItem(exitBtn, 2, 2, 1, 1, 0, 0, false)

	gui.AddMsgInput()
}

// AddMsgInput adds the input field for typing in a chat message to the layout, this is needed
// because to clear an InputField in tview, we have to create a new InputField, so this code needs to run often
func (gui *ChatGUI) AddMsgInput() {
	gui.msgInput = tview.NewInputField()
	gui.msgInput.SetDoneFunc(gui.MsgInputHandler).
		SetBorder(true).
		SetTitle("Message").
		SetTitleAlign(tview.AlignLeft)

	gui.layout.AddItem(gui.msgInput, 1, 0, 1, 4, 0, 0, true)
	gui.app.SetFocus(gui.layout)
}

// FormatChatMessage formats a chat message to human readable format
func formatChatMessage(sender string, message []byte, timestamp int64) []byte {
	var buf bytes.Buffer

	tm := time.Unix(timestamp/1000, 0)
	buf.WriteString(fmt.Sprintf("[dimgray]%02d-%02d %02d:%02d[white]", tm.Day(), tm.Month(), tm.Hour(), tm.Minute()))
	buf.WriteString(" [blue]<")
	buf.WriteString(string(sender))
	buf.WriteString("> [white]")
	buf.WriteString(string(message))
	buf.WriteRune('\n')

	return buf.Bytes()
}

// MsgInputHandler is the key handler for the chat message input field
func (gui *ChatGUI) MsgInputHandler(key tcell.Key) {
	if key == tcell.KeyEnter {
		gui.SendChatMessageHandler(gui.msgInput.GetText())
		gui.layout.RemoveItem(gui.msgInput)
		gui.AddMsgInput()
	}
}

// WriteUserList adds the currently connected users to the list of users
func (gui *ChatGUI) WriteUserList(cs *ChatSession) {
	gui.userList.Clear()
	for _, user := range cs.users {
		gui.userList.Write([]byte(user.Username + "\n"))
	}
}

// OnChatInfo is called whenver a ChatInfo message is received from the server. It is responsible for
// displaying all chat messages and users from the chat room in the interface
func (gui *ChatGUI) OnChatInfo(err error, cs *ChatSession, chatInfo *websock.ChatInfoMessage) {
	gui.app.QueueUpdate(func() {
		if err != nil {
			gui.ShowDialog(err.Error(), nil)
			gui.app.Draw()
			return
		}

		gui.msgView.Clear()
		gui.WriteUserList(cs)

		for _, msg := range chatInfo.Messages {
			fmtMsg := formatChatMessage(msg.Sender, msg.Message, msg.Timestamp)
			gui.msgView.Write(fmtMsg)
			gui.msgView.ScrollToEnd()
		}
		gui.app.Draw()
	})
}

// OnChatMessage is called whenver a chat message is received from the server. It is responsible for
// displaying the new chat message in the chat message view
func (gui *ChatGUI) OnChatMessage(err error, cs *ChatSession, chatMessage *websock.ChatMessage) {
	gui.app.QueueUpdate(func() {
		if err != nil {
			gui.ShowDialog(err.Error(), nil)
			gui.app.Draw()
			return
		}

		fmtMsg := formatChatMessage(chatMessage.Sender, chatMessage.Message, chatMessage.Timestamp)
		gui.msgView.Write(fmtMsg)
		gui.app.Draw()
	})
}

// OnUserJoined is called when the server notifies that a new user has joined. It is responsible for
// adding the new user to the displayed list of online users
func (gui *ChatGUI) OnUserJoined(err error, cs *ChatSession, user *websock.User) {
	gui.app.QueueUpdate(func() {
		if err != nil {
			gui.ShowDialog(err.Error(), nil)
			gui.app.Draw()
			return
		}

		gui.WriteUserList(cs)
		var buf bytes.Buffer
		buf.WriteString("[dimgray]")
		buf.WriteString(user.Username)
		buf.WriteString(" connected\n")
		gui.msgView.Write(buf.Bytes())
		gui.app.Draw()
	})
}

// OnUserLeft is called when the server notifies that a user has left the chat room. It is responsible for
// removing the user from the displayed list of online users
func (gui *ChatGUI) OnUserLeft(cs *ChatSession, username string) {
	gui.app.QueueUpdate(func() {
		gui.WriteUserList(cs)
		var buf bytes.Buffer
		buf.WriteString("[dimgray]")
		buf.WriteString(username)
		buf.WriteString(" disconnected\n")
		gui.msgView.Write(buf.Bytes())
		gui.app.Draw()
	})
}

// KeyHandler is the keyboard input handler for the chat rooms interface
func (gui *ChatGUI) KeyHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEsc {
		gui.LeaveChatHandler()
	}
	return key
}
