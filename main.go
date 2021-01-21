package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
)

type (
	inputEvent struct {
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

	// db holds all the read events
	db struct {
		userEvents map[string][]inputEvent // user id -> events sorted by timestamp
		eventIndex map[string]int          // event id -> index in events list
	}

	// Timestamp for reading and writing the event time format
	Timestamp time.Time

	server struct {
		db  *db
		srv *http.Server
	}
)

var addr = flag.String("addr", ":8888", "address to listen on")

func (e inputEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&struct {
			ID           string    `json:"id"`
			Timestamp    Timestamp `json:"timestamp"`
			CallerIP     string    `json:"caller_ip"`
			URL          string    `json:"url"`
			Method       string    `json:"method"`
			CallPath     string    `json:"call_path"`
			ServerURL    string    `json:"server_url"`
			StatusCode   int       `json:"status_code"`
			UserID       string    `json:"user_id"`
			ServiceName  string    `json:"service_name"`
			EndpointPath string    `json:"endpoint_path"`
			EndpointID   string    `json:"endpoint_id"`
		}{
			ID:           e.ID,
			Timestamp:    e.Timestamp,
			CallerIP:     e.CallerIP,
			URL:          e.URL,
			Method:       e.Method,
			CallPath:     e.CallPath,
			ServerURL:    e.ServerURL,
			StatusCode:   e.StatusCode,
			UserID:       e.UserID,
			ServiceName:  e.ServiceName,
			EndpointPath: e.EndpointPath,
			EndpointID:   e.EndpointID,
		})
}

func (t Timestamp) Before(u Timestamp) bool {
	return time.Time(t).Before(time.Time(u))
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {

	// assuming dates are in UTC without a TZ
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

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t))
}

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

	flag.Parse()

	log.Print("starting server. listening on ", *addr)
	var filename string = "testdata/events-sample.json"

	events, err := readEventFile(filename)
	if err != nil {
		panic(err)
	}

	srv := newServer(*addr, prepareDb(events))

	log.Fatal(srv.ListenAndServe())
}

func (s *server) eventsHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodGet {
		log.Printf("method not allowed: %v", request.Method)
		http.Error(writer, "no user id was provided", http.StatusMethodNotAllowed)
		return
	}

	query := request.URL.Query()
	userId := query.Get("userId")
	eventId := query.Get("eventId")
	limit := query.Get("limit")
	page := query.Get("page")

	if userId == "" {
		log.Printf("no user id was provided")
		http.Error(writer, "no user id was provided", http.StatusBadRequest)
		return
	}

	log.Print("params:", eventId, userId)

	events, err := s.getEventList(userId, eventId, limit, page)
	if err != nil {
		log.Printf("error while processing event list: %v", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	//if limit, err := strconv.Atoi(limitQry); err == nil {
	//	log.Print("limiting event list ", limit, " ", limitQry)
	//	events = events[:limit]
	//}

	//for _, evs := range events {
	//	fmt.Fprintf(writer, "%s: %+v<br>\n", evs.ID, evs.Timestamp)
	//}
	writer.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(writer).Encode(&events); err != nil {
		log.Print("error while writing response: ", err)
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
	}
}

// reading the json file from disk and unmarshalling into an inputEvent list
func readEventFile(name string) ([]inputEvent, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	var events []inputEvent
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		var event inputEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

//preparing the userEvent and lists and the event index map
func prepareDb(events []inputEvent) *db {

	userEvents := make(map[string][]inputEvent)
	eventIndex := make(map[string]int)

	// grouping events by user
	for _, event := range events {
		eventList, ok := userEvents[event.UserID]
		if !ok {
			eventList = make([]inputEvent, 0)
		}
		userEvents[event.UserID] = append(eventList, event)
	}

	for user, eventList := range userEvents {

		//sorting the user events by their timestamp
		sort.Slice(eventList, func(i, j int) bool {
			return eventList[i].Timestamp.Before(eventList[j].Timestamp)
		})
		userEvents[user] = eventList

		//initializing the event-id->event-index map
		for i, event := range eventList {
			eventIndex[event.ID] = i
		}
	}

	return &db{
		userEvents: userEvents,
		eventIndex: eventIndex,
	}
}

// reading events from the user event list, according to requirements
func (s *server) getEventList(userId, eventId, limitStr, pageStr string) ([]inputEvent, error) {

	log.Printf("reading event list. user %s, event %s, limit: %s, page: %s", userId, eventId, limitStr, pageStr)
	events, ok := s.db.userEvents[userId]
	if !ok {
		return nil, fmt.Errorf("cannot find user events for %s", userId)
	}

	log.Printf("found %d events in for user %s", len(events), userId)

	start, end := 0, len(events)
	if eventId != "" {
		index := 0
		log.Printf("looking for event index in timeline list %s", eventId)

		if index, ok = s.db.eventIndex[eventId]; !ok {
			return nil, fmt.Errorf("cannot find index for event: %s", eventId)
		}

		log.Printf("found index for event: %s, %d", eventId, index)
		start = index
		end = index + 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err == nil {
		start = int(math.Max(0, float64(start-limit/2)))
		end = int(math.Min(float64(len(events)), float64(start+limit)))

		log.Printf("using start and end indexes: %d, %d", start, end)

		// reading some more from the bottom in case the end is last element in list
		if delta := end - start; delta < limit {
			start = int(math.Max(0, float64(start-(limit-delta))))
		}
	}

	start, end = paginate(start, end, limit, len(events), pageStr)

	return events[start:end], nil
}

func paginate(start, end, limit, max int, pageStr string) (int, int) {

	page, err := strconv.Atoi(pageStr)
	if err == nil {
		offset := page * limit
		start = int(math.Min(math.Max(0, float64(start+offset)), float64(max))) //make sure it is within boundaries
		end = int(math.Max(math.Min(float64(max), float64(end+offset)), float64(0)))
	}

	return start, end
}
