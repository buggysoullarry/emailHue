package main

import (
	"encoding/json"
	"fmt"
	"hueemail/common"
	"hueemail/gmailapi"
	"hueemail/huefuncs"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amimof/huego"
	"github.com/microcosm-cc/bluemonday"
)

type activeRooms struct {
	Rooms []huego.Group
}

type keyphrase struct {
	Phrases []string
}

var (
	activeKeyPhrase keyphrase
	currentRooms    activeRooms
	lastCheck       time.Time
)

func main() {

	// link up Gmail
	gmailapi.CreateClient()
	//gmailapi.GetLabels()
	//gmailapi.GetUnread()
	// link ot local hue hub
	huefuncs.FindHub()

	// check for keyphrases file
	if !common.FileExists("keyphrase.json") {
		createKeyPhraseFile()
	}
	_, err := getKeyPhraseFromFile()
	common.Check("problem getting keyphrases from file", err)

	fmt.Println(len(activeKeyPhrase.Phrases), "Keyphrases loaded from file.")

	//check for activeRooms
	if !common.FileExists("activerooms.json") {
		createActiveRoomFile()
	}
	err = getCurrentRoomsfromFile()
	common.Check("Problem reading active rooms file", err)

	fmt.Println("Current Room to Alert:", currentRooms.Rooms[0].Name)
	checker()
}

// main function which runs on a loop to check email and it key phrases are found in an emila it will call the alert function
func checker() {
	p := bluemonday.UGCPolicy()
	for {

		fmt.Println("Checking Gmail...")
		msgs := gmailapi.GetUnread(lastCheck)
		ct := time.Now()
		fmt.Println(ct.Format("01/02 03:04:05 pm"), "Checking", len(msgs), "unread emails")
	out:
		for _, msg := range msgs {

			for _, v := range activeKeyPhrase.Phrases {
				if strings.Contains(p.Sanitize(msg.Body), v) {
					fmt.Println("Found Matching email,ALERT!")
					alert()
					break out

				}
			}

		}
		lastCheck = time.Now()
		time.Sleep(1 * time.Minute)
	}
}

// alert- When call it wil flash the lights in all of the active rooms from the save active rooms file
func alert() {
	for _, r := range currentRooms.Rooms {
		huefuncs.FlashRoom(r.ID)
		time.Sleep(1 * time.Second)
	}

}

func createActiveRoomFile() {
	fmt.Println("No Active rooms selected!  ")
	rooms := huefuncs.GetRooms()
	for _, room := range rooms {
		fmt.Printf("\nRoom:%s ID:%d", room.Name, room.ID)
	}
	fmt.Println("\nType the id for the room you wish to alert:")
	var response string
	_, err := fmt.Scanln(&response)
	common.Check("bad response", err)
	intVar, err := strconv.Atoi(response)
	common.Check("Not an ID", err)
	//check if id is for a room
	room, err := huefuncs.BridgeConn.GetGroup(intVar)
	if err != nil {
		log.Fatal("Not a valid ID")
	}
	currentRooms = activeRooms{Rooms: []huego.Group{*room}}

	f, err := os.OpenFile("activerooms.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to save activerooms file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(currentRooms)
	fmt.Println("Room Saved!")

}

func createKeyPhraseFile() {
	fmt.Println("No key phases exist! \n Type the first Key Phrase to search emails for:")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	activeKeyPhrase = keyphrase{Phrases: []string{response}}

	f, err := os.OpenFile("keyphrase.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to save keyphrase file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(activeKeyPhrase)
	fmt.Println("You can add more key phrases in the keyphrase.json file ")
}

func getKeyPhraseFromFile() (*keyphrase, error) {
	tok := &activeKeyPhrase
	f, err := os.Open("keyphrase.json")
	if err != nil {
		return tok, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getCurrentRoomsfromFile() error {
	tok := &currentRooms
	f, err := os.Open("activerooms.json")
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(tok)
	return err
}
