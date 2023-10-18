package main

import (
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/gorilla/mux"

)

// endpoint request structs

type ListWorkingHoursRequest struct {
	PayloadId string `json:"payloadId"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ListBlockHoursRequest struct {
	PayloadId string `json:"payloadId"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ListBookingRequest struct {
	PayloadId string `json:"payloadId"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

// entities

type Duration struct {
	Seconds int64 `json:"seconds"`
}

type Payload struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type WorkingHour struct {
	Id         string    `json:"id"`
	PayloadId string    `json:"resource_id"`      
	Quantity   int64     `json:"quantity"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

type BlockHour struct {
	Id         string    `json:"id"`
	PayloadId string    `json:"resource_id"`      
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

type Booking struct {
	Id         string    `json:"id"`
	PayloadId string    `json:"resource_id"`     
	Quantity   int64     `json:"quantity"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

type Slot struct {                                  //for output
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// helper functions

func TimeToString(tm time.Time) string {
	return tm.Format(time.RFC3339)
}

func StringToTime(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func main() {

	// create router
	router := mux.NewRouter()

	// create /availability endpoint
	router.HandleFunc("/availability", availabilityHandler).Methods("GET")

	// Run server
	http.ListenAndServe(":8000", router)

	// - payloadId [Required]: ID of the payload
	// - date [Required]: date in YYYY-MM-DD format
	// - duration [Required]: time duration in minutes (e.g., 30, 60, 120)
	// - quantity [Required]:  capacity to reserve

}





// creating function for API Call with param endpoint and payload (object)
func apiCall(endpoint string, payload map[string]interface{}) string {

	url := "http://api.internship.appointy.com:8000/space-payload/v1"
	method := "GET"

	newurl := url + endpoint

	// Adding get parameters to url
	if payload != nil {
		newurl = newurl + "?"
		for key, value := range payload {
			newurl = newurl + key + "=" + value.(string) + "&"
		}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, newurl, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOiIyMDIzLTA4LTEwVDAwOjAwOjAwWiIsInVzZXJfaWQiOjY4fQ.LXzG86TTgpjXj1ScR6lCxfgY66k3iMguexIgYVBwhJE")
	req.Header.Add("Content-Type", "application/json")

	// sending request
	res, err := client.Do(req)
	if err != nil {           
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println(endpoint, " API Call Successful")
	}

	return string(body)
}



func availabilityHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	// get query parameters
	queryParams := r.URL.Query()

	// get payloadId
	payloadIdParam := queryParams.Get("payloadId")  
	dateParam := queryParams.Get("date")
	durationParam := queryParams.Get("duration")
	quantityParam := queryParams.Get("quantity")

	if payloadIdParam == "" || dateParam == "" || durationParam == "" || quantityParam == "" {   
		fmt.Println("Missing Parameters")
		return
	}

	inputParam := map[string]interface{}{
		"payloadId": payloadIdParam,
		"date":       dateParam,
		"duration":   durationParam,
		"quantity":   quantityParam,
	}
    // Example inputParam
	// inputParam := map[string]interface{}{
	// 	"payloadId": "pyl_2",
	// 	"date":       "2023-08-05",
	// 	"duration":   "30",
	// 	"quantity":   4,
	// }

	// Create startTime and EndTime in format YYYY-MM-DDTHH:mm:ss.sssZ from inputParam
	payloadId := inputParam["payloadId"].(string)
	startTime := inputParam["date"].(string) + "T00:00:00Z"
	endTime := inputParam["date"].(string) + "T23:59:00Z"
	// quantity is number of slots to be booked
	quantityInt, _ := strconv.Atoi(inputParam["quantity"].(string))
	quantity := int64(quantityInt)

	// declare payload
	payload := map[string]interface{}{
		"payload": payloadId,
		"startTime":  startTime,
		"endTime":    endTime,
	}

	workinghours := apiCall("/working-hours", payload)
	blockhours := apiCall("/block-hours", payload)
	booking := apiCall("/bookings", payload)

	// change workinghours from string to Maps of WorkingHour struct
	var workinghoursMap []WorkingHour
	json.Unmarshal([]byte(workinghours), &workinghoursMap)

	var blockhoursMap []BlockHour
	json.Unmarshal([]byte(blockhours), &blockhoursMap)

	var bookingMap []Booking
	json.Unmarshal([]byte(booking), &bookingMap)

	var availableSlots []Slot

	// check for availability of resource on given date for given duration
	for i := 0; i < len(workinghoursMap); i++ {
		// fmt.Println(workinghoursMap[i].StartTime)
		// fmt.Println(workinghoursMap[i].EndTime)

		// convert string to time
		startTime, _ := StringToTime(workinghoursMap[i].StartTime)
		endTime, _ := StringToTime(workinghoursMap[i].EndTime)

		// convert duration from string to int64
		duration, _ := time.ParseDuration(inputParam["duration"].(string) + "m")

		// fmt.Println("Working Hours: ", i+1)
		// fmt.Println("Start Time: ", startTime, "End Time: ", endTime)
		// fmt.Println("Duration: ", duration)

		// check all available slots in workinghours

		for j := startTime; j.Before(endTime); j = j.Add(duration) {
			var available bool = true

			// fmt.Println("Slot: ", j, "to", j.Add(duration))

			// check if j is in blockhours
			for k := 0; k < len(blockhoursMap); k++ {
				// convert string to time
				blockStartTime, _ := StringToTime(blockhoursMap[k].StartTime)
				blockEndTime, _ := StringToTime(blockhoursMap[k].EndTime)

				if (j.Equal(blockStartTime) || j.After(blockStartTime)) && (j.Before(blockEndTime) || j.Before(blockEndTime)) {

					available = false
					break
				}
			}

			// check if j is in appointment hours and quantity is available
			for l := 0; l < len(bookingMap); l++ {
				// convert string to time
				bookingStartTime, _ := StringToTime(bookingMap[l].StartTime)
				bookingEndTime, _ := StringToTime(bookingMap[l].EndTime)

				if (j.Equal(bookingStartTime) || j.After(bookingStartTime)) && (j.Before(bookingEndTime) || j.Before(bookingEndTime)) {

					workingHourQuantity := workinghoursMap[i].Quantity
					bookingQuantity := bookingMap[l].Quantity

					// change into int64 and check if quantity is available
					if (workingHourQuantity-bookingQuantity < quantity) {
						available = false
						break
					}

				}
			}

			if available {
				fmt.Println("available")
				availableSlots = append(availableSlots, Slot{TimeToString(j), TimeToString(j.Add(duration))})
			} else {
				fmt.Println("blocked")
			}

		}

		fmt.Println("")

	}

	fmt.Println("Available Slots: ", availableSlots)
	// Convert availableSlots to json
	availableSlotsJson, _ := json.Marshal(availableSlots)
	fmt.Println("Available Slots Json: ", string(availableSlotsJson))

	// write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	elapsed := time.Since(start)
	elapsedSeconds := int64(elapsed.Seconds())

	Result := map[string]interface{}{
		"available_slots": availableSlots,
		"time_taken":      elapsedSeconds,
	}

	// finalResult := json.NewEncoder(w).Encode(Result)
	
	finalResult, err := json.Marshal(Result)
	if err != nil {
		fmt.Println(err)
	}

	w.Write(finalResult)
}