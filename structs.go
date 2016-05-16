package main

import (
    "github.com/jinzhu/gorm"
)

type Monitor struct {
    gorm.Model
    UserID int
    OriginationCode string
    DestinationCode string
    FlightNumber string
    Flights []*Flight
}

type User struct {
    gorm.Model
    UserIdentifier string 
    Credentials Credentials
    Searches []Search
    Monitored []Monitor
}

type Search struct {
    gorm.Model
    ArrivalAirport string
    DepartureAirport string
    DepartureDate string
    UserID int 
    Results []*Flight
}

type Passrider struct {
    gorm.Model
    FlightID int
    Class string
    BoardingDate string
    Seats int
    Cabin string
    Name string
    Position int
}

type Seats struct {
    SeatType string
    FlightID int
    First int
    Business int
    Coach int
    Total int
}

type TravelTime struct {
    FlightID int
    Hour int
    Minute int
    Total int
}

type FlightSearch struct {
    ArrivalAirportCode string `json:"arrival_airport_code"`
    ArrivalAirportCode2 string `json:"arrival_airport_code2"`
    ArrivalDisplayName string `json:"arrival_display_name"`
    ArrivalDisplayName2 string `json:"arrival_display_name2"`
    DepartingAirportCode string `json:"departing_airport_code"`
    DepartingAirportCode2 string `json:"departing_airport_code2"`
    DepartureDate string `json:"departure_date"`
    DepartureDisplayName string `json:"departure_display_name"`
    DepartureDisplayName2 string `json:"departure_display_name2"`
    DepartingTime string `json:"departing_time"`
    EN string `json:"en"`
    FlightType string `json:"flight_type"`
    MaximumConnections string `json:"maximum_connections"`
    PT string `json:"PT"`
    QE string `json:"QE"`
    RD string `json:"RD"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type Airport struct {
    FlightID int
    AirportType string
    Code string
    Name string
    Country string
    City string
}

type Flight struct {
    gorm.Model
    SearchID int
    MonitorID int
    UUID string `json:"uuid"`
    ArrivalDate string `json:"arrival_date"`
    ArrivalTime string `json:"arrival_time"`
    Authorized []Seats `json:"authorized"`
    Available []Seats `json:"available"`
    Booked []Seats `json:"booked"`
    Capacity []Seats `json:"capacity"`
    Destination Airport `json:"destination"`
    DepartureDate string `json:"departure_date"`
    DepartureTime string `json:"departure_time"`
    FlightNumber string `json:"flight_number"`
    Duration TravelTime `json:"duration"`
    Origination Airport `json:"origination"`
    Passriders []Passrider `json:"passriders"`
}

type Credentials struct {
    UserID int 
    Username string `json:"username"`
    Password string `json:"password"`
}
