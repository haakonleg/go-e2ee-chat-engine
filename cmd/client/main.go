package main

func login() {

}

func main() {
	NewLoginGUI().Show()

	/*
		ws, err := websock.NewClient(os.Args[1])
		if err != nil {
			log.Fatal(err)

		msg := &websock.Message{"Hello"}
		if err := websocket.JSON.Send(ws, msg); err != nil {
			log.Fatal(err)
		}

		res := new(websock.Message)
		if err := websocket.JSON.Receive(ws, res); err != nil {
			log.Fatal(err)
		}

		log.Println("Recieved message from server: %v", res)*/
}
