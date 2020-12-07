package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	tokenConfig   = "INCOMING_SLACK_TOKEN"
	webhookConfig = "INCOMING_SLACK_WEBHOOK"
	// Incoming payload form will have the following keys:
	// (See: https://api.slack.com/slash-commands)
	keyToken       = "token"
	keyTeamID      = "team_id"
	keyChannelId   = "channel_id"
	keyChannelName = "channel_name"
	keyUserID      = "user_id"
	keyUserName    = "user_name"
	keyCommand     = "command"
	keyText        = "text"
)

type slackMsg struct {
	Text     string `json:"text"`
	Username string `json:"username"` // Anonymous animal sender
	Channel  string `json:"channel"`  // Recipient
}

var (
	port int
	// Random animals cribbed from Google Drive's "Anonymous [Animal]" notifications
	animals = []string{
		"Alligator", "Anteater", "Armadillo", "Auroch", "Axolotl", "Badger", "Bat", "Beaver", "Buffalo",
		"Camel", "Chameleon", "Cheetah", "Chipmunk", "Chinchilla", "Chupacabra", "Cormorant", "Coyote",
		"Crow", "Dingo", "Dinosaur", "Dolphin", "Duck", "Elephant", "Ferret", "Fox", "Frog", "Giraffe",
		"Gopher", "Grizzly", "Hedgehog", "Hippo", "Hyena", "Jackal", "Ibex", "Ifrit", "Iguana", "Koala",
		"Kraken", "Lemur", "Leopard", "Liger", "Llama", "Manatee", "Mink", "Monkey", "Narwhal", "Nyan Cat",
		"Orangutan", "Otter", "Panda", "Penguin", "Platypus", "Python", "Pumpkin", "Quagga", "Rabbit", "Raccoon",
		"Rhino", "Sheep", "Shrew", "Skunk", "Slow Loris", "Squirrel", "Turtle", "Walrus", "Wolf", "Wolverine", "Wombat",
	}
	// Username must be first.
	payloadExp = regexp.MustCompile(`([@#][^\s]+):?(.*)`)
)

// readAnonymousMessage parses the username and re-routes
// the message to the user from an anonymous animal
func readAnonymousMessage(r *http.Request) string {
	err := r.ParseForm()
	// TODO: Change HTTP status code
	if err != nil {
		return string(err.Error())
	}
	// Incoming POST's token should match the one set in Heroku
	if len(r.Form[keyToken]) == 0 || r.Form[keyToken][0] != os.Getenv(tokenConfig) {
		return "Config error."
	}
	if len(r.Form[keyText]) == 0 {
		return "Slack bug; inform the team."
	}
	msg := strings.TrimSpace(r.Form[keyText][0])
	matches := payloadExp.FindStringSubmatch(msg)
	if matches == nil {
		return "Failed; message should be like: /anon @ashwin hey what's up?"
	}
	user := matches[1]
	msg = strings.TrimSpace(matches[2])
	err = sendAnonymousMessage(user, msg)
	if err != nil {
		return "Failed to send message."
	}
	return fmt.Sprintf("Anonymously sent [%s] to %s", msg, user)
}

// sendAnonymousMessage uses an incoming hook to Direct Message
// the given user the message, from a random animal.
func sendAnonymousMessage(username, message string) error {
	url := os.Getenv(webhookConfig)
	payload, err := json.Marshal(slackMsg{
		Text:     message,
		Channel:  username,
		Username: fmt.Sprintf("Santa %s", animals[rand.Intn(len(animals))]),
	})
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(payload))
	return err
}

func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		result := readAnonymousMessage(r)
		fmt.Fprintf(w, result)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func init() {
	flag.IntVar(&port, "port", 5000, "HTTP server port")
	flag.Parse()
}
