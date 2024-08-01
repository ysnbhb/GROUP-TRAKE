package serve

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

func Artists(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	} else if r.Method != http.MethodGet {
		http.Error(w, "mothod not allow", http.StatusMethodNotAllowed)
		return
	}
	artists := []Artist{}
	err := GetData("https://groupietrackers.herokuapp.com/api/artists", &artists)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}
	tmp, err := template.ParseFiles("./template/index.html")
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}
	tmp.Execute(w, artists)
}

func Song(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if len(r.URL.Query()) > 1 || id > "52" || id < "1" {
		http.Error(w, "bad request 400", http.StatusBadRequest)
		return
	} else if _, err := strconv.Atoi(id); err != nil {
		http.Error(w, "bad request 400", http.StatusBadRequest)
		return
	}
	var wg sync.WaitGroup
	artis := &Artis{}

	fetcheData := func(url string, data any) {
		defer wg.Done()
		err := GetData(url, data)
		if err != nil {
			return
		}
	}
	wg.Add(4)
	fetcheData(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%v", id), &artis.Artist)
	fetcheData(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%v", id), &artis.Location)
	fetcheData(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/dates/%v", id), &artis.Date)
	fetcheData(fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%v", id), &artis.Relatoin)
	wg.Wait()
	if artis.Artist.Id == 0 || artis.Date.Id == 0 || artis.Location.Id == 0 || artis.Relatoin.Id == 0 {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-type", "application/json")
	json.NewEncoder(w).Encode(artis)
}

func Style(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/css/" {
		http.NotFound(w, r)
		return
	}
	styleserv := http.FileServer(http.Dir("style"))
	http.StripPrefix("/css", styleserv).ServeHTTP(w, r)
}

func GetData(url string, data any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("errer")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(data)
}

func (r *Serve) Start() error {
	http.HandleFunc("/", Artists)
	http.HandleFunc("/art", Song)
	http.HandleFunc("/api/", Api)
	http.HandleFunc("/api/{id}", GetArtist)
	http.HandleFunc("/css/", Style)
	return http.ListenAndServe(r.Port, nil)
}
