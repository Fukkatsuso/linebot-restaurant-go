package places

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

// AlternativePhotoURI returns uri of line-cdn-clip
func AlternativePhotoURI() string {
	id := "12" // 1 ~ 13
	clip := "clip" + id + ".jpg"
	uri := "https://scdn.line-apps.com/n/channel_devcenter/img/flexsnapshot/clip/" + clip
	return uri
}

// ErrRedirectAttempted errors
var ErrRedirectAttempted = errors.New("redirect")

// GoogleMapPhotoURI returns uri of googlemap-photo
func GoogleMapPhotoURI(params map[string]string) string {
	uri := "https://maps.googleapis.com/maps/api/place/photo?"
	for k, v := range params {
		if uri[len(uri)-1] != '?' {
			uri += "&"
		}
		uri += k + "=" + v
	}

	client := &http.Client{
		Timeout: time.Duration(3) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		},
	}
	resp, err := client.Head(uri)
	if urlError, ok := err.(*url.Error); !(ok && urlError.Err == ErrRedirectAttempted) {
		return ""
	}
	defer resp.Body.Close()
	return resp.Header["Location"][0]
}
