package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/text/language"
)

//go:embed assets
var assets embed.FS

//go:embed templates
var templates embed.FS

var t = template.Must(template.ParseFS(templates, "templates/*.html"))

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

var profileImageMux = sync.Mutex{}
var profileImageLastUpdate = time.Unix(0, 0)
var profileImage = ""

func getProfileImage() string {
	profileImageMux.Lock()
	defer profileImageMux.Unlock()

	if time.Since(profileImageLastUpdate) < time.Hour {
		return profileImage
	}

	resp, err := http.Get("https://graph.facebook.com/oauth/access_token?client_id=" + FB_CLIENT_ID + "&client_secret=" + FB_CLIENT_SECRET + "&grant_type=client_credentials")
	if err != nil {
		return profileImage
	}

	var token tokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return profileImage
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest("GET", "https://graph.facebook.com/v12.0/"+FB_USER_ID+"/picture?height=320&access_token="+url.QueryEscape(token.AccessToken), nil)
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", TRAKT_API_KEY)

	resp, err = client.Do(req)
	if err != nil {
		return profileImage
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return profileImage
	}

	profileImageLastUpdate = time.Now()
	profileImage = location
	return profileImage
}

type PageData struct {
	ProfileImageURL string
	Locale          string
}

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.AmericanEnglish,
	language.BritishEnglish,
	language.Spanish,
	language.EuropeanSpanish,
	language.LatinAmericanSpanish,
})

func getLocale(r *http.Request) string {
	var langCookie string

	lang, err := r.Cookie("lang")
	if err == nil {
		langCookie = lang.Value
	}

	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, langCookie, accept)

	switch tag {
	case language.Spanish:
		return "es"
	case language.EuropeanSpanish:
		return "es"
	case language.LatinAmericanSpanish:
		return "es"
	default:
		return "en"
	}
}

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		url := getProfileImage()

		t.ExecuteTemplate(rw, "index", PageData{
			ProfileImageURL: url,
			Locale:          getLocale(r),
		})
	})

	http.HandleFunc("/top-trakt.js", handleTopTrakt)

	http.Handle("/assets/", http.FileServer(http.FS(assets)))
	http.ListenAndServe(":8090", nil)
}
