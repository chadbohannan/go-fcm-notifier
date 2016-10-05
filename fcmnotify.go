package fcmnotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TODO:: support asynchronous API calls

var FcmServiceUrl = "https://fcm.googleapis.com/fcm/send"

const MaxTTL = 2419200 // 4 weeks, in seconds

// FcmNotifier is the root structure of this library
type FcmNotifier struct {
	ApiKey     string
	Message    FcmMsg
	httpClient *http.Client
}

// body of the POST to the Firebase API
type FcmMsg struct {
	Data                  interface{}         `json:"data,omitempty"`
	To                    string              `json:"to,omitempty"`
	RegistrationIds       []string            `json:"registration_ids,omitempty"`
	CollapseKey           string              `json:"collapse_key,omitempty"`
	Priority              string              `json:"priority,omitempty"`
	Notification          NotificationPayload `json:"notification,omitempty"`
	ContentAvailable      bool                `json:"content_available,omitempty"`
	TimeToLive            int                 `json:"time_to_live,omitempty"`
	RestrictedPackageName string              `json:"restricted_package_name,omitempty"`
	DryRun                bool                `json:"dry_run,omitempty"`
	Condition             string              `json:"condition,omitempty"`
}

// NotificationPayload is an optional substructure of the message body
type NotificationPayload struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

// SendResponse captures the result of the HTTP POST
type SendResponse struct {
	Ok            bool
	StatusCode    int
	MulticastId   int                 `json:"multicast_id"`
	Success       int                 `json:"success"`
	Fail          int                 `json:"failure"`
	Canonical_ids int                 `json:"canonical_ids"`
	Results       []map[string]string `json:"results,omitempty"`
	MsgId         int                 `json:"message_id,omitempty"`
	Err           string              `json:"error,omitempty"`
}

// Initialize an FcmNotifier with which to create Firebase Notifications
func NewFcmNotifier(client *http.Client, apiKey string) *FcmNotifier {
	fcmn := new(FcmNotifier)
	fcmn.ApiKey = apiKey
	fcmn.httpClient = client
	return fcmn
}

// SetTopic sets the notification topic.
func (fcmn *FcmNotifier) SetTopic(to string) *FcmNotifier {
	fcmn.Message.To = to
	return fcmn
}

// SetRegistrationIds sets the targets to the specified devices.
// Not to be used in conjuction with SetTopic(string).
func (fcmn *FcmNotifier) SetRegistrationIds(list []string) *FcmNotifier {
	fcmn.Message.RegistrationIds = make([]string, len(list))
	copy(fcmn.Message.RegistrationIds, list) // deep copy the token list
	return fcmn
}

// SetTitle sets a reserved-name value in the Notification payload
func (fcmn *FcmNotifier) SetTitle(value string) *FcmNotifier {
	fcmn.Message.Notification.Title = value
	return fcmn
}

// SetBody sets a reserved-name value in the Notification payload
func (fcmn *FcmNotifier) SetBody(value string) *FcmNotifier {
	fcmn.Message.Notification.Body = value
	return fcmn
}

// SetBody sets a reserved-name value in the Notification payload
func (fcmn *FcmNotifier) SetIcon(value string) *FcmNotifier {
	fcmn.Message.Notification.Icon = value
	return fcmn
}

// SetCollapseKey identifies a group of messages (e.g., with collapse_key:
// "Updates Available") that can be collapsed, so that only the last message
// gets sent when delivery can be resumed. This is intended to avoid sending
// too many of the same messages when the device comes back online or becomes
// active (see delay_while_idle).
func (fcmn *FcmNotifier) SetCollapseKey(value string) *FcmNotifier {
	fcmn.Message.CollapseKey = value
	return fcmn
}

// SetCondition specifies a logical expression of conditions that determine the message target.
// Supported condition: Topic, formatted as "'yourTopic' in topics". This value is case-insensitive.
// Supported operators: &&, ||. Maximum two operators per topic message supported.
func (fcmn *FcmNotifier) SetCondition(value string) *FcmNotifier {
	fcmn.Message.Condition = value
	return fcmn
}

// SetContentAvailable On iOS, use this field to represent content-available
// in the APNS payload. When a notification or message is sent and this is set
// to true, an inactive client app is awoken. On Android, data messages wake
// the app by default. On Chrome, currently not supported.
func (fcmn *FcmNotifier) SetContentAvailable(value bool) *FcmNotifier {
	fcmn.Message.ContentAvailable = value
	return fcmn
}

// SetData sets the data payload
func (fcmn *FcmNotifier) SetData(value interface{}) *FcmNotifier {
	fcmn.Message.Data = value
	return fcmn
}

// SetDryRun This parameter, when set to true, allows developers to test
// a request without actually sending a message.
// The default value is false.
func (fcmn *FcmNotifier) SetDryRun(value bool) *FcmNotifier {
	fcmn.Message.DryRun = value
	return fcmn
}

// SetHighPriority overrides the default 'normal' priority
func (fcmn *FcmNotifier) SetHighPriority() {
	fcmn.Message.Priority = "high"
}

// SetRestrictedPackageName specifies the package name of the application
// where the registration tokens must match in order to receive the message.
func (fcmn *FcmNotifier) SetRestrictedPackageName(value string) *FcmNotifier {
	fcmn.Message.RestrictedPackageName = value
	return fcmn
}

// SetTimeToLive specifies how long (in seconds) the message should be kept
// in FCM storage if the device is offline. The maximum time to live supported
// is 4 weeks, and the default value is 4 weeks. For more information, see
// https://firebase.google.com/docs/cloud-messaging/concept-options#ttl
func (fcmn *FcmNotifier) SetTTL(value int) *FcmNotifier {
	if value > MaxTTL {
		fcmn.Message.TimeToLive = MaxTTL
	} else {
		fcmn.Message.TimeToLive = value
	}
	return fcmn
}

// Send executes a POST to the Firebase API
func (fcmn *FcmNotifier) Send() (*SendResponse, error) {
	sendResponse := new(SendResponse)
	body, err := json.Marshal(fcmn.Message)
	if err != nil {
		return sendResponse, err
	}

	request, err := http.NewRequest("POST", FcmServiceUrl, bytes.NewBuffer(body))
	request.Header.Set("Authorization", fmt.Sprintf("key=%v", fcmn.ApiKey))
	request.Header.Set("Content-Type", "application/json")

	var response *http.Response
	if response, err = fcmn.httpClient.Do(request); err == nil {
		defer response.Body.Close()
		if body, err = ioutil.ReadAll(response.Body); err == nil {
			sendResponse.StatusCode = response.StatusCode
			if response.StatusCode == http.StatusOK {
				if err = json.Unmarshal(body, sendResponse); err == nil {
					sendResponse.Ok = true
				}
			}
		}
	}
	return sendResponse, err
}
