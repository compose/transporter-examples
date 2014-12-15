package tstreamer

import (
	"encoding/json"
	"log"
)
import "net/http"
import "bufio"

// TimelineListen - dedicated to listening to a timeline
type TimelineListen struct {
	response *http.Response
	stream   chan Tweet
	client   *api
}

// New provides new reference for specified TimelineListen
func New(endpoint, consumerKey, consumerSecret, accessToken, accessTokenSecret string) (tl *TimelineListen, e error) {
	tl = &TimelineListen{
		client: initAPI(
			consumerKey,
			consumerSecret,
			accessToken,
			accessTokenSecret,
		),
	}
	response, e := tl.client.Get(
		endpoint,
		map[string]string{},
	)
	tl.response = response
	tl.stream = make(chan Tweet)
	return
}

// Listen bytes sent from Twitter Streaming API
// and send completed status to the channel.
func (tl *TimelineListen) Listen() <-chan Tweet {
	scanner := bufio.NewScanner(tl.response.Body)
	go func() {
		for {
			if ok := scanner.Scan(); !ok {
				log.Println(scanner.Err())
				return
				//continue
			}
			status := new(Tweet)

			if err := json.Unmarshal(scanner.Bytes(), &status.Content); err != nil {
				log.Println("(abort)")
				continue
			}
			tl.stream <- *status
		}
	}()
	return tl.stream
}
