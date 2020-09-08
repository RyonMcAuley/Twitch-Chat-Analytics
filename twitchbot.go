package twitchbot

//TwitchBot interface for accessing chat
type TwitchBot interface {
	//Connects to the twitch chat server
	Connect()
	//Disconnects from the twitch chat server
	Disconnect()
	//Joins a channel's chat in order to access it
	JoinChannel()
	//Listens to chat & maintains connection
	HandleChat() error
	//Keeps bot connected and handling chat
	Start()
}
