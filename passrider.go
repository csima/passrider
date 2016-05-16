
package main

import (
    "fmt"
    "os"
    "errors"
    "time"
    "strconv"
    "github.com/gorilla/mux"
    //"github.com/jasonlvhit/gocron"
    "github.com/gorilla/sessions"
    "github.com/codegangsta/negroni"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "net/http"
    "net/http/cookiejar"
    "strings"
    "github.com/Jeffail/gabs"
    //"github.com/davecgh/go-spew/spew"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

var client *http.Client
var db *gorm.DB
type PostType int

const ( 
   Normal PostType = iota 
   JSON // 2 (i.e. 1 << 1)
   Other // 4 (i.e 1 << 2)
)


func main() {
    dbInit()
    restAPIServer()
}

func restAPIServer() {
    fmt.Printf("%s:%s\n", os.Getenv("IP"), os.Getenv("PORT"))
    router := mux.NewRouter()

    n := negroni.Classic()

    router.HandleFunc("/search", searchFlightsAPI)
    router.HandleFunc("/login", loginAPI)
    router.HandleFunc("/cookie", getCookieAPI)
    router.HandleFunc("/printFlights", printFlightsAPI)
    router.HandleFunc("/monitorFlight", monitorFlightAPI)
    
    n.UseHandler(router)
    n.Run("0.0.0.0:8080")
}

func searchFlights(arrivalAirportCode string, departureAirportCode string, departureDate string) ([]*Flight, error) {
    jsonObj := gabs.New()
    flightSearch := FlightSearch{ArrivalAirportCode: arrivalAirportCode, DepartingAirportCode: departureAirportCode, DepartureDate: departureDate, DepartingTime: "0000", FlightType: "ow", MaximumConnections: "nonstop"}
    jsonObj.Set(flightSearch.ArrivalAirportCode, "search", "AC")
    jsonObj.Set(flightSearch.ArrivalAirportCode2, "search", "AC2")
    jsonObj.Set(flightSearch.ArrivalDisplayName, "search", "AN")
    jsonObj.Set(flightSearch.ArrivalDisplayName2, "search", "AN2")
    jsonObj.Set(flightSearch.DepartingAirportCode, "search", "DC")
    jsonObj.Set(flightSearch.DepartingAirportCode2, "search", "DC2")
    jsonObj.Set(flightSearch.DepartureDate, "search", "DD")
    jsonObj.Set(flightSearch.DepartureDisplayName, "search", "DN")
    jsonObj.Set(flightSearch.DepartureDisplayName2, "search", "DN2")
    jsonObj.Set(flightSearch.DepartingTime, "search", "DT")
    jsonObj.Set(flightSearch.EN, "search", "EN")
    jsonObj.Set(flightSearch.FlightType, "search", "FT")
    jsonObj.Set(flightSearch.MaximumConnections, "search", "M")
    jsonObj.Set(flightSearch.PT, "search", "PT")
    jsonObj.Set(flightSearch.QE, "search", "QE")
    jsonObj.Set(flightSearch.RD, "search", "RD")
    
    response := doPost("https://employeerespss.coair.com/employeeres/Ajax_Pages/FlightSearch.asmx/Simple", client, jsonObj.String(), JSON)
    if strings.Contains(response, "There was an error processing the request") == true {
        return nil, errors.New("Session has been logged out. Log back in")
    }
    if strings.Contains(response, "success") == false {
        return nil, errors.New("Flight search failed. Ensure you passed in the correct parameters")
    }
    response = doGet("https://employeerespss.coair.com/employeeres/flightlist.aspx", client)
    if strings.Contains(response, "var fd") == false || strings.Contains(response, "All flights have departed or no flights found for this request") == true {
        return nil, errors.New("All flights have departed or no flights found for this request")
    }
    
    flights, err := parseSearchResponse(response)
    if err != nil {
        return nil, err
    }
    
    return flights, nil
}

func getAirports(data string) {
    response := doPost("https://employeerespss.coair.com/employeeres/Ajax_Pages/FlightSearch.asmx/GetAirports", client, "{\"Q\":\"" + data + "\"}", JSON)
//    fmt.Println(response)

    jsonParsed, err := gabs.ParseJSON([]byte(response))
    if err != nil {
        fmt.Printf("error parsing json: %v", err)
    }

    children, _ := jsonParsed.S("d").Children()
    for _, child := range children {
        fun := child.Data().(map[string]interface{})
        fmt.Println(fun["DisplayName"])
    }
}

func login(username string, password string) (*http.Client, error) {
    fmt.Println("logging in user " + username)
    j, err := cookiejar.New(nil)
    if err != nil {
        panic(err)
    }
    
    client := &http.Client{Jar: j}
    response := doGet("https://employeerespss.coair.com/employeeres/PassRiderLogin.aspx", client)
    postdata := parseLoginForm(response)
    postdata = postdata + "&txtLogin=" + username + "&txtPsswd=" + password + "&__EVENTTARGET=btnSubmit"

    response = doPost("https://employeerespss.coair.com/employeeres/PassRiderLogin.aspx", client, postdata, Normal)
    if strings.Contains(response, "PassRiderTerms") == false {
        return nil, errors.New("Login failed")
    }
 
    response = doGet("https://employeerespss.coair.com/employeeres/PassRiderTerms.aspx", client)
    postdata = parseLoginForm(response)
    postdata = postdata + "&__EVENTTARGET=ctl00$MainContent$btnSubmit"
    response = doPost("https://employeerespss.coair.com/employeeres/PassRiderTerms.aspx", client, postdata, Normal)
    //fmt.Println(response)
    fmt.Println("logged in")
    return client, nil
}

func getPassriderList(myFlights []*Flight) {
    urls := make(map[string]string)
    for _, flight := range myFlights {
        url := "https://employeerespss.coair.com/employeeres/passriderlist.aspx?showposition=true&fltno=" + flight.FlightNumber + "&org=" + flight.Origination.Code + "&fltdate=" + flight.DepartureDate + "&dest=" + flight.Destination.Code
        urls[flight.UUID] = url
    }
    //fmt.Printf("url count: %d\n",  len(urls))
    
    responses := AsyncGet(urls)
    //fmt.Printf("response count: %d\n", len(responses))
    
    for _, value := range responses {
        for i, flight := range myFlights {
            if flight.UUID == value.Id {
                myFlights[i].Passriders = parsePassriderList(value.Resp)
            }
        }
    }   
    //spew.Dump(myFlights)
}

func getFlight(originCode string, destinationCode string, departureDate string, flightNumber string) (Flight, error) {
    finished := false
    response := ""
    requestCount := 0
    for finished == false {
        response = doGet("https://employeerespss.coair.com/employeeres/pbts.aspx?origin=" + originCode + "&destination=" + destinationCode + "&departDate=" + departureDate + "&marketingCode=UA&flightNumber=" + flightNumber, client)
        if strings.Contains(response, "Departs:") {
            finished = true
        }
        if requestCount >= 10 {
            // something went wrong
            //fmt.Println(response)
            fmt.Println("Unable to retrieve flight details from pbts.aspx")
            return Flight{}, errors.New("Unable to retrieve flight details from pbts.aspx")
        }
        
        if requestCount < 6 {
	        time.Sleep(time.Duration(1) * time.Second)
        }
        if requestCount > 6 {
	        time.Sleep(time.Duration(3) * time.Second)
        }

        requestCount++
    }
    
    flight := Flight{}
    flight.DepartureDate = departureDate
    parseSingleFlight(response, &flight)
    getPassriderList([]*Flight{&flight})
    return flight, nil
    //spew.Dump(flight)
}

func getFlightHistory(origCode string, destCode string, flightNumber string, monthYear string) {
    dateLayout := "1/2/2006"
    date, err := time.Parse(dateLayout, monthYear)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    for i := 1; i <= 7; i++ {
        date = date.AddDate(0,0,1)
        year, month, day := date.Date()
        dateString := fmt.Sprintf("%s/%s/%s", strconv.Itoa(int(month)), strconv.Itoa(day), strconv.Itoa(year))
        flight, err := getFlight(origCode, destCode, dateString, flightNumber)
        if err != nil {
            fmt.Printf("%s->%s %s(%s):%s\n", origCode,destCode, dateString, flightNumber, err)
            continue
        }
        
        printFlight(&flight)
        //fmt.Println(dateString)
    }
}

func monitorFlight(userID string, originCode string, destinationCode string, departureDate string, flightNumber string) {
    fmt.Printf("Monitoring flight: %s -> %s %s %s\n", originCode, destinationCode, departureDate, flightNumber)
    user := User{}
    creds := Credentials{}
    
    fmt.Printf("userid = %s\n", userID)
    db.Where(&User{UserIdentifier: userID}).First(&user)

    if user.UserIdentifier == "" {
        // No valid session 
        fmt.Println("no valid account found")
        return
    }
        
    flight, err := getFlight(originCode, destinationCode, departureDate, flightNumber)

    if err != nil {
        db.Model(&user).Related(&creds)
        fmt.Println("Re-logging in user %s", creds.Username)
        client, _ = login(creds.Username,creds.Password)
        flight, err = getFlight(originCode, destinationCode, departureDate, flightNumber)
        if err != nil {
            fmt.Println("Fatal error: Unable to monitor flight")
            return
        }
    } else {
        fmt.Println("Success")
    }
    
    monitor := Monitor{}
    db.Where(&Monitor{OriginationCode: originCode, DestinationCode: destinationCode, FlightNumber: flightNumber}).First(&monitor)
    if monitor.FlightNumber == "" { 
        fmt.Println("Creating new monitored flight")
        monitor = Monitor{OriginationCode: originCode, DestinationCode: destinationCode,
        FlightNumber: flightNumber}
    } else {
        fmt.Println("Add to existing monitored flight")
    }
    
    monitor.Flights = append(monitor.Flights, &flight)
    user.Monitored = append(user.Monitored, monitor)
    db.Save(&user)
}





