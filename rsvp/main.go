package main

import (
	"github.com/bwmarrin/discordgo"
	"gopkg.in/redis.v3"

	"strings"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var redisClient = redis.NewClient(&redis.Options{
    Addr:     os.Getenv("REDIS_ADDR"),
    Password: os.Getenv("REDIS_PASS"),
    DB:       0,
})

func main() {
	dg, err := discordgo.New(os.Getenv("DISCORD"))
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(messageCreate)
	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	alls := strings.Fields(strings.ToLower(m.Content))

	if len(alls) < 1 || alls[0] != "!rsvp" {
		return
	}

	channelID := m.ChannelID
	authorID := m.Author.ID

	if channelID != "375680179009224715" {
		s.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> Please use the <#%s> channel to RSVP **Weekly General Development Meeting** on Saturday (24 Feb) 4pm UTC.", authorID, "375680179009224715"))
		return
	}

	if len(alls) == 2 && alls[1] == "ping" {
		if authorID == "358106236564144128" {
			members, err := redisClient.SMembers("rsvpbot:375680179009224715").Result()
			if err != nil {
				log.Fatal(err)
			}
			var memberStrings []string
			for _, str := range members {
				memberStrings = append(memberStrings, fmt.Sprintf("<@%s>", str))
			}
			s.ChannelMessageSend(channelID, strings.Join(memberStrings, " "))
		}
		return
	}

	if len(alls) == 2 && alls[1] == "clear" {
		if authorID == "358106236564144128" {
			_, err := redisClient.Del("rsvpbot:375680179009224715").Result()
			if err != nil {
				log.Fatal(err)
			}
			s.ChannelMessageSend(channelID, "RSVP list cleared!")
		}
		return
	}

	if len(alls) == 2 && alls[1] == "total" {
		count, err := redisClient.SCard("rsvpbot:375680179009224715").Result()
		if err != nil {
			log.Fatal(err)
		}
		s.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> There are in total %d people RSVPed to the next **Weekly General Development Meeting**.", authorID, count))
		return
	}

	if len(alls) == 1 || (len(alls) > 1 && alls[1] != "no") {
		isMember, err := redisClient.SIsMember("rsvpbot:375680179009224715", authorID).Result()
		if err != nil {
			log.Fatal(err)
		}

		if isMember {
			s.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> You have already RSVPed!", authorID))
		} else {
			_, err = redisClient.SAdd("rsvpbot:375680179009224715", authorID).Result()
			if err != nil {
				log.Fatal(err)
			}
			s.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> You have RSVPed to **Weekly General Development Meeting** on Saturday (24 Feb) 4pm UTC. Remember to be there!", authorID))
		}

		return
	}
}

