package huefuncs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amimof/huego"
)

var AppName = "emailhue"
var BridgeConn *huego.Bridge

type hueCreds struct {
	IP       string `json:"ip"`
	UserName string `json:"username"`
}

func FindHub() {
	//192.168.183.4
	creds, err := CredsFromFile()
	if err != nil {
		fmt.Println("no username file found")
		creds = setupHub()
	}

	bridge := huego.New(creds.IP, creds.UserName)

	fmt.Println(bridge.Host)

	_, err = bridge.GetLight(1)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(3)
	}
	BridgeConn = bridge

}

func setupHub() *hueCreds {
	bridge, _ := huego.Discover()
	fmt.Println("need to install app\nClick the button on the Hue HUB and THEN press enter.")
	fmt.Scanln()

	user, err := bridge.CreateUser(AppName)
	if err != nil {
		fmt.Printf("Error creating App Please try again:%s", err.Error())
		os.Exit(3)
	}
	creds := hueCreds{IP: bridge.Host, UserName: user}
	saveCreds(&creds)
	return &creds
}

func CredsFromFile() (*hueCreds, error) {
	f, err := os.Open("hueCreds.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	creds := &hueCreds{}
	err = json.NewDecoder(f).Decode(creds)
	return creds, err
}

func saveCreds(creds *hueCreds) {
	f, err := os.OpenFile("hueCreds.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to save hue creds: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(creds)
}

func GetRooms() []huego.Group {
	var roomGroups []huego.Group
	groups, err := BridgeConn.GetGroups()
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, group := range groups {
		if group.Type == "Room" {
			roomGroups = append(roomGroups, group)
			//fmt.Println(group.Name, len(group.Lights), group.ID)
		}
	}
	return roomGroups
}

func FlashRoom(id int) {

	group, _ := BridgeConn.GetGroup(id)

	group.Bri(254)
	time.Sleep(300 * time.Millisecond)
	group.Alert("lselect")
}
