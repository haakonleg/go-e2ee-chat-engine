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
	SendChatMessageHandler func(message string)

	gui      *GUI
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

	gui.layout = tview.NewGrid()
	gui.layout.SetRows(0, 3).
		SetColumns(0, 30).
		AddItem(gui.msgView, 0, 0, 1, 1, 0, 0, false).
		AddItem(gui.userList, 0, 1, 2, 1, 0, 0, false)

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

	gui.layout.AddItem(gui.msgInput, 1, 0, 1, 1, 0, 0, true)
	gui.gui.app.SetFocus(gui.layout)
}

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
	for _, user := range cs.Users {
		gui.userList.Write([]byte(user.Username + "\n"))
	}
}

func (gui *ChatGUI) onChatInfo(err error, cs *ChatSession, chatInfo *websock.ChatInfoMessage) {
	gui.gui.app.QueueUpdate(func() {
		if err != nil {
			gui.gui.ShowDialog(err.Error())
			gui.gui.app.Draw()
			return
		}

		gui.WriteUserList(cs)

		for _, msg := range chatInfo.Messages {
			fmtMsg := formatChatMessage(msg.Sender, msg.Message, msg.Timestamp)
			gui.msgView.Write(fmtMsg)
		}
		gui.gui.app.Draw()
	})
}

func (gui *ChatGUI) onChatMessage(err error, cs *ChatSession, chatMessage *websock.ChatMessage) {
	gui.gui.app.QueueUpdate(func() {
		if err != nil {
			gui.gui.ShowDialog(err.Error())
			gui.gui.app.Draw()
			return
		}

		fmtMsg := formatChatMessage(chatMessage.Sender, chatMessage.Message, chatMessage.Timestamp)
		gui.msgView.Write(fmtMsg)
		gui.gui.app.Draw()
	})
}
