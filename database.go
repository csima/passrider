package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "fmt"
    "github.com/davecgh/go-spew/spew"
)
func dbInit() {
    dbtmp, err := gorm.Open("sqlite3", "/db/test.db")
    if err != nil {
        panic("failed to connect database")
    }
    db = dbtmp
    //db.LogMode(true)

    if db.HasTable("users") == false {
        db.CreateTable(&User{})
    }
    if db.HasTable("searches") == false {
        db.CreateTable(&Search{})
    }
    if db.HasTable("flights") == false {
        db.CreateTable(&Flight{})
    }
    if db.HasTable("seats") == false {
        db.CreateTable(&Seats{})
    }
    if db.HasTable("airports") == false {
        db.CreateTable(&Airport{})
    }
    if db.HasTable("travel_times") == false {
        db.CreateTable(&TravelTime{})
    }
    if db.HasTable("passriders") == false {
        db.CreateTable(&Passrider{})
    }
    if db.HasTable("monitors") == false {
        db.CreateTable(&Monitor{})
    }
    if db.HasTable("credentials") == false {
        db.CreateTable(&Credentials{})
    }
}

func testDB() {
    
    user := User{}
    searches := Search{}
    db.Table("users").Where("user_identifier = ?", "poop").Scan(&user)
    db.Model(&user).Related(&searches)
    spew.Dump(user)
    spew.Dump(searches)
    flights := []*Flight{}
    db.Model(&searches).Association("Results").Find(&flights)
    seats := Seats{}
    passriders := Passrider{}
    for _,flight := range flights {
        spew.Dump(flight)
        db.Model(&flight).Association("Authorized").Find(&seats)
        db.Model(&flight).Association("Passriders").Find(&passriders)
        spew.Dump(passriders)
        fmt.Println("**************")
    }
}