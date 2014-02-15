package edea

import (
	oauth "github.com/araddon/goauth"
	"github.com/araddon/httpstream"
	"log"
	"os"
)

type Configuration struct {
	UserName, ConsumerKey, ConsumerSecret, Token, TokenSecret string
}

type Application struct {
	oc	*oauth.OAuthConsumer
	at	*oauth.AccessToken

	server	*Server
	curator	*Curator
}


func (a *Application) Run(){
	log.Print("Edea application started.")
	a.curator.Run()
	a.server.Run()
	awaitQuitKey(a.curator.done) // blocks
	log.Print("Edea application exited gracefully.")
}

func awaitQuitKey(done chan bool) {
	var buf [1]byte
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil || buf[0] == 'q' {
			return
		}
	}
}

func NewApplication(c *Configuration) *Application {
	a := &Application{}
	a.curator = NewCurator()
	a.server = NewServer(a.curator)

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
		a.curator.stream <- line
	}))

	client.SetMaxWait(5)

	// client.Sample(a.done)
	client.User(a.curator.done)

	return a
}
