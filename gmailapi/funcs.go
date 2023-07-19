package gmailapi

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type MsgInfo struct {
	// Sender is the entity that originally created and sent the message
	Sender string
	// From is the entity that sent the message to you (e.g. googlegroups). Most
	// of the time this information is only relevant to mailing lists.
	From string
	// Subject is the email subject
	Subject string
	// body of email
	Body string
	//recieved time
	Date time.Time
}

var dateLayouts []string

func init() {
	// Generate layouts based on RFC 5322, section 3.3.

	dows := [...]string{"", "Mon, "}   // day-of-week
	days := [...]string{"2", "02"}     // day = 1*2DIGIT
	years := [...]string{"2006", "06"} // year = 4*DIGIT / 2*DIGIT
	seconds := [...]string{":05", ""}  // second
	// "-0700 (MST)" is not in RFC 5322, but is common.
	zones := [...]string{"-0700", "MST", "-0700 (MST)"} // zone = (("+" / "-") 4DIGIT) / "GMT" / ...

	for _, dow := range dows {
		for _, day := range days {
			for _, year := range years {
				for _, second := range seconds {
					for _, zone := range zones {
						s := dow + day + " Jan " + year + " 15:04" + second + " " + zone
						dateLayouts = append(dateLayouts, s)
					}
				}
			}
		}
	}
}

func parseDate(date string) (time.Time, error) {
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("mail: header could not be parsed")
}

func GetLabels() {
	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(Client))

	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	//req := srv.Users.Messages.List("me").
	// list all labels
	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range r.Labels {
		fmt.Printf("- %s\n", l.Name)
	}

}

func GetUnread(timeCutOff time.Time) []MsgInfo {
	ctx := context.Background()
	var allUnreadMsg []MsgInfo
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(Client))
	if err != nil {
		log.Fatalf("Cound not create client: %v", err)
	}
	req := srv.Users.Messages.List("me").Q("label:unread")

	r, err := req.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	for i, m := range r.Messages {

		msg, err := srv.Users.Messages.Get("me", m.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message %v: %v", m.Id, err)
		}
		newMsgINfo := GetMsgHeaders(msg)
		//check if time of email is over 1 hour old sinse last check
		if newMsgINfo.Date.Before(timeCutOff.Add(-1 * time.Hour)) {
			fmt.Println("Found time cut off on email", i+1)
			break
		}
		body, err := GetBody(msg, "text/html")
		if err == nil {
			newMsgINfo.Body = body
		}
		allUnreadMsg = append(allUnreadMsg, newMsgINfo)

		//break
	}
	return allUnreadMsg
}

// GetMsgHeaders gets some of the useful info from the headers.
func GetMsgHeaders(msg *gmail.Message) MsgInfo {
	info := MsgInfo{}
	for _, v := range msg.Payload.Headers {
		switch strings.ToLower(v.Name) {
		case "sender":
			info.Sender = v.Value
		case "from":
			info.From = v.Value
		case "subject":
			info.Subject = v.Value
		case "date":
			newDate, err := parseDate(v.Value)
			if err == nil {
				info.Date = newDate
			}
		default:

		}
	}
	return info
}

func GetBody(msg *gmail.Message, mimeType string) (string, error) {
	// Loop through the message payload parts to find the parts with the
	// mimetypes we want.
	for _, v := range msg.Payload.Parts {
		if v.MimeType == "multipart/alternative" {
			for _, l := range v.Parts {
				if l.MimeType == mimeType && l.Body.Size >= 1 {
					dec, err := decodeEmailBody(l.Body.Data)
					if err != nil {
						return "", err
					}
					return dec, nil
				}
			}
		}
		if v.MimeType == mimeType && v.Body.Size >= 1 {
			dec, err := decodeEmailBody(v.Body.Data)
			if err != nil {
				return "", err
			}
			return dec, nil
		}
	}
	return "", errors.New("couldn't read body")
}

func decodeEmailBody(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
