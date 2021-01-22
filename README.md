# neosec Coding Task
Implementation of an API server for processing and reading application's access logs

## Getting Stated
### Building
Note that the Go (golang) sdk must be installed for building the server binary, on the build machine.  
In order to build the binary issue the following command from project directory:
```
> go build
```
the `neosec` binary will be built in the current working directory

### Running
The binary does not need any special dependencies for runtime, and it's default parameters can be customized with the following:
#### Parameters:
```
Usage of neosec:
  -addr string
        address to listen on (default ":8888")
  -sampleFile string
        sample file path (default "testdata/events-sample.json")
```
### Testing
Using the go sdk run:
```
> go test -v
```

## Design

### The `events` Endpoint
The server exposes the `/events` endpoint for querying and reading the events for a specific user id.  
For a successful query result, the server will return a JSON formatted list of event objects, sorted by their timeline

#### Query Parameters
The accepted query parameters are:
* `userId` - required; the id of the user to query events for
* `eventId` - the id of the event. the `eventId` is also used as an anchor when combined with the `limit` and `page` parameters
* `limit` - a limit for the returned events result, list size
* `page` - used for events result list pagination and works in relation to the `eventId` anchor.  
  accepted values are positive or negative numbers

#### Error Handling
* 400/Bad Request - returned when the user id query param is missing
* 405/Method Not Allowed - returned for non http `GET` method requests
* 503/Service Unavailable - returned for unexpected errors while returning the query results 

### Model
#### The `Event` Type Model
The event entity captures the information required for the identification of the event, its source, target and outcome, formatted as follows:
```json
{
  "id": "<string>",
  "timestamp": "<yyyy-mm-dd hh:MM:ss.SSS string formatted time>",
  "caller_ip": "<string>",
  "url": "<string>",
  "method": "<string>",
  "call_path": "<string>",
  "server_url": "<string>",
  "status_code": "<integer>",
  "user_id": "<string>",
  "service_name": "<string>",
  "endpoint_path": "<string>",
  "endpoint_id": "<string>"
}
```
### Examples
#### Query Events by Event ID
This is an example for querying events for a user with id `31b20726-b870-47ba-bbcd-372b38527c89` anchored by an event with id `9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3`
 while limiting the returned result list to 5 entries and moving to the second page on the result list
##### Request
```
> curl "http://localhost:8888/events?userId=31b20726-b870-47ba-bbcd-372b38527c89&eventId=9d5a1220-7fa0-4e8b-a04a-e47e630e5ef3&limit=5&page=2"
```
##### Response
```json
[
  {
    "id": "73b5426e-73de-4566-b8e1-6c9355664ee6",
    "timestamp": "2021-01-13 00:25:32.472",
    "caller_ip": "0.0.0.0",
    "url": "//0.0.0.0:7000/v1/server",
    "method": "GET",
    "call_path": "/v1/server",
    "server_url": "//0.0.0.0:7000",
    "status_code": 200,
    "user_id": "31b20726-b870-47ba-bbcd-372b38527c89",
    "service_name": "0.0.0.0:7000",
    "endpoint_path": "/v1/server",
    "endpoint_id": "6b9ca9a8-7e21-2241-67ac-42483fc6b86b"
  },
  {
    "id": "630f30c6-35b6-40b6-8e20-96651bd5d6ca",
    "timestamp": "2021-01-13 00:25:32.483",
    "caller_ip": "0.0.0.0",
    "url": "//0.0.0.0:7000/v1/events/2fbfdbd7-f27e-45cf-b45b-ccc99014e6de",
    "method": "GET",
    "call_path": "/v1/events/2fbfdbd7-f27e-45cf-b45b-ccc99014e6de",
    "server_url": "//0.0.0.0:7000",
    "status_code": 200,
    "user_id": "31b20726-b870-47ba-bbcd-372b38527c89",
    "service_name": "0.0.0.0:7000",
    "endpoint_path": "/v1/events/{event_id}",
    "endpoint_id": "24c7d2a2-e706-134b-6411-607c46d695a8"
  },
  {
    "id": "92e5fb59-bb19-4588-a043-ad81f5e900ad",
    "timestamp": "2021-01-13 00:25:32.661",
    "caller_ip": "0.0.0.0",
    "url": "//0.0.0.0:7000/v1/user/profile",
    "method": "GET",
    "call_path": "/v1/user/profile",
    "server_url": "//0.0.0.0:7000",
    "status_code": 200,
    "user_id": "31b20726-b870-47ba-bbcd-372b38527c89",
    "service_name": "0.0.0.0:7000",
    "endpoint_path": "/v1/user/profile",
    "endpoint_id": "60f6b253-6dd9-b01b-a860-103f7567b7a4"
  },
  {
    "id": "3f88eecc-7e79-4bf1-9e90-0c101783e4d9",
    "timestamp": "2021-01-13 00:25:35.721",
    "caller_ip": "0.0.0.0",
    "url": "//0.0.0.0:7000/v1/events/2fbfdbd7-f27e-45cf-b45b-ccc99014e6de/enter",
    "method": "POST",
    "call_path": "/v1/events/2fbfdbd7-f27e-45cf-b45b-ccc99014e6de/enter",
    "server_url": "//0.0.0.0:7000",
    "status_code": 200,
    "user_id": "31b20726-b870-47ba-bbcd-372b38527c89",
    "service_name": "0.0.0.0:7000",
    "endpoint_path": "/v1/events/{event_id}/enter",
    "endpoint_id": "a0f53f5c-01f0-b7fd-90d8-da395954e2ec"
  },
  {
    "id": "ee25ba70-0d19-43a6-87f1-1428d6f57607",
    "timestamp": "2021-01-13 00:25:55.636",
    "caller_ip": "45.72.212.77",
    "url": "https://api.app.staging.env.paypronto.events/v1/user/profile",
    "method": "GET",
    "call_path": "/v1/user/profile",
    "server_url": "https://api.app.staging.env.paypronto.events",
    "status_code": 200,
    "user_id": "31b20726-b870-47ba-bbcd-372b38527c89",
    "service_name": "api.app.staging.env.paypronto.events",
    "endpoint_path": "/v1/user/profile",
    "endpoint_id": "8c0fc451-39ad-82bb-6179-8529d1119b9c"
  }
]
```