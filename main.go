package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/orisano/httpc"
	"github.com/rt33/fitbit/fitibit"
	"github.com/rt33/fitbit/tanita"
)

func main() {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: jar,
	}

	tanitaClient := tanita.NewClient(httpClient, os.Getenv("TANITA_USERNAME"), os.Getenv("TANITA_PASSWORD"))

	ctx := context.Background()

	tanitaClient.Login(ctx)
	t := time.Date(2018, 5, 11, 0, 0, 0, 0, time.UTC)
	bodyCompostion, err := tanitaClient.GetBodyComposition(ctx, t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bodyCompostion)

	httpc.InjectDebugTransport(http.DefaultClient, os.Stderr)

	fitbitClient := fitbit.NewClient("FITBIT_CLIENTID", "FITBIT_CLIENT_SECRET")
	if err := fitbitClient.LogBodyFat(ctx, bodyCompostion.BodyFat, t); err != nil {
		log.Fatal(err)
	}

	if err := fitbitClient.LogWeight(ctx, bodyCompostion.Weight, t); err != nil {
		log.Fatal(err)
	}
}
