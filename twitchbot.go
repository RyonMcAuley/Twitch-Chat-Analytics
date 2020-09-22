package twitchbot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

//ESTFormat is used to define timestamp format
const ESTFormat = "Jan 2 15:05:05 EST"

//Regular expression for parsing PRIVMSG strings
//Parses for username, message type, and the user's message
var msgRegex *regexp.Regexp = regexp.MustCompile(`^:(\w+)!\w+@\w+\.tmi\.twitch\.tv (PRIVMSG) #\w+(?: :(.*))?$`)

//TwitchBot interface for accessing chat
type TwitchBot interface {
	//Connects to the twitch chat server
	Connect()
	//Disconnects from the twitch chat server
	Disconnect()
	//Listens to chat & maintains connection
	HandleChat() error
	//Joins a channel's chat in order to access it
	JoinChannel()
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

//Connect This function will connect the bot to the Twitch Chat server
func (bb *BasicBot) Connect() {
	var err error
	fmt.Printf("[%s] Connecting to %s...\n", timeStamp(), bb.Server)
	//makes connection to server
	bb.conn, err = net.Dial("tcp", bb.Server+":"+bb.Port)
	if nil != err {
		fmt.Printf("[%s] Cannot connect to %s, retrying.\n", timeStamp(), bb.Server)
	}
}

//Disconnect cleanly disconnects the bot from the chat Server
func (bb *BasicBot) Disconnect() {
	bb.conn.Close()
	fmt.Printf("[%s] Closed connection from %s\n", timeStamp(), bb.Server)
}

//HandleChat this function does the heavy lifting of handling chat messages
func (bb *BasicBot) HandleChat() error {
	fmt.Printf("[%s] Watching #%s...\n", timeStamp(), bb.Channel)
	tp := textproto.NewReader(bufio.NewReader(bb.conn))

	for {
		line, err := tp.ReadLine()
		if nil != err {
			bb.Disconnect()
			return errors.New("bb.Bot.HandleChat: Failed to read line from channel. Disconnected")
		}
		fmt.Printf("[%s] %s\n", timeStamp(), line)
		if "PING :tmi.twitch.tv" == line {
			//maintains connection by replying to PING message from server
			bb.conn.Write([]byte("PONG: tmi.twitch.tv\r\n"))
			continue
		} else {
			//parse lines
			matches := msgRegex.FindStringSubmatch(line)
			if nil != matches {
				userName := matches[1]
				msgType := matches[2]

				//switch statement for possible future additions
				switch msgType {
				case "PRIVMSG":
					msg := matches[3]
					fmt.Printf("[%s] %s: %s\n", timeStamp(), userName, msg)

					if userName == bb.Channel {
						switch msg {
						case "!tbdown":
							fmt.Printf("[%s] Shutdown command received. Shutting down now...\n", timeStamp())
							bb.Say("goodbye")
							bb.Disconnect()
							return nil
						default:
							//do nothing
						}
					}
				}
			}
		}
		//Sleep to follow Twitch message rate restrictions
		time.Sleep(bb.MsgRate)
	}
}

//JoinChannel connects the bot to the specified channel
func (bb *BasicBot) JoinChannel() {
	fmt.Printf("[%s] Joining #%s...\n", timeStamp(), bb.Channel)
	bb.conn.Write([]byte("PASS " + bb.Credentials.Password + "\r\n"))
	bb.conn.Write([]byte("NICK " + bb.Name + "\r\n"))
	bb.conn.Write([]byte("JOIN #" + bb.Channel + "\r\n"))

	fmt.Printf("[%s] Joined #%s as @%s!\n", timeStamp(), bb.Channel, bb.Name)
}

//ReadCredentials accesses the json auth token and establishes credentials
func (bb *BasicBot) ReadCredentials() error {
	credFile, err := ioutil.ReadFile(bb.PrivatePath)
	if nil != err {
		return err
	}
	bb.Credentials = &OAuthCred{}

	//parse file contents
	dec := json.NewDecoder(strings.NewReader(string(credFile)))
	if err = dec.Decode(bb.Credentials); nil != err && io.EOF != err {
		return err
	}

	return nil
}

//Say sends a message to the channel from the bot's account
func (bb *BasicBot) Say(msg string) error {
	if "" == msg {
		return errors.New("BasicBot.Say: msg was empty")
	}
	_, err := bb.conn.Write([]byte(fmt.Sprintf("PRIVMSG #%s %s\r\n", bb.Channel, msg)))
	/* Should be irrelevant
	if nil != err {
		return err
	}
	return nil
	*/
	return err
}

//Start loops calling HandleChat, attempts to reconnct if connection drops.
//Attempts to reconnect until shutdown.
func (bb *BasicBot) Start() {
	err := bb.ReadCredentials()
	if nil != err {
		fmt.Println(err)
		fmt.Println("Aborting...")
		return
	}

	for {
		bb.Connect()
		bb.JoinChannel()
		err = bb.HandleChat()
		if nil != err {
			//attempt reconnect
			time.Sleep(1000 * time.Millisecond)
			fmt.Println(err)
			fmt.Println("Starting again...")
		} else {
			return
		}
	}
}

//timeStamp is used to return a timestamp in the correct format
func timeStamp() string {
	return TimeStamp(ESTFormat)
}

//TimeStamp calls time function to format time string correctly
func TimeStamp(format string) string {
	return time.Now().Format(format)
}
