package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Goscord/goscord"
	"github.com/Goscord/goscord/gateway"
	"github.com/PuerkitoBio/goquery"
)

const emoji_url = "https://emojipedia.org"

var client *gateway.Session

func main() {

	// env vars
	token, found_token := os.LookupEnv("EMOJI_WATCH_TOKEN")
	channel, found_channel := os.LookupEnv("EMOJI_WATCH_CHANNEL")
	if !found_token || !found_channel {
		log.Fatalln("Set env vars: EMOJI_WATCH_TOKEN EMOJI_WATCH_CHANNEL")
	}

	// connect to discord
	client = goscord.New(&gateway.Options{
		Token: token,
	})

	// fetch site
	resp, err := http.Get(emoji_url)
	if err != nil {
		_, _ = client.Channel.Send(channel, "Can't find emoji site!")
		log.Fatalln(err.Error())
	}
	defer resp.Body.Close()

	// parse html
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		_, err = client.Channel.Send(channel, "Can't parse emoji site!")
		log.Fatalln(err.Error())
	}

	// select the popular emojis
	pops := doc.Find("div.block").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.ChildrenFiltered("h2").First().Text() == "Most Popular"
	})
	emojis := pops.ChildrenFiltered("ul").ChildrenFiltered("li")

	// construct message
	var message string
	emojis.Each(func(i int, s *goquery.Selection) {
		message += strconv.Itoa(i+1) + ": " + s.Text() + "\n"
	})
	message = strings.TrimSpace(message)

	// send to discord
	_, err = client.Channel.Send(channel, message)

	// discord error handler
	if err != nil {
		log.Fatalln("Error with discord message send")
		log.Fatalln(err.Error())
	}
}
