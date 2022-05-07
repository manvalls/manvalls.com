package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type TraktFavourite struct {
	Show  *TraktShow  `json:"show"`
	Movie *TraktMovie `json:"movie"`
}

type TraktMovieIds struct {
	Tmdb  uint64 `json:"tmdb"`
	Slug  string `json:"slug"`
	Trakt string `json:"trakt"`
}

type TraktMovie struct {
	Ids TraktMovieIds `json:"ids"`
}

type TraktShowIds struct {
	Tmdb  uint64 `json:"tmdb"`
	Slug  string `json:"slug"`
	Trakt string `json:"trakt"`
}

type TraktShow struct {
	Ids TraktShowIds `json:"ids"`
}

type ImageLink struct {
	ImageUrl string
	Link     string
}

type TmdbEntity struct {
	PosterPath string `json:"poster_path"`
}

func getShowPoster(id string) string {
	result := ""

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/tv/"+id+"?api_key="+TMDB_API_KEY, nil)
	res, _ := client.Do(req)

	var show TmdbEntity
	json.NewDecoder(res.Body).Decode(&show)
	if show.PosterPath != "" {
		result = "https://image.tmdb.org/t/p/w154" + show.PosterPath
	}

	return result
}

func getMoviePoster(id string) string {
	result := ""

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/movie/"+id+"?api_key="+TMDB_API_KEY, nil)
	res, _ := client.Do(req)

	var movie TmdbEntity
	json.NewDecoder(res.Body).Decode(&movie)
	if movie.PosterPath != "" {
		result = "https://image.tmdb.org/t/p/w154" + movie.PosterPath
	}

	return result
}

var traktFavouritesMux = sync.Mutex{}
var traktFavouritesLastUpdate = time.Unix(0, 0)
var traktFavourites = []TraktFavourite{}

func getTraktFavourites() []TraktFavourite {
	traktFavouritesMux.Lock()
	defer traktFavouritesMux.Unlock()

	if time.Since(traktFavouritesLastUpdate) < time.Hour {
		return traktFavourites
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.trakt.tv/users/"+TRAKT_USER+"/lists/favourites/items?limit=20", nil)
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", TRAKT_API_KEY)
	res, _ := client.Do(req)

	favourites := traktFavourites
	json.NewDecoder(res.Body).Decode(&favourites)

	traktFavouritesLastUpdate = time.Now()
	traktFavourites = favourites
	return favourites
}

type TopTraktData struct {
	ImageLinks []ImageLink
}

func getTopTraktImageLinks() []ImageLink {
	ratings := getTraktFavourites()
	links := []ImageLink{}

	for _, rating := range ratings {
		if rating.Movie != nil {
			imageUrl := getMoviePoster(strconv.FormatUint(rating.Movie.Ids.Tmdb, 10))
			if imageUrl != "" {
				links = append(links, ImageLink{
					ImageUrl: imageUrl,
					Link:     "https://trakt.tv/movies/" + rating.Movie.Ids.Slug,
				})
			}
		}

		if rating.Show != nil {
			imageUrl := getShowPoster(strconv.FormatUint(rating.Show.Ids.Tmdb, 10))
			if imageUrl != "" {
				links = append(links, ImageLink{
					ImageUrl: imageUrl,
					Link:     "https://trakt.tv/shows/" + rating.Show.Ids.Slug,
				})
			}
		}
	}

	return links
}

func handleTopTrakt(rw http.ResponseWriter, r *http.Request) {
	links := getTopTraktImageLinks()
	rw.Write([]byte("document.getElementById('top-trakt').innerHTML = `"))

	t.ExecuteTemplate(rw, "top-trakt", TopTraktData{
		ImageLinks: links,
	})

	rw.Write([]byte("`;"))
}
