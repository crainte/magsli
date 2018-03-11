package mailgun

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// Data defines the fields from MailGun that we are going to use for our Slack message.
type Data struct {
	EventType string
	Domain    string
	Recipient string
	MessageID string
	Subject   string

	// bounced
	// https://documentation.mailgun.com/en/latest/user_manual.html#tracking-bounces
	SMTPCode  string
	SMTPError string

	// dropped
	// https://documentation.mailgun.com/en/latest/user_manual.html#tracking-failures
	Reason      string
	ESPCode     string
	Description string
}

// IsErrorEvent returns a bool indicating if the event type should be considered an error.
func (d *Data) IsErrorEvent() bool {
	return (d.EventType == "bounced") ||
		(d.EventType == "dropped") ||
		(d.EventType == "failed") ||
		(d.EventType == "rejected")
}

// VerifyMessage verifies that a message is coming from our MailGun API.
//  See https://documentation.mailgun.com/en/latest/user_manual.html#webhooks
func VerifyMessage(apiKey string, req *http.Request) (bool, error) {
	mac := hmac.New(sha256.New, []byte(apiKey))

	mac.Write([]byte(req.FormValue("timestamp")))
	mac.Write([]byte(req.FormValue("token")))

	expectedMAC := mac.Sum(nil)

	signature, err := hex.DecodeString(req.FormValue("signature"))

	if err != nil {
		return false, err
	}

	if len(expectedMAC) != len(signature) {
		return false, nil
	}

	return hmac.Equal([]byte(signature), expectedMAC), nil
}

// NewMailGunData decodes a MailGun POST message and returns the data we are going to use.
func NewMailGunData(r *http.Request) (Data, error) {
	err := r.ParseForm()

	if err != nil {
		return Data{}, err
	}

	// For debugging
	// for key, values := range r.PostForm {
	// 	fmt.Printf("%v : %v\n", key, values)
	// }

	event := r.FormValue("event")

	data := Data{
		EventType: event,
		Domain:    r.FormValue("domain"),
		Recipient: r.FormValue("recipient"),
		MessageID: r.FormValue("Message-Id"),
		Subject:   getSubjectFromHeaders(r.FormValue("message-headers")),
	}

	switch event {
	case "bounced":
		data.SMTPCode = r.FormValue("code")
		data.SMTPError = r.FormValue("error")

	case "dropped":
		data.Reason = r.FormValue("reason")
		data.ESPCode = r.FormValue("code")
		data.Description = r.FormValue("description")
	}

	return data, nil
}

func getSubjectFromHeaders(h string) string {
	// We need some funkiness to get our subject because the headers are stored in "not-really-JSON".
	//  See https://documentation.mailgun.com/en/latest/user_manual.html#webhooks
	// The MailGun docs say:
	//	"String list of all MIME headers of the original message dumped to a JSON string (order of headers preserved)."
	// It's in a strange format that's technically JSON, but not convenient for developers.
	// It's an array of string arrays for key/value pairs like this:
	//	["Subject", "Test bounces webhook"], ["From", "Bob <bob@foo.com>"]
	// So it's complicated to use - but since the order is preserved we can index to get what what we want.

	var arbitraryJSON []interface{}

	json.Unmarshal([]byte(h), &arbitraryJSON)

	subjectArray := (arbitraryJSON[3]).([]interface{})

	if subjectArray[0].(string) != "Subject" {
		return "<could not get subject>"
	}

	return subjectArray[1].(string)
}
