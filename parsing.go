package main

import (
    "net/url"
    "io/ioutil"
    "regexp"
    "net/http"
    "fmt"
    "github.com/Jeffail/gabs"
    "errors"
    "strings"
    //"os"
    "github.com/twinj/uuid"
    "encoding/json"
    "github.com/davecgh/go-spew/spew"
    "strconv"
    "github.com/PuerkitoBio/goquery"
)

func parseSearchResponse(data string) ([]*Flight, error) {
    jsonData := ""
    myFlights := []*Flight{}
    
    re1, err := regexp.Compile(`var\sfd\s=\s'([^']+)';`) // Prepare our regex
    if err != nil {
        fmt.Printf("error in regex %v", err)
    }
    result := re1.FindAllStringSubmatch(data, -1)
    for i := range result {
        jsonData = result[i][1]
        //fmt.Printf("viewstate: %s", result[i][1])
    }
    
    jsonParsed, err := gabs.ParseJSON([]byte(jsonData))
    if err != nil {
        fmt.Printf("error parsing json: %v", err)
        return nil, errors.New("error parsing json in parseSearchResponse")
    }
    
    flights, err := jsonParsed.Path("ar.r.s").Children()
    if err != nil {
        fmt.Printf("error in Children() %v", err)
        fmt.Println(data)
        return nil, errors.New("error parsing json children in parseSearchResponse")
    }
    
    spew.Dump("")

    for _, flight := range flights {
        tmp := flight.Data().([]interface{})
        data := tmp[0].(map[string]interface{})
        
        flightRecord := new(Flight)
        flightRecord.UUID = uuid.NewV4().String()
        flightRecord.ArrivalTime = data["at"].(string)
        flightRecord.ArrivalDate = data["ad"].(string)
        flightRecord.DepartureDate = data["dd"].(string)
        flightRecord.DepartureTime = data["dt"].(string)
        flightRecord.FlightNumber = data["fn"].(string)
        
        flightRecord.Authorized = append(flightRecord.Authorized, parseSeats("authorized", data["au"].(map[string]interface{})))
        flightRecord.Available = append(flightRecord.Available, parseSeats("available", data["av"].(map[string]interface{})))
        flightRecord.Booked = append(flightRecord.Booked, parseSeats("booked", data["bo"].(map[string]interface{})))
        flightRecord.Capacity = append(flightRecord.Capacity, parseSeats("capacity", data["ca"].(map[string]interface{})))

        flightRecord.Destination = parseAirport("destination", data["d"].(map[string]interface{}))
        flightRecord.Origination = parseAirport("origin", data["o"].(map[string]interface{}))
        flightRecord.Duration = *new(TravelTime)
        for key, value := range data["tt"].(map[string]interface{}) {
            time, err := strconv.Atoi(value.(string))
            if err != nil {
                fmt.Printf("error converting string to int: %v\n", err)
            }
            switch key {
                case "h": 
                    flightRecord.Duration.Hour = time
                case "m":
                    flightRecord.Duration.Minute = time
                case "tm":
                    flightRecord.Duration.Total = time
            }
        }
        myFlights = append(myFlights, flightRecord)
    }
    
    getPassriderList(myFlights)
    return myFlights, nil
}



func testSearchParser() {
    file, err := ioutil.ReadFile("response.txt")
    if err != nil {
        fmt.Printf("error in reading file response.txt %v", err)
    }
    flights, err := parseSearchResponse(string(file))
    if err != nil {
        fmt.Println(err)
    }
    
    for _, flight := range flights {
        fmt.Println(flight.DepartureTime)
    }
    
    str, err := json.Marshal(flights)
    if err != nil {
        fmt.Printf("error in marshal %v", err)
    }
    spew.Dump(string(str))
}

func parseLoginForm(data string) string {
    viewstate := ""
    eventvalidation := ""
    
    re1, err := regexp.Compile(`__VIEWSTATE"\svalue="([^"]+)`) // Prepare our regex
    if err != nil {
        fmt.Printf("error in regex %v", err)
    }
    result := re1.FindAllStringSubmatch(data, -1)
    for i := range result {
        viewstate = url.QueryEscape(result[i][1])
        //fmt.Printf("viewstate: %s", result[i][1])
    }
    
    re1, err = regexp.Compile(`__EVENTVALIDATION"\svalue="([^"]+)`)
    result = re1.FindAllStringSubmatch(data, -1)
    for i := range result {
        eventvalidation = url.QueryEscape(result[i][1])
        //fmt.Printf("eventvalidation: %s", result[i][1])
    }
    
    return "__EVENTARGUMENT=&__VIEWSTATE=" + viewstate + "&__EVENTVALIDATION=" + eventvalidation
}

func parseAirport(airportType string, data map[string]interface{}) Airport {
    airport := new(Airport)
    airport.AirportType = airportType
    for key, value := range data {
        switch key {
            case "ac":
                airport.Code = value.(string)
            case "an":
                airport.Name = value.(string)
            case "cn":
                airport.City = value.(string)
            case "cc":
                airport.Country = value.(string)
            default:
                fmt.Printf("Did not recognize airport key of %s in: %v",key, data["au"])
        }
    }
    
    return *airport
}

func parseSeats(seatType string, data map[string]interface{}) Seats {
    authorizedSeats := new(Seats)
    authorizedSeats.SeatType = seatType
    for key, value := range data {
        seatCount, err := strconv.Atoi(value.(string))
        if err != nil {
            fmt.Printf("error converting string to int: %v\n", err)
        }
        switch key {
            case "b":
                authorizedSeats.Business = seatCount
            case "f":
                authorizedSeats.First = seatCount
            case "c":
                authorizedSeats.Coach = seatCount
            case "t":
                authorizedSeats.Total = seatCount
            default:
                fmt.Printf("Did not recognize seat key of %s in: %v",key, data["au"])
        }
    }
    
    return *authorizedSeats
}

func parsePostdata(data string) url.Values {
    postData := url.Values{}

    params := strings.Split(data, "&")
    for i := range params {
        param := strings.Split(params[i], "=")
        //fmt.Printf("adding %s=%s\r\n", param[0], param[1])
        value,_ := url.QueryUnescape(param[1])
        postData.Set(param[0], value)
    }
    
    return postData
}

func parsePassrider(s *goquery.Selection, position int) Passrider {
    passrider := Passrider{}
    passrider.Position = position
    
    s.Find("div").Each(func(count int, m *goquery.Selection) {
        switch count {
            case 0:
                passrider.Class = m.Text()
            case 1:
                passrider.BoardingDate = m.Text()
            case 2:
                passrider.Seats, _ = strconv.Atoi(m.Text())
            case 3:
                passrider.Cabin = m.Text()
            case 4:
                passrider.Name = m.Text()        
        }
    })
    
    return passrider
}

func parsePassriderList(response *http.Response) []Passrider {
    passRiders := []Passrider{}
  
    doc, err := goquery.NewDocumentFromResponse(response)
    if err != nil {
        fmt.Printf("error in goquery %v", err)
    }
        
    position := 1
    doc.Find("div.passRiderMain").Children().Each(func (i int, s *goquery.Selection) {
            class, _ := s.Attr("class")
            if class == "divRowPassRider" || class == "divRowAltPassRider" {
                passrider := parsePassrider(s,position)
                passRiders = append(passRiders, passrider)
                position++
            }
    })
   
   return passRiders
}

func parseSingleFlightDetails(flightDetails string, flight *Flight) {
    r := regexp.MustCompile(`\s|\n`)
    clean := r.ReplaceAllString(flightDetails, "_")
    r = regexp.MustCompile(`_+`)
    clean = r.ReplaceAllString(clean, " ")
    test := strings.Split(clean," ")
    duration := TravelTime{}
    for index, item := range test {
        switch index {
            case 2:
                //fmt.Println("Departs: " + item)
                flight.Origination = Airport{Code: item[0:3]}
                flight.DepartureTime = item[3:len(item)]
            case 4:
                //fmt.Println("Arrives: " + item)
                flight.Destination = Airport{Code: item[0:3]}
                flight.ArrivalTime = item[3:len(item)]
            case 7:
                //fmt.Println("Hour: " + item)
                duration.Hour,_ = strconv.Atoi(item)
            case 9:
                //fmt.Println("Minute: " + item)
                duration.Minute,_ = strconv.Atoi(item)
            case 20:
                //fmt.Println("Flight: " + item)
                flight.FlightNumber = item[2:len(item)]
        }
    }
    flight.Duration = duration
}

func parseSingleFlightSeats(seatType string, data string) Seats {
    splitClasses := strings.Split(data, "/")
    seats := Seats{}
    seats.SeatType = seatType
    seats.First,_ = strconv.Atoi(splitClasses[0])
    seats.Business,_ = strconv.Atoi(splitClasses[1])
    seats.Coach,_ = strconv.Atoi(splitClasses[2])
    
    return seats
}

func parseSingleFlightSeatTotals(s *goquery.Selection) {
    s.Children().Each(func (index int, sel *goquery.Selection) {
        value := strings.TrimSpace(sel.Text())
        switch index {
            case 1:
                fmt.Println("Avail Total: " + value)
            case 2:
                fmt.Println("Capacity Total: " + value)
            case 3:
                fmt.Println("Auth Total: " + value)
            case 4:
                fmt.Println("Booked Total: " + value)
        }
    })
}

func parseSingleFlightSeatNumbers(s *goquery.Selection, flight *Flight) {
    s.Children().Each(func (index int, sel *goquery.Selection) {
        value := strings.TrimSpace(sel.Text())
        switch index {
            case 1:
                //fmt.Println("Avail: " + value)
                flight.Available = append(flight.Available, parseSingleFlightSeats("available", value))
            case 2:
                //fmt.Println("Capacity: " + value)
                flight.Capacity = append(flight.Capacity, parseSingleFlightSeats("capacity", value))
            case 3:
                //fmt.Println("Auth: " + value)
                flight.Authorized = append(flight.Authorized, parseSingleFlightSeats("authorized", value))
            case 4:
                //fmt.Println("Booked: " + value)
                flight.Booked = append(flight.Booked, parseSingleFlightSeats("booked", value))
        }
    })
}

func parseSingleFlight(data string, flight *Flight) {
    //f, _ := os.Open("single-flight-response.html")
    f := strings.NewReader(data)
    doc, _ := goquery.NewDocumentFromReader(f)
    flightDetails := ""
    doc.Find("tBody").Children().Each(func (i int, s *goquery.Selection) {
        switch i {
            case 0:
                flightDetails = s.Text()
            case 2:
                flightDetails = flightDetails + s.Text()
            case 5:
                // parse seat numbers
                parseSingleFlightSeatNumbers(s, flight)
            //case 6:
                // seat totals
                //parseSingleFlightSeatTotals(s)
        }
    })
    
    parseSingleFlightDetails(flightDetails, flight)
}