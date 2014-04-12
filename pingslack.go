package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Message struct {
	Check       string `json:"check"`
	IncidentId  int    `json:"incidentid"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

func (m Message) Title() string {
	switch m.Action {
	case "notify_of_close":
		return "Event Closed"
	case "assign":
		return "Event Assigned"
	}

	return strings.Title(m.Action)
}

func (m Message) Color() string {
	if m.Action == "notify_of_close" {
		return "good"
	} else {
		return "danger"
	}
}

type Payload struct {
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback string `json:"fallback"`
	Text     string `json:"text"`
	Pretext  string `json:"pretext"`
	Color    string `json:"color"`
}

var (
	slackDomain  string
	slackToken   string
	slackChannel string
	slackHookUrl string
)

func main() {
	if slackDomain = os.Getenv("SLACK_DOMAIN"); slackDomain == "" {
		log.Fatalf("Please specify a SLACK_DOMAIN environment variable")
	}

	if slackToken = os.Getenv("SLACK_TOKEN"); slackToken == "" {
		log.Fatalf("Please specify a SLACK_TOKEN environment variable")
	}

	if slackChannel = os.Getenv("SLACK_CHANNEL"); slackChannel == "" {
		log.Fatalf("Please specify a SLACK_CHANNEL environment variable")
	}

	slackHookUrl = "https://" + slackDomain + "/services/hooks/incoming-webhook?token=" + slackToken

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		msg := &Message{}
		if err := json.Unmarshal([]byte(r.URL.Query().Get("message")), msg); err != nil {
			respond(w, 500, map[string]string{"error": err.Error()})
			return
		}

		log.Printf("received from pingdom: %+v", msg)

		// Create a Slack notification payload
		payload := Payload{
			Channel:     slackChannel,
			Attachments: make([]Attachment, 1),
		}

		payload.Attachments[0] = Attachment{
			Fallback: msg.Description,
			Text:     msg.Description,
			Pretext:  msg.Title(),
			Color:    msg.Color(),
		}

		// Encode the payload
		buf, err := json.Marshal(payload)
		if err != nil {
			log.Printf("error: %s", err)
			respond(w, 500, map[string]string{"error": err.Error()})
			return
		}
		data := url.Values{}
		data.Set("payload", string(buf))

		// Send it off
		log.Printf("notifying slack: %+v", payload)
		resp, err := http.PostForm(slackHookUrl, data)
		defer resp.Body.Close()
		if err != nil {
			log.Printf("error: %s", err)
			respond(w, 500, map[string]string{"error": err.Error()})
			return
		}
		if resp.StatusCode != 200 {
			rbuf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("error: %s", err)
				respond(w, 500, map[string]string{"error": err.Error()})
			} else {
				log.Printf("unexpected response (%d): %s", resp.StatusCode, rbuf)
				respond(w, 500, map[string]string{"error": string(rbuf)})
			}
			return
		}

		log.Printf("success!")

		respond(w, 200, map[string]string{"status": "notified"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s...", port)

	http.ListenAndServe(":"+port, nil)
}

func respond(w http.ResponseWriter, status int, object interface{}) {
	buf, _ := json.Marshal(object)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(len(buf)))
	w.WriteHeader(status)
	w.Write(buf)
}
