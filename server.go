package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"

	"github.com/gorilla/mux"
)

//Player Joueur nouveau Joueur
type Player struct {
	Name  string    `json:"nom"`
	Ordre int       `json:"ordre"`
	Date  time.Time `json:"date"`
}

//NotifMess message de la notification
type NotifMess struct {
	Body string `json:"body"`
}

//Notif contenu de la notification
type Notif struct {
	Notification NotifMess `json:"notification"`
	To           string    `json:"to"`
}

func getPlayersKey(c appengine.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Players", "default_match", 0, nil)
}

func makeRouter() *mux.Router {
	r := mux.NewRouter()
	app := r.Headers("X-Requested-With", "XMLHttpRequest").Subrouter()
	app.HandleFunc("/getplayers", getPlayers).Methods(http.MethodGet)
	app.HandleFunc("/sendnotif/{name}", sendNotif).Methods(http.MethodGet)
	app.HandleFunc("/removeplayers", removePlayers).Methods(http.MethodGet)
	return r
}

func init() {
	router := makeRouter()

	http.Handle("/", router)
}

func dataGetPlayers(c appengine.Context) (ps []Player, err error) {

	q := datastore.NewQuery("Players").Ancestor(getPlayersKey(c)).Order("-Date").Limit(10)
	ps = make([]Player, 0, 10)
	_, err = q.GetAll(c, &ps)

	return
}

func getPlayers(w http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	ps, err := dataGetPlayers(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ps); err != nil {
		panic(err)
	}
	//for _, p := range ps {
	//	fmt.Fprintf(w, "Nom : %v\n", p.Name)
	//}
}

//senNotif enregistre les joureus et envoie une notification à tos les joureurs
func sendNotif(w http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	v := mux.Vars(req)
	p := Player{
		Name: v["name"],
		Date: time.Now(),
	}
	ps, err := dataGetPlayers(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ps) >= 6 {
		fmt.Fprint(w, "MAX")
		return
	}
	key := datastore.NewKey(c, "Players", p.Name, 0, getPlayersKey(c))
	var pp Player

	if erg := datastore.Get(c, key, pp); erg != datastore.ErrNoSuchEntity {
		fmt.Fprint(w, "Player is already in the match")
		return
	}
	if _, err := datastore.Put(c, key, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := urlfetch.Client(c)

	//AAAAk_ocS08:APA91bEtTBcNmAvgUI7O_nexdl7XkDVbsCEOazr-hvBh__US77ct34xadwZLqUqgob1q6WlhmBAy9jio32iVLSU2DO-R7sM7LlULRrzUKfsjjWjmwQnFyTh3hDDNnw6Xd5EooVkMEda-

	var notif Notif
	notif.Notification = NotifMess{
		Body: `"Alerte baby - Let's GO !"`,
	}
	notif.To = "/topics/baby"

	b := new(bytes.Buffer)
	if erro := json.NewEncoder(b).Encode(notif); erro != nil {
		panic(erro)
	}

	//reqc, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", strings.NewReader(`{"notification" : { "body" : "Hello baby - Yo!"},"to" : "/topics/baby"}`))
	reqc, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", b)
	reqc.Header.Add("Authorization", "key=AAAAk_ocS08:APA91bEtTBcNmAvgUI7O_nexdl7XkDVbsCEOazr-hvBh__US77ct34xadwZLqUqgob1q6WlhmBAy9jio32iVLSU2DO-R7sM7LlULRrzUKfsjjWjmwQnFyTh3hDDNnw6Xd5EooVkMEda-")
	reqc.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(reqc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "HTTP POST returned status %v", resp.Status)

}

func removePlayers(w http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)

	q := datastore.NewQuery("Players").Ancestor(getPlayersKey(c))
	ps := make([]Player, 0, 10)
	keys, err := q.GetAll(c, &ps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = datastore.DeleteMulti(c, keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// envoi d'une notif aux joureurs
	var notif Notif
	notif.Notification = NotifMess{
		Body: `"Partie Supprimée..."`,
	}
	notif.To = "/topics/baby"

	b := new(bytes.Buffer)
	if erro := json.NewEncoder(b).Encode(notif); erro != nil {
		panic(erro)
	}

	//reqc, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", strings.NewReader(`{"notification" : { "body" : "Hello baby - Yo!"},"to" : "/topics/baby"}`))
	reqc, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", b)
	reqc.Header.Add("Authorization", "key=AAAAk_ocS08:APA91bEtTBcNmAvgUI7O_nexdl7XkDVbsCEOazr-hvBh__US77ct34xadwZLqUqgob1q6WlhmBAy9jio32iVLSU2DO-R7sM7LlULRrzUKfsjjWjmwQnFyTh3hDDNnw6Xd5EooVkMEda-")
	reqc.Header.Add("Content-Type", "application/json")

	client := urlfetch.Client(c)

	_, err = client.Do(reqc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Delete OK")
}
