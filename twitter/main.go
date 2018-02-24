package main

import (
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/redis.v3"
	"gopkg.in/bsm/ratelimit.v1"

	"strings"
	"fmt"
	"log"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"math/big"
	"sync"
	"time"
)

var twitterClient = twitter.NewClient(oauth1.NewConfig(os.Getenv("TWITTER_CONFIG_USER"), os.Getenv("TWITTER_CONFIG_PASS")).Client(oauth1.NoContext, oauth1.NewToken(os.Getenv("TWITTER_TOKEN_USER"), os.Getenv("TWITTER_TOKEN_PASS"))))
var rpcClient = ethrpc.NewEthRPC(os.Getenv("RPC_HOST"))
var redisClient = redis.NewClient(&redis.Options{
    Addr:     os.Getenv("REDIS_ADDR"),
    Password: os.Getenv("REDIS_PASS"),
    DB:       0,  // use default DB
})

func getTweet(id int64) (*twitter.Tweet, error) {
	tweet, _, err := twitterClient.Statuses.Show(id, &twitter.StatusShowParams{
		TweetMode: "extended",
	})
	return tweet, err
}

func hasELLA(id int64) (bool, string) {
	tweet, err := getTweet(id)
	if err != nil {
		return false, ""
	}
	tweetTime, err := tweet.CreatedAtTime()
	if err != nil {
		return false, ""
	}
	if time.Since(tweetTime) > 24*time.Hour {
		return false, ""
	}
	return strings.Contains(strings.ToLower(tweet.FullText), "$ella"), tweet.User.IDStr
}

func main() {
	_, err := rpcClient.Web3ClientVersion()
    if err != nil {
        log.Fatal(err)
    }

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

func sendFaucet(s *discordgo.Session, m *discordgo.MessageCreate, id int64, address string) {
	visitor := getVisitor(m.Author.ID)
	if visitor.Limit() {
		s.ChannelMessageSend(m.ChannelID, "You used the faucet too much. Try again tomorrow!")
		return
	}

	hasELLA, twitterID := hasELLA(id)
	if !hasELLA {
		visitor.Undo()
		s.ChannelMessageSend(m.ChannelID, "Tweet does not have $ELLA hashtag.")
		return
	}

	if isAbuse(m.Author.ID, twitterID) {
		visitor.Undo()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> You might be abusing the faucet. Please contact <@358106236564144128> if you think this is a mistake.", m.Author.ID))
		return
	}

	balance, err := rpcClient.EthGetBalance("0x231ea5595788f704522c630a22c0b7cc49318ef6", "latest")
	if err != nil {
		log.Fatal(err)
	}

	idString := fmt.Sprintf("%d", id)
	isMember, err := redisClient.SIsMember("twitterfaucet:tweets", idString).Result()
	if err != nil {
		log.Fatal(err)
	}
	if isMember {
		visitor.Undo()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("This Tweet ID has already been claimed."))
		return
	}

	_, err = redisClient.SAdd("twitterfaucet:tweets", idString).Result()
	if err != nil {
		log.Fatal(err)
	}

	count, err := redisClient.SCard("twitterfaucet:tweets").Result()
	if err != nil {
		log.Fatal(err)
	}

	appendRecord(m.Author.ID, twitterID)

	amount := big.NewInt(0).Div(&balance, big.NewInt(2000 + count))
	if amount.Cmp(big.NewInt(100000000000000000)) == 1 {
		amount = big.NewInt(100000000000000000)
	}
	transaction, err := rpcClient.EthSendTransaction(ethrpc.T{
		From: "0x231ea5595788f704522c630a22c0b7cc49318ef6",
		To: address,
		GasPrice: big.NewInt(0),
		Value: amount,
	})
	if err != nil {
		log.Fatal(err)
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> Amount %d Wei sent in transaction %s, claiming https://twitter.com/statuses/%d", m.Author.ID, amount, transaction, id))
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID != "381489991647232025" {
		return
	}

	alls := strings.Fields(m.Content)

	if len(alls) != 4 || alls[0] != "!faucet" {
		return
	}

	if alls[1] == "claim" {
		id, err := strconv.ParseInt(alls[2], 10, 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid Tweet ID.")
			return
		}
		if !common.IsHexAddress(alls[3]) {
			s.ChannelMessageSend(m.ChannelID, "Invalid Ellaism address.")
			return
		}
		sendFaucet(s, m, id, alls[3])
	}
}

type record struct {
	time time.Time
	discordID string
	twitterID string
}

var history = make(map[record]bool)

func appendRecord(discordID string, twitterID string) {
	time := time.Now()
	record := record{time, discordID, twitterID}
	history[record] = true
}

func cleanupHistory() {
	for k, _ := range history {
		if time.Now().Sub(k.time) > 7*24*time.Hour {
			delete(history, k)
		}
	}
}

func isAbuse(discordID string, twitterID string) bool {
	cleanupHistory()

	allTwitterIDs := make(map[string]bool)
	for k, _ := range history {
		if k.discordID == discordID {
			allTwitterIDs[k.twitterID] = true
		}
	}
	allTwitterIDs[twitterID] = true
	if len(allTwitterIDs) > 2 {
		return true
	} else {
		return false
	}
}

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
    limiter  *ratelimit.RateLimiter
    lastSeen time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mtx sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.
func init() {
    go cleanupVisitors()
}

func addVisitor(ip string) *ratelimit.RateLimiter {
    limiter := ratelimit.New(2, 24*time.Hour)
    mtx.Lock()
    // Include the current time when creating a new visitor.
    visitors[ip] = &visitor{limiter, time.Now()}
    mtx.Unlock()
    return limiter
}

func getVisitor(ip string) *ratelimit.RateLimiter {
    mtx.Lock()
    v, exists := visitors[ip]
    if !exists {
        mtx.Unlock()
        return addVisitor(ip)
    }

    // Update the last seen time for the visitor.
    v.lastSeen = time.Now()
    mtx.Unlock()
    return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func cleanupVisitors() {
    for {
        time.Sleep(24*time.Hour)
        mtx.Lock()
        for ip, v := range visitors {
            if time.Now().Sub(v.lastSeen) > 24*time.Hour {
                delete(visitors, ip)
            }
        }
        mtx.Unlock()
    }
}

