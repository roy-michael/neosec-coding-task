package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
)

type (

	// Timestamp for reading and writing the sample event time format
	Timestamp time.Time

	// event is used as the response payload
	event struct {
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
	}

	// inputEvent used for reading the input file
	inputEvent struct {
		event
		RequestContentType  string   `json:"request_content_type"`
		RequestSize         int      `json:"request_size"`
		ResponseContentType string   `json:"response_content_type"`
		ResponseSize        int      `json:"response_size"`
		AuthType            string   `json:"auth_type"`
		AttributesName      []string `json:"attributes.name"`
		AttributesIn        []string `json:"attributes.in"`
		AttributesPartOf    []string `json:"attributes.part_of"`
		AttributesValue     []string `json:"attributes.value"`
		AttributesValueType []string `json:"attributes.value_type"`
	}
)

// json unmarshalling only the inner `event` and not the whole `inputEvent`
func (e inputEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.event)
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

func (t Timestamp) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat))
	return []byte(fmt.Sprintf(`"%s"`, time.Time(t).AppendFormat(b, timeFormat))), nil
}

func (t Timestamp) String() string {
	return time.Time(t).String()
}
