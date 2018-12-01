package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nmaupu/gonaomi/core"
	"golang.org/x/sync/semaphore"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type GameFile struct {
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
}

var (
	nIp      string
	nPort    int
	rPath    string
	sem      *semaphore.Weighted
	signal   chan bool
	stopChan chan bool
)

func Start(port int, naomiIp string, naomiPort int, romsPath string) {
	nIp = naomiIp
	nPort = naomiPort
	rPath = romsPath
	sem = semaphore.NewWeighted(int64(1))
	signal = make(chan bool, 1)
	stopChan = make(chan bool, 1)

	router := mux.NewRouter()
	router.HandleFunc("/load/{id}", Load).Methods("GET")
	router.HandleFunc("/list", List).Methods("GET")
	router.HandleFunc("/health", Health).Methods("GET")
	router.HandleFunc("/ui", Ui).Methods("GET")

	log.Printf("Starting server listening on 0.0.0.0:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}

func Load(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var game GameFile
	_ = json.NewDecoder(r.Body).Decode(&game)

	game.Name = params["id"]

	select {
	case <-signal:
		log.Println("One time limit hack loop running")
		stopChan <- false
		// Wait process is done before continuing
		<-signal
	default:
		log.Println("There is no time limit hack loop to break")
	}

	if !sem.TryAcquire(1) {
		game.Message = "The naomi is already busy ! Try again later."
		json.NewEncoder(w).Encode(game)
		return
	}

	go func() {
		defer func() {
			sem.Release(1)

			if r := recover(); r != nil {
				log.Println("There was a problem loading the game")
				game.Message = fmt.Sprintf("There was a problem loading the game: %s", r)
				json.NewEncoder(w).Encode(game)
			}
		}()

		naomi := core.NewNaomi(nIp, nPort)
		naomi.ProgressBar = false

		naomi.HOST_SetMode(0, 1)
		naomi.SECURITY_SetKeycode()
		naomi.DIMM_UploadFile(filepath.Join(rPath, fmt.Sprintf("%s.bin", game.Name)))
		naomi.HOST_Restart()

		go func() {
			// Send signal preventing from having two loops at the same time
			signal <- true
			defer func() {
				log.Println("Exiting time limit hack loop")
				naomi.Close()
				signal <- true
			}()

			log.Println("Entering time limit hack loop...")
			loop := true
			for loop {
				select {
				case <-stopChan:
					log.Println("Breaking after receiving the signal")
					loop = false
				default:
					// No stopChan received, looping
					naomi.TIME_SetLimit(10 * 60 * 1000)
					time.Sleep(5000 * time.Millisecond)
				}
			}
		}()
	}()

	// Send response
	game.Message = "Game successfuly sent to the Naomi board."
	json.NewEncoder(w).Encode(game)
}

func getGamesList() []string {
	var games []string

	matches, err := filepath.Glob(filepath.Join(rPath, "*.bin"))
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, m := range matches {
		games = append(
			games,
			strings.TrimSuffix(filepath.Base(m), ".bin"),
		)
	}

	return games
}

func List(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(getGamesList())
}

func Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("UP")
}

func Ui(w http.ResponseWriter, r *http.Request) {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{ .Title }}</title>
	</head>
	<body>
		<h1>{{ .Title }}</h1>
		<ul>
		{{- range .Items }}
		<li><a href="/load/{{ . }}">{{ . }}</a></li>
		{{- else }}
		<li><strong>No game to display</strong></li>
		{{- end }}
		</ul>
	</body>
</html>`

	check := func(err error) {
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
		}
	}

	t, err := template.New("webpage").Parse(tpl)
	check(err)

	data := struct {
		Title string
		Items []string
	}{
		Title: "Naomi games' list",
		Items: getGamesList(),
	}

	err = t.Execute(w, data)
	check(err)
}
