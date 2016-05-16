package main

import (
    "encoding/base64"
    "fmt"
)

func debug(data []byte, err error) {
    if err == nil {
        fmt.Printf("%s\n\n", data)
    } else {
        fmt.Printf("%s\n\n", err)
    }
}

func encodeB64(message string) (retour string) {
    base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
    base64.StdEncoding.Encode(base64Text, []byte(message))
    return string(base64Text)
}

func printFlight(flight *Flight) string {
    seatCount := 0
    for _, rider := range flight.Passriders {
        seatCount = seatCount + rider.Seats
    }
    
    result := fmt.Sprintf("Origination: %s Destination: %s DepartureTime: %s %s Available(First: %d Business: %d Coach: %d) Passrider Count: %d Passrider Seats Taken: %d",
        flight.Origination.Code,
        flight.Destination.Code,
        flight.DepartureDate,
        flight.DepartureTime,
        flight.Available[0].First,
        flight.Available[0].Business,
        flight.Available[0].Coach,
        len(flight.Passriders),
        seatCount)
        
    fmt.Println(result)
    return result
}