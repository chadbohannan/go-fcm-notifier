# go-fcm-notifier
Wrapper around Google Firebase Cloud Messaging REST API that ecapsulates sending notification (messages) to devices from the App Engine Standard Environment.

# Usage in App Engine

```
import (
  "appengine"
  "appengine/urlfetch"
  "net/http"
  ...
)

const FirebaseApiKey = "AIz..."

func HandleCreateNotificationFromFormPost(w http.ResponseWriter, r *http.Request) {
  data := map[string]interface{}{
    "foo": r.FormValue("foo"),
    "bar": r.FormValue("bar"),
  }
  
  c := appengine.NewContext(r)
  client := urlfetch.Client(c)
  fcmn := fcmnotify.NewFcmNotifier(client, FirebaseApiKey).
    SetTopic("/topics/lolwut").
    SetTitle("Notice").
    SetBody("Your attention is needed.").
    SetData(data)

	if _, err := fcmn.Send(); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
		w.Write(err.Error())
  }
}
```
