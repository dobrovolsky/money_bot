package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

var c = Client{}

func main() {
	guid := uuid.Must(uuid.NewRandom())

	respBody, err := c.SendOTP(guid, "+380958790190")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("response: %s", respBody)

	var code string
	fmt.Print("enter code: ")
	fmt.Scanf("%s", &code)

	respBody, err = c.ConfirmationOtp(guid, "+380958790190", code)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("response: %s", respBody)
	// c.setTokens([]byte(``))
	data, err := c.GetLastChequeHeaders()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

}
