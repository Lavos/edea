package edea

import (
	"bytes"
	oauth "github.com/araddon/goauth"
	"github.com/araddon/httpstream"
	"github.com/cznic/kv"
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	UserName, ConsumerKey, ConsumerSecret, Token, TokenSecret string
}

type Application struct {
	stream chan []byte
	// expire chan []byte
	done chan bool

	tweetDB *kv.DB
	flagDB	*kv.DB

	oc	*oauth.OAuthConsumer
	at	*oauth.AccessToken

	server	*Server
}


func (a *Application) Run(){
	go awaitQuitKey(a.done)
	log.Print("Edea application started.")
	a.server.Run()

loop:
	for {
		select {
		case b := <-a.stream:
			switch {
			case bytes.HasPrefix(b, []byte(`{"created_at":`)):
				tweet := httpstream.Tweet{}
				err := json.Unmarshal(b, &tweet)

				if err != nil {
					break
				}

				log.Printf("%#v", tweet)
				a.tweetDB.Set([]byte(tweet.Id_str), b)

				// check filter
				// push into database
			}

		case <-a.done:
			log.Print("Client lost connnection.")
			break loop
		}
	}

	a.tweetDB.Close()
	a.flagDB.Close()
	log.Print("Edea application exited gracefully.")
}

func awaitQuitKey(done chan bool) {
	var buf [1]byte
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil || buf[0] == 'q' {
			done <- true
		}
	}
}

func NewApplication(c *Configuration) *Application {
	a := &Application{
		stream: make(chan []byte, 1000),
		done: make(chan bool),
	}

	tweetDB, _ := kv.CreateTemp(".", "tweet_", ".db", &kv.Options{})
	flagDB, _ := kv.CreateTemp(".", "flag_", ".db", &kv.Options{})

	a.tweetDB = tweetDB
	a.flagDB = flagDB

	a.server = NewServer(tweetDB, flagDB)

	a.oc = &oauth.OAuthConsumer{
		Service:          "twitter",
		RequestTokenURL:  "http://twitter.com/oauth/request_token",
		AccessTokenURL:   "http://twitter.com/oauth/access_token",
		AuthorizationURL: "http://twitter.com/oauth/authorize",
		ConsumerKey:      c.ConsumerKey,
		ConsumerSecret:   c.ConsumerSecret,
		CallBackURL:      "oob",
		UserAgent:        "go/httpstream",
	}

	httpstream.OauthCon = a.oc

	a.at = &oauth.AccessToken{
		Id:       "",
		Token:    c.Token,
		Secret:   c.TokenSecret,
		UserRef:  c.UserName,
		Verifier: "",
		Service:  "twitter",
	}

	client := httpstream.NewOAuthClient(a.at, httpstream.OnlyTweetsFilter(func(line []byte) {
		a.stream <- line
	}))

	client.SetMaxWait(5)

	// client.Sample(a.done)
	client.User(a.done)

	return a
}
