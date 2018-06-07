package fitbit

import (
	"context"
	"net/http"

	"net/url"
	"strconv"

	"time"

	"fmt"

	"net/http/httputil"

	"github.com/orisano/httpc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

type Client struct {
	httpClient *http.Client
	rb         *httpc.RequestBuilder
}

func NewClient(clientID, clientSecret string) *Client {
	rb, _ := httpc.NewRequestBuilder("https://api.fitbit.com/1", nil)

	cfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     fitbit.Endpoint,
		Scopes:       []string{"weight"},
		RedirectURL:  "http://localhost:5000/",
	}

	u := cfg.AuthCodeURL("")
	fmt.Println(u)

	tokenCh := make(chan *oauth2.Token)
	dump := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		b, err := httputil.DumpRequest(req, true)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Write(b)
		resp.Write([]byte(req.URL.RawQuery))
		fmt.Fprintln(resp)
		resp.Write([]byte(req.URL.Path))

		code := req.URL.Query().Get("code")

		token, err := cfg.Exchange(context.Background(), code)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenCh <- token
	})
	fmt.Println("redirect waiting...")
	go http.ListenAndServe(":5000", dump)

	token := <-tokenCh

	c := &Client{
		httpClient: oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token)),
		rb:         rb,
	}
	return c
}

func (c *Client) LogBodyFat(ctx context.Context, fat float64, date time.Time) error {
	params := url.Values{}
	params.Set("fat", strconv.FormatFloat(fat, 'f', 2, 64))
	params.Set("date", date.Format("2006-01-02"))

	req, _ := c.rb.NewRequest(ctx, http.MethodPost, "/user/-/body/log/fat.json", httpc.WithForm(params))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) LogWeight(ctx context.Context, weight float64, date time.Time) error {
	params := url.Values{}
	params.Set("weight", strconv.FormatFloat(weight, 'f', 2, 64))
	params.Set("date", date.Format("2006-01-02"))

	req, _ := c.rb.NewRequest(ctx, http.MethodPost, "/user/-/body/log/weight.json", httpc.WithForm(params))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
