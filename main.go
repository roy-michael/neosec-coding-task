package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
)

type (
	Event struct {
		ID                  string    `json:"id"`
		Timestamp           Timestamp `json:"timestamp"`
		CallerIP            string    `json:"caller_ip"`
		URL                 string    `json:"url"`
		Method              string    `json:"method"`
		CallPath            string    `json:"call_path"`
		ServerURL           string    `json:"server_url"`
		RequestContentType  string    `json:"request_content_type"`
		RequestSize         int       `json:"request_size"`
		StatusCode          int       `json:"status_code"`
		ResponseContentType string    `json:"response_content_type"`
		ResponseSize        int       `json:"response_size"`
		UserID              string    `json:"user_id"`
		ServiceName         string    `json:"service_name"`
		EndpointPath        string    `json:"endpoint_path"`
		AuthType            string    `json:"auth_type"`
		EndpointID          string    `json:"endpoint_id"`
		AttributesName      []string  `json:"attributes.name"`
		AttributesIn        []string  `json:"attributes.in"`
		AttributesPartOf    []string  `json:"attributes.part_of"`
		AttributesValue     []string  `json:"attributes.value"`
		AttributesValueType []string  `json:"attributes.value_type"`
	}

	db struct {
		userEvents map[string][]Event
		eventIndex map[string]int
	}

	Timestamp time.Time

	server struct {
		db  *db
		srv *http.Server
	}
)

func (t Timestamp) Before(u Timestamp) bool {
	return time.Time(t).Before(time.Time(u))
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {

	ts, err := time.Parse(timeFormat, strings.Trim(string(data), "\""))
	if err != nil {
		return err
	}
	*t = Timestamp(ts)

	return nil
}

func (t Timestamp) String() string {
	return time.Time(t).String()
}

//
//func (t Timestamp) MarshalJSON() ([]byte, error) {
//
//	time.Parse("", )
//
//}

func newServer(addr string, db *db) *server {
	mux := http.NewServeMux()

	s := server{
		db: db,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	mux.HandleFunc("/events", s.eventsHandler)

	return &s
}

func (s *server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func main() {
	fmt.Println("hello world")
	var filename string = "testdata/events-sample.json"

	events, err := readEventFile(filename)
	if err != nil {
		panic(err)
	}

	srv := newServer(":8888", prepareDb(events))

	log.Fatal(srv.ListenAndServe())
}

func (s *server) eventsHandler(writer http.ResponseWriter, request *http.Request) {

	query := request.URL.Query()
	userId := query.Get("userId")
	if userId == "" {
		//todo: error
	}

	eventId := query.Get("eventId")
	//userId := query.Get("userId")

	log.Print("ev", eventId, userId)

	events, ok := s.db.userEvents[userId]
	if !ok {

	}

	for _, evs := range events {
		fmt.Fprintf(writer, "%s: %+v<br>\n", evs.ID, evs.Timestamp)
	}
}

func readEventFile(name string) ([]Event, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	var events []Event
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}
func prepareDb(events []Event) *db {

	userEvents := make(map[string][]Event)
	eventIndex := make(map[string]int)

	// grouping events by user
	for _, event := range events {
		eventList, ok := userEvents[event.UserID]
		if !ok {
			eventList = make([]Event, 0)
		}
		userEvents[event.UserID] = append(eventList, event)
	}

	for user, eventList := range userEvents {

		//sorting the user events by time
		sort.Slice(eventList, func(i, j int) bool {
			return eventList[i].Timestamp.Before(eventList[j].Timestamp)
		})
		userEvents[user] = eventList

		//initializing the event->event index map
		for i, event := range eventList {
			eventIndex[event.ID] = i
		}
	}

	return &db{
		userEvents: userEvents,
		eventIndex: eventIndex,
	}
}
