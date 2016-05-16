package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    //"fmt"
    //"github.com/davecgh/go-spew/spew"
)

var db *gorm.DB

func populateFlight(flight *Flight) {
    db.Model(&flight).Related(&flight.Passriders)
    db.Model(&flight).Related(&flight.Authorized)
    db.Model(&flight).Related(&flight.Available)
    db.Model(&flight).Related(&flight.Booked)
    db.Model(&flight).Related(&flight.Capacity)
    db.Model(&flight).Related(&flight.Origination)
    db.Model(&flight).Related(&flight.Destination)
    //spew.Dump(flight)
}

func main() {
    user := User{}
    flights := []Flight{}
    //searches := Search{}
    monitored := []Monitor{}
    dbInit()
    db.First(&user)
    db.Model(&user).Related(&monitored)
    db.Model(&monitored[0]).Related(&flights)
    for _,flight := range flights {
        //fmt.Println(flight.FlightNumber)
        //printFlight(&flight)
        populateFlight(&flight)
        printFlight(&flight)
        //spew.Dump(flight)
    }
    //db.Model(&monitored).Related(&flights)
   
   /* 
    spew.Dump(user)
    spew.Dump(monitored)
    spew.Dump(flights)*/
}