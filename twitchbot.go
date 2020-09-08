package twitchbot

import (
	"net"
	"time"
)

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

//OAuthCred credentials
type OAuthCred struct {
	Password string `json:"password,omitempty"`
}

//BasicBot struct object that does the interacting
// with the chat
type BasicBot struct {
	//Name of the channel to join
	Channel string

	//Path to private json auth token file
	PrivatePath string

	//Reference to bot's network connection
	conn net.Conn

	// The credentials necessary for authentication.
	Credentials *OAuthCred

	//forced delay between messages to avoid breaking twitch guidelines
	MsgRate time.Duration

	//Name for the bot to use in chat
	Name string

	//time that the bot is starting
	// used for logging
	startTime time.Time

	//server domain of twitch chat server
	Server string

	// port for twitch chat server
	Port string
}
