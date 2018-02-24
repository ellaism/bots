package main

import (
	"log"
    "math/big"
	"time"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strings"

    "github.com/onrik/ethrpc"
	"github.com/bwmarrin/discordgo"
)

const bytecode = "0x6060604052341561000f57600080fd5b5b60bf8061001e6000396000f30060606040525b3415600f57600080fd5b5b3373ffffffffffffffffffffffffffffffffffffffff167fd66fd10d93c3fcf37a27c11c0e12214976632505c7954b53c023093d843fc1c460405160405180910390a2600034111560905773ffffffffffffffffffffffffffffffffffffffff33163480156108fc0290604051600060405180830381858888f150505050505b5b0000a165627a7a723058202d5fc395865d1bb6c72766e0d63b91b796ad635bce68bdaa13c76a82c3a1d9980029"
const txCheckInterval = 5 * time.Second

var client = ethrpc.NewEthRPC(os.Getenv("RPC_HOST"))

func main() {
	_, err := client.Web3ClientVersion()
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

func messageDeploy(s *discordgo.Session, m *discordgo.MessageCreate) {
	transaction, err := client.EthSendTransaction(ethrpc.T{
		From: "0x9e2d4a8116c48649ff8b26a67e3f8e4b9ed7cef6",
		GasPrice: big.NewInt(0),
		Value: big.NewInt(0),
		Data: bytecode,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf(transaction)

	for {
		log.Printf("Waiting for tx confirmation: %v", transaction)
		time.Sleep(txCheckInterval)
		receipt, err := client.EthGetTransactionReceipt(transaction)
		if err != nil {
			log.Printf("Failed to get tx receipt for %v", transaction)
			continue
		}

		if receipt != nil {
			s.ChannelMessageSend(m.ChannelID, receipt.ContractAddress)
			break
		}
	}
}

func messageQuery(address string, s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Printf(address)

	logs, err := client.EthGetLogs(ethrpc.FilterParams {
		Address: []string{address},
		FromBlock: "0x0",
	})
	if err != nil {
		log.Printf("Err: %v", err)
		return
	}

	total := big.NewInt(0)
	allVoters := make(map[string]bool)

	for _, log := range logs {
		voter := "0x" + log.Topics[1][len(log.Topics[1])-40:]
		if _, exist := allVoters[voter]; exist {
			continue
		}

		allVoters[voter] = true
		balance, err := client.EthGetBalance(voter, "latest")
		if err != nil {
			fmt.Printf("Err: %v", err)
			balance = *big.NewInt(0)
		}
		fmt.Printf("voter: %v, balance: %v", voter, balance)
		total.Add(total, &balance)
	}

	total.Div(total, big.NewInt(1000000000000000000))
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Total carbon: %v", total))
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!carbon new" {
		messageDeploy(s, m)
		return
	}

	if strings.TrimPrefix(m.Content, "!carbon ") != m.Content {
		messageQuery(strings.TrimPrefix(m.Content, "!carbon "), s, m)
		return
	}
}

