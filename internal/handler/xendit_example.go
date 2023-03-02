package handler

import (
	"fmt"
	"github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
	"log"
)

func init() {
	fmt.Printf("xendit ---- init")
	// Setup
	xendit.Opt.SecretKey = "xnd_development_1EaVNSrEDbbvMOcj2BOMC1vGiBvjuKwafpPiith8xcCvN9Fh399M0Ddqqan9RQ"

	//xendit.SetAPIRequester() // optional, useful for mocking

	data := invoice.CreateParams{
		ExternalID:  "invoice-example",
		Amount:      2,
		PayerEmail:  "customer@customer.com",
		Description: "invoice #1",
	}

	resp, err := invoice.Create(&data)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("resp -->>>> %v \n", resp)
	}

	p := invoice.GetParams{
		ID:        resp.ID,
		ForUserID: data.ForUserID,
	}
	// Get
	resp, err = invoice.Get(&p)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("resp -->>>> %v \n", resp)
	}
}
