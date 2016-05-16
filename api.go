package main

import (
    "fmt"
    "net/http"
    "github.com/davecgh/go-spew/spew"
    "github.com/twinj/uuid"
    "encoding/json"
    "errors"
    "net/url"
    "strconv"
    "github.com/jasonlvhit/gocron"
)

func monitorFlightAPI(w http.ResponseWriter, r *http.Request) {
    user, err := initAPI(w, r)
    if err != nil {
        return
    }
    
    //originCode string, destinationCode string, departureDate string, flightNumber string
    dcode := r.FormValue("departing_airport_code")
    acode := r.FormValue("arrival_airport_code")
    ddate := r.FormValue("departure_date")
    flightNumber := r.FormValue("flight_number")
    interval := r.FormValue("interval")
    
    /*
    layout := "2006-01-02T15:04:05.000Z"
    str := "2014-11-12T11:45:26.371Z"
    t, err := time.Parse(layout, str) */
    
    //monitorFlight(user.UserIdentifier, dcode, acode, ddate, flightNumber)
    fmt.Fprintf(w, "{\"monitor\":\"success\"}")
    go kickoffCron(user.UserIdentifier, dcode, acode, ddate, flightNumber, interval)
}

func kickoffCron(userIdentifier string, dcode string, acode string, ddate string, flightNumber string, interval string) {
    i,_ := strconv.Atoi(interval)
    gocron.Every(uint64(i)).Minutes().Do(monitorFlight, userIdentifier, dcode, acode, ddate, flightNumber)
    <- gocron.Start()
}

func getCookieAPI(w http.ResponseWriter, r *http.Request) {
    spew.Dump(r)
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    if session.Values["cookie"] != nil {
        cookies := session.Values["cookie"].(string)
        encodedCookies := encodeB64(cookies)
        fmt.Fprintf(w, "{\"cookie\":\"" + encodedCookies + "\"}")
    } else {
        fmt.Fprintf(w, returnJsonError("no cookie"))
    }
}

func loginAPI(w http.ResponseWriter, r *http.Request) {
    user := &User{}

    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    accountID := session.Values["accountid"]
    if accountID == nil {
        fmt.Println("generating new accountid")
        u := uuid.NewV4()
        session.Values["accountid"] = u.String()
        user = &User{UserIdentifier: u.String()}
        db.Create(&user)
    } else {
        db.Table("users").Where("user_identifier = ?", accountID).Scan(&user)

        if user.UserIdentifier == "" {
            // No valid session 
            fmt.Fprintf(w, returnJsonError("no valid account found"))
        }
    }
    
    decoder := json.NewDecoder(r.Body)
    var creds Credentials
    err = decoder.Decode(&creds)
    if err != nil {
        fmt.Printf("error in json decoder %v", err)
        fmt.Fprintf(w, returnJsonError("error in parsing request"))
        return
    }
    
    url, _ := url.Parse("https://employeerespss.coair.com/employeeres/")

    fmt.Println("logging in")
    client, err = login(creds.Username, creds.Password)
    if err != nil {
        fmt.Println(err)
        fmt.Fprintf(w, returnJsonError("login failed"))
        return
    }
    
    user.Credentials = creds
    db.Save(&user)

    cookieString := ""
    cookieCount := len(client.Jar.Cookies(url))
    for i, cookie := range client.Jar.Cookies(url) {
        if i < cookieCount {
            cookieString = cookieString + cookie.String() + ";"
        } else {
            cookieString = cookieString + cookie.String()
        }
    }
    
    fmt.Println(session.Values["cookie"])
    fmt.Println(session.Values["accountid"])
    
    session.Values["cookie"] = cookieString
    session.Save(r, w)
    
    fmt.Fprintf(w, "{\"login\":\"success\"}")
}

func initAPI(w http.ResponseWriter, r *http.Request) (User, error) {
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return User{}, errors.New("Failed to get session-name")
    }

    if session.Values["cookie"] != nil {
        restoreCookieJar(session, "https://employeerespss.coair.com/employeeres/")
    } else {
        http.Error(w, "test", 403)
        return User{}, errors.New("Failed to restore cookies")
    }
    
    userIdentifier := session.Values["accountid"].(string)
    user := User{}
    db.Table("users").Where("user_identifier = ?", userIdentifier).Scan(&user)

    if user.UserIdentifier == "" {
        // No valid session 
        fmt.Fprintf(w, returnJsonError("no valid account found"))
        return User{}, errors.New("No valid account found")
    }
    
    return user, nil
}

func printFlightsAPI(w http.ResponseWriter, r *http.Request) {
    user, err := initAPI(w,r)
    if err != nil {
        return
    }
    spew.Dump(user)
    dcode := r.FormValue("departing_airport_code")
    acode := r.FormValue("arrival_airport_code")
    ddate := r.FormValue("departure_date")
    
    aFlights := []string{}
    flights,_ := searchFlights(acode, dcode, ddate)
    for _,flight := range flights {
        aFlights = append(aFlights, printFlight(flight))
    }
    
    jsonResponse, err := json.Marshal(aFlights)
    w.Write(jsonResponse)
}

func searchFlightsAPI(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    if session.Values["cookie"] != nil {
        restoreCookieJar(session, "https://employeerespss.coair.com/employeeres/")
    } else {
        http.Error(w, "test", 403)
    }
    
    userIdentifier := session.Values["accountid"].(string)
    user := User{}
    db.Table("users").Where("user_identifier = ?", userIdentifier).Scan(&user)

    if user.UserIdentifier == "" {
        // No valid session 
        fmt.Fprintf(w, returnJsonError("no valid account found"))
        return
    }
    
    decoder := json.NewDecoder(r.Body)
    var t FlightSearch
    err = decoder.Decode(&t)
    if err != nil {
        fmt.Printf("error in json decoder %v", err)
        fmt.Fprintf(w, returnJsonError("error in parsing request"))
        return
    }
    
    if t.ArrivalAirportCode == "" || t.DepartingAirportCode == "" || t.DepartureDate == "" {
        fmt.Println("Missing required parameters. Requires arrival_airport_code, departing_airport_code, departure_date")
        fmt.Fprintf(w, returnJsonError("Missing required parameters. Requires arrival_airport_code, departing_airport_code, departure_date"))
        return
    }
    
    flights, err := searchFlights(t.ArrivalAirportCode, t.DepartingAirportCode, t.DepartureDate)
    if err != nil {
        fmt.Println(err)
        fmt.Fprintf(w, returnJsonError("%v"), err)
        return
    }
    
    jsonResponse, err := json.Marshal(flights)
    if err != nil {
        fmt.Printf("error in marshal %v", err)
        fmt.Fprintf(w, returnJsonError("error in marshal"))
    }
    
    fmt.Printf("Found %d flights\n", len(flights))
    
    search := Search{ArrivalAirport: t.ArrivalAirportCode,
        DepartureAirport: t.DepartingAirportCode,
        DepartureDate: t.DepartureDate, 
        Results: flights}
    user.Searches = append(user.Searches, search)
    db.Save(&user)
    
    w.Write(jsonResponse)
    //spew.Dump(flights)
}