package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type BadRequest struct {
	reason string
}

func (e BadRequest) Error() string {
	return fmt.Sprintf("bad request: %s", e.reason)
}

type BadResponse struct {
	status int
	body   string
}

func (e BadResponse) Error() string {
	return fmt.Sprintf("bad response (status %d): %s", e.status, e.body)
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
		fmt.Println("please specify a SLACK_DOMAIN environment variable")
		os.Exit(1)
	}

	if slackToken = os.Getenv("SLACK_TOKEN"); slackToken == "" {
		fmt.Println("please specify a SLACK_TOKEN environment variable")
		os.Exit(1)
	}

	if slackChannel = os.Getenv("SLACK_CHANNEL"); slackChannel == "" {
		fmt.Println("please specify a SLACK_CHANNEL environment variable")
		os.Exit(1)
	}

	slackHookUrl = fmt.Sprintf("https://%s/services/hooks/incoming-webhook?token=%s", slackDomain, slackToken)

	http.HandleFunc("/notify", notify)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("listening on port %s...\n", port)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func notify(w http.ResponseWriter, r *http.Request) {
	// Extract the message
	message := r.URL.Query().Get("message")
	fmt.Printf("received message: '%s'\n", message)

	// Generate a payload
	data, err := encode(message)
	if err != nil {
		fmt.Printf("error encoding payload: %s\n", err)
		respond(w, 500, err.Error())
		return
	}

	if err = post(data); err != nil {
		fmt.Printf("error sending payload: %s\n", err)
		respond(w, 500, err.Error())
		return
	}

	fmt.Printf("successfully notified slack: '%s'\n", message)

	respond(w, 200, "success")
}

// Encodes a payload as url.Values suitable for a POST
func encode(message string) (data url.Values, err error) {
	if message == "" {
		err = BadRequest{reason: "empty message"}
		return
	}

	p := Payload{
		Channel:     slackChannel,
		Attachments: make([]Attachment, 1),
	}

	p.Attachments[0] = Attachment{
		Fallback: message,
		Text:     message,
	}

	if strings.Contains(message, "UP") {
		p.Attachments[0].Color = "good"
	} else if strings.Contains(message, "DOWN") {
		p.Attachments[0].Color = "danger"
	} else {
		p.Attachments[0].Color = "#cfcfcf"
	}

	buf, err := json.Marshal(p)
	if err != nil {
		return
	}

	data = url.Values{}
	data.Set("payload", string(buf))

	return
}

// Sends a POST request with the given values to Slack
func post(data url.Values) (err error) {
	resp, err := http.PostForm(slackHookUrl, data)
	defer resp.Body.Close()

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		rbuf, _ := ioutil.ReadAll(resp.Body)
		err = BadResponse{status: resp.StatusCode, body: string(rbuf)}
	}

	return
}

func respond(w http.ResponseWriter, status int, response string) {
	w.Header().Add("Content-Length", strconv.Itoa(len([]byte(response))))
	w.WriteHeader(status)
	w.Write([]byte(response))
}
