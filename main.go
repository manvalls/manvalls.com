package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/ioutil"
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
var profileImage = []byte{}

func getProfileImage() []byte {
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

	req, _ = http.NewRequest("GET", location, nil)

	resp, err = client.Do(req)
	if err != nil {
		return profileImage
	}

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return profileImage
	}

	profileImageLastUpdate = time.Now()
	profileImage = imageData
	return profileImage
}

type Languages struct {
	Spanish string
	English string
}

type Social struct {
	Facebook  string
	Linkedin  string
	Github    string
	Xbox      string
	Lastfm    string
	Trakt     string
	Goodreads string
	Grouvee   string
	Oculus    string
	Discord   string
	Steam     string
	Twitter   string
	Reddit    string
	VRChat    string
}

type HobbyInfo struct {
	Description string
	Emoji       string
}

type Hobbies struct {
	Title string
	List  []HobbyInfo
}

type About struct {
	Title        string
	Atheist      string
	Bisexual     string
	OpenMarriage string
	Parent       string
}

type Misc struct {
	MoreQuotes      string
	GameAccounts    string
	FavouriteMovies string
	MoreMovies      string
	FavouriteGames  string
	MoreGames       string
	FavouriteBooks  string
	MoreBooks       string
	PoweredBy       string

	Facebook  string
	Youtube   string
	Goodreads string
	TMDB      string
	Trakt     string
	Grouvee   string
}

type PageData struct {
	Locale      string
	Title       string
	Description string
	Languages
	Social
	Hobbies
	About
	Misc
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
		locale := getLocale(r)
		pageData := PageData{}

		switch locale {
		case "es":
			pageData = PageData{
				Locale:      locale,
				Title:       "Manuel Valls Fernández",
				Description: "Hola! 😎 Aquí te dejo las cosas que me gustan, enlaces a mis redes sociales y otras mierdas, a seguir bien! 🤘",
				Languages: Languages{
					English: "English",
					Spanish: "Español",
				},
				Social: Social{
					Facebook:  "Facebook",
					Linkedin:  "LinkedIn",
					Github:    "GitHub",
					Xbox:      "Xbox",
					Lastfm:    "last.fm",
					Trakt:     "Trakt",
					Goodreads: "goodreads",
					Grouvee:   "Grouvee",
					Oculus:    "Oculus",
					Discord:   "Discord",
					Steam:     "Steam",
					Twitter:   "Twitter",
					Reddit:    "Reddit",
					VRChat:    "VRChat",
				},
				Hobbies: Hobbies{
					Title: "Cosas que me gustan:",
					List: []HobbyInfo{
						{
							Emoji:       "☕",
							Description: "Café",
						},
						{
							Emoji:       "🇯🇵",
							Description: "Aprender Japonés",
						},
						{
							Emoji:       "🧑🏻‍💻",
							Description: "Desarrollo de Software",
						},
						{
							Emoji:       "🎌",
							Description: "Música Japonesa",
						},
						{
							Emoji:       "🤖",
							Description: "Realidad virtual",
						},
						{
							Emoji:       "👩🏻‍🎤",
							Description: "Música en directo",
						},
						{
							Emoji:       "🍥",
							Description: "Anime y Manga",
						},
						{
							Emoji:       "🎮",
							Description: "Videojuegos",
						},
						{
							Emoji:       "🎞️",
							Description: "Pelis y series",
						},
						{
							Emoji:       "🎼",
							Description: "Música",
						},
						{
							Emoji:       "📖",
							Description: "Leer",
						},
					},
				},
				About: About{
					Title:        "Sobre mí:",
					Atheist:      "Ateo",
					Bisexual:     "Bisexual",
					OpenMarriage: "Matrimonio abierto",
					Parent:       "Padre",
				},
				Misc: Misc{
					MoreQuotes:      "Más citas",
					GameAccounts:    "Cuentas de juego:",
					FavouriteMovies: "Pelis y series favoritas:",
					MoreMovies:      "Ver más",
					FavouriteGames:  "Juegos favoritos:",
					MoreGames:       "Más juegos",
					FavouriteBooks:  "Libros favoritos:",
					MoreBooks:       "Más libros",
					PoweredBy:       "Con tecnología de:",

					Facebook:  "Facebook",
					Youtube:   "YouTube",
					Goodreads: "Goodreads",
					TMDB:      "Themoviedb",
					Trakt:     "Trakt",
					Grouvee:   "Grouvee",
				},
			}
		case "en":
			pageData = PageData{
				Locale:      locale,
				Title:       "Manuel Valls Fernández",
				Description: "Hi there! 😎 Here you'll find things I like, social media links and some more random stuff, keep it up! 🤘",
				Languages: Languages{
					English: "English",
					Spanish: "Español",
				},
				Social: Social{
					Facebook:  "Facebook",
					Linkedin:  "LinkedIn",
					Github:    "GitHub",
					Xbox:      "Xbox",
					Lastfm:    "last.fm",
					Trakt:     "Trakt",
					Goodreads: "goodreads",
					Grouvee:   "Grouvee",
					Oculus:    "Oculus",
					Discord:   "Discord",
					Steam:     "Steam",
					Twitter:   "Twitter",
					Reddit:    "Reddit",
					VRChat:    "VRChat",
				},
				Hobbies: Hobbies{
					Title: "Things I like:",
					List: []HobbyInfo{
						{
							Emoji:       "☕",
							Description: "Coffee",
						},
						{
							Emoji:       "🇯🇵",
							Description: "Learning Japanese",
						},
						{
							Emoji:       "🧑🏻‍💻",
							Description: "Software development",
						},
						{
							Emoji:       "🎌",
							Description: "Japanese music",
						},
						{
							Emoji:       "🤖",
							Description: "Virtual reality",
						},
						{
							Emoji:       "👩🏻‍🎤",
							Description: "Live music",
						},
						{
							Emoji:       "🍥",
							Description: "Anime & Manga",
						},
						{
							Emoji:       "🎮",
							Description: "Videogames",
						},
						{
							Emoji:       "🎞️",
							Description: "Movies & TV",
						},
						{
							Emoji:       "🎼",
							Description: "Music",
						},
						{
							Emoji:       "📖",
							Description: "Reading",
						},
					},
				},
				About: About{
					Title:        "About me:",
					Atheist:      "Atheist",
					Bisexual:     "Bisexual",
					OpenMarriage: "Open marriage",
					Parent:       "Parent",
				},
				Misc: Misc{
					MoreQuotes:      "More quotes:",
					GameAccounts:    "Game accounts:",
					FavouriteMovies: "Favourite movies and shows:",
					MoreMovies:      "View more",
					FavouriteGames:  "Favourite games:",
					MoreGames:       "More games",
					FavouriteBooks:  "Favourite books:",
					MoreBooks:       "More books",
					PoweredBy:       "Powered by:",

					Facebook:  "Facebook",
					Youtube:   "YouTube",
					Goodreads: "Goodreads",
					TMDB:      "Themoviedb",
					Trakt:     "Trakt",
					Grouvee:   "Grouvee",
				},
			}
		}

		t.ExecuteTemplate(rw, "index", pageData)
	})

	http.HandleFunc("/profile.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Write(getProfileImage())
	})

	http.HandleFunc("/top-trakt.js", handleTopTrakt)
	http.Handle("/assets/", http.FileServer(http.FS(assets)))
	http.ListenAndServe(":8090", nil)
}
