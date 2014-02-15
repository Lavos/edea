package edea

import (
	"github.com/cznic/kv"
	"github.com/araddon/httpstream"
	"log"
	"bytes"
	"encoding/json"
)

type Curator struct {
	stream chan []byte
	done chan bool

	tweetDB *kv.DB
	silenceDB *kv.DB
	flagDB *kv.DB
}

func NewCurator() *Curator {
	c := &Curator{
		stream: make(chan []byte),
		done: make(chan bool),
	}

	tweetDB, _ := kv.Open("tweet.db", &kv.Options{})
	flagDB, _ := kv.Open("flag.db", &kv.Options{})
	silenceDB, _ := kv.Open("silence.db", &kv.Options{})

	c.tweetDB = tweetDB
	c.flagDB = flagDB
	c.silenceDB = silenceDB

	return c
}

func (c *Curator) Run() {
	// add timers to maintain silence DB

	log.Print("Edea Curator started.")

	go c.run()
}

func (c *Curator) run() {

loop:
	for {
		select {
		case b := <-c.stream:
			switch {
			case bytes.HasPrefix(b, []byte(`{"created_at":`)):
				tweet := httpstream.Tweet{}
				err := json.Unmarshal(b, &tweet)

				if err != nil {
					break
				}

				log.Printf("%#v", tweet)
				c.tweetDB.Set([]byte(tweet.Id_str), b)
			}

		case <-c.done:
			break loop
		}
	}

	c.tweetDB.Close()
	c.flagDB.Close()
	c.silenceDB.Close()
	log.Print("Edea Curator stopped.")
}

func (c *Curator) GetNext() []byte {
	key, value, _ := c.tweetDB.First()

	c.tweetDB.Delete(key)
	return value
}
