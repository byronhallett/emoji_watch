package main

import (
	"bufio"
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
const emoji_file = "emoji_data"

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

	// load the previous ranks
	prev_ranks := make(map[string]int)
	// try to fetch old store
	file, emoji_err := os.Open(emoji_file)
	if emoji_err != nil {
		log.Println("Save file not found, creating.")
	} else {
		scan := bufio.NewScanner(file)
		rank := 0
		for scan.Scan() {
			prev_ranks[scan.Text()] = rank
			rank += 1
		}
	}

	// construct message
	var message string
	var emoji_string string
	emojis.Each(func(i int, s *goquery.Selection) {
		current_emoji := s.Text()

		// modify rank before aappending
		prev, present := prev_ranks[current_emoji]
		if !present {
			prev = emojis.Size()
		}
		shift := prev - i

		emoji_string += current_emoji + "\n"
		shift_string := strconv.Itoa(shift)
		if shift > 0 {
			shift_string = "↑" + shift_string
		}
		if shift < 0 {
			shift_string = "↓" + shift_string[1:]
		}
		message += strconv.Itoa(i+1) + ": " + current_emoji + ", shift: " + shift_string + "\n"
	})
	message = strings.TrimSpace(message)

	// store emojis for comparison
	file, err = os.Create(emoji_file)
	if err != nil {
		_, err = client.Channel.Send(channel, "Can't save emoji data!")

		log.Fatalln(err.Error())
	}
	file.Write([]byte(emoji_string))

	// send to discord
	_, err = client.Channel.Send(channel, message)

	// discord error handler
	if err != nil {
		log.Fatalln("Error with discord message send")
		log.Fatalln(err.Error())
	}
}
