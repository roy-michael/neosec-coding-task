package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

const filename = "testdata/events-sample.json"

func TestPrepareDB(t *testing.T) {

	events, err := readEventFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	db := prepareDb(events)
	if count := len(db.userEvents); count != 2 {
		t.Errorf("expected user count of %d, but got %d", 2, count)
	}

	for _, events := range db.userEvents {
		for i, event := range events {

			if index := db.eventIndex[event.ID]; index != i {
				t.Errorf("wrong event index. expected %d but got %d", i, index)
			}

			if i != 0 && time.Time(events[i-1].Timestamp).After(time.Time(event.Timestamp)) {
				t.Errorf("event list is not sorted well: %+v, %+v", events[i-1], event)
			}
		}
	}
}

func TestServer_Events(t *testing.T) {

	events, err := readEventFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	db := prepareDb(events)

	type args struct {
		userId  string
		eventId string
		limit   string
		page    string
	}

	tests := []struct {
		name           string
		args           args
		expectedCode   int
		expectedLen    int
		expectedEvents []string
	}{
		{
			name:         "userId not found",
			args:         args{userId: "31b20726-b870-47ba-bbcd-372b38527777"},
			expectedCode: http.StatusBadRequest,
			expectedLen:  0,
		},
		{
			name:         "userId found",
			args:         args{userId: "31b20726-b870-47ba-bbcd-372b38527c89"},
			expectedCode: http.StatusOK,
			expectedLen:  89,
		},
		{
			name:           "eventId list start",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "06ced2b3-9b5a-4f11-9d55-efdccf650aaa"},
			expectedCode:   http.StatusOK,
			expectedLen:    1,
			expectedEvents: []string{"06ced2b3-9b5a-4f11-9d55-efdccf650aaa"},
		},
		{
			name:           "eventId list start, limit 5",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "06ced2b3-9b5a-4f11-9d55-efdccf650aaa", limit: "5"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"06ced2b3-9b5a-4f11-9d55-efdccf650aaa", "193b93be-bbc4-4fb4-be46-926185536c1a", "c573ac81-3a04-4812-a44d-576d91e9fa7e", "b161ade0-e1bf-4cc0-b729-e66f9c87e4f1", "a805a1d3-23a5-4779-a6aa-c4514014d75e"},
		},
		{
			name:           "eventId list start, limit 5, page 2",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "06ced2b3-9b5a-4f11-9d55-efdccf650aaa", limit: "5", page: "2"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"ab7a1328-8ff8-4098-a91b-fe588eb49485", "4cdfe358-68a4-4c90-9d4e-3138e2ed807c", "e73e0031-d1e0-4cd4-8a84-cef08cb906f6", "95e9683f-0b0d-43b0-9b9d-d4f76399430a", "78cdb479-045a-48ee-a2d2-02a88e765c89"},
		},
		{
			name:         "eventId list start, limit 5, page -2 (out of range)",
			args:         args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "06ced2b3-9b5a-4f11-9d55-efdccf650aaa", limit: "5", page: "-2"},
			expectedCode: http.StatusOK,
			expectedLen:  0,
		},
		{
			name:         "eventId list start, limit 5, page 20 (out of range)",
			args:         args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "06ced2b3-9b5a-4f11-9d55-efdccf650aaa", limit: "5", page: "20"},
			expectedCode: http.StatusOK,
			expectedLen:  0,
		}, {
			name:           "eventId list middle, limit 5",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3", limit: "5"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"ef070f21-85ca-4db4-a0e5-d01746372437", "1b2f0f34-0bf8-48d7-878d-72a169c6dfb4", "9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3", "620573db-dc13-439b-a3ed-7cec35989b6b", "3bad5888-f4af-47f3-a710-44703e0f595f"},
		},
		{
			name:           "eventId list middle, limit 5, page 2",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3", limit: "5", page: "2"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"73b5426e-73de-4566-b8e1-6c9355664ee6", "630f30c6-35b6-40b6-8e20-96651bd5d6ca", "92e5fb59-bb19-4588-a043-ad81f5e900ad", "3f88eecc-7e79-4bf1-9e90-0c101783e4d9", "ee25ba70-0d19-43a6-87f1-1428d6f57607"},
		},
		{
			name:           "eventId list middle, limit 5, page -2",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3", limit: "5", page: "-2"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"1acf8be9-65da-4180-a22b-8a3123dbb62c", "99bc053f-ac1b-45c1-a100-fa79fa26d6bd", "01a09b63-5f98-4ac1-9c6e-96119bb5009f", "5f2a28bb-8864-4bb1-9a3e-7d7b01716401", "69c34b34-c49b-4167-81f4-d62e69f571e6"},
		},
		{
			name:           "eventId list end, limit 5",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "7390e294-cd17-4b25-b132-6e3d4c61911c", limit: "5"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"bf674053-019b-4ad4-94e2-ae80bf5fc876", "85009409-e57e-454f-a10e-5e93769831e1", "9868953a-7acc-41dd-8ab6-3b0813812d66", "2aad03f8-ce99-4000-8cb1-71d99429cb6a", "7390e294-cd17-4b25-b132-6e3d4c61911c"},
		},
		{
			name:         "eventId list end, limit 5, page 2 (out of range)",
			args:         args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "7390e294-cd17-4b25-b132-6e3d4c61911c", limit: "5", page: "2"},
			expectedCode: http.StatusOK,
			expectedLen:  0,
		},
		{
			name:           "eventId list end, limit 5, page -2",
			args:           args{userId: "31b20726-b870-47ba-bbcd-372b38527c89", eventId: "7390e294-cd17-4b25-b132-6e3d4c61911c", limit: "5", page: "-2"},
			expectedCode:   http.StatusOK,
			expectedLen:    5,
			expectedEvents: []string{"a6f056d7-ed72-4d5d-8a53-d57532fff92f", "5d4f112d-ac08-46c7-af54-f0d26782850e", "b8060093-d3d3-4c28-b8d3-af40ed7d6e69", "b25c18f7-0ce7-45fa-bae7-e3de1ce74970", "22e197ca-8473-4d0b-bff3-978449e9aed2"},
		},
	}

	s := newServer(":1234", db)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			query := url.Values{
				"userId":  {test.args.userId},
				"eventId": {test.args.eventId},
				"limit":   {test.args.limit},
				"page":    {test.args.page},
			}

			req, err := http.NewRequest("GET", "/events?"+query.Encode(), nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.eventsHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, test.expectedCode)
			}

			var events []event
			if rr.Code == http.StatusOK {
				if err := json.NewDecoder(rr.Body).Decode(&events); err != nil {
					t.Error("error while parsing event result ", err)
				}
			}

			if l := len(events); l != test.expectedLen {
				t.Errorf("expected %d events but got %d", test.expectedLen, l)
			}

			if len(test.expectedEvents) > 0 {
				for i, e := range events {
					if expected := test.expectedEvents[i]; expected != e.ID {
						t.Errorf("expected eventId %s but got %s", expected, e.ID)
					}
				}
			}
		})
	}
}
