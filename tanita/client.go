package tanita

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/orisano/httpc"
	"github.com/pkg/errors"
)

const (
	BaseURL = "https://www.healthplanet.jp"
)

type Client struct {
	username   string
	password   string
	httpClient *http.Client
	rb         *httpc.RequestBuilder
}

func NewClient(client *http.Client, username, password string) *Client {
	rb, _ := httpc.NewRequestBuilder(BaseURL, nil)
	return &Client{
		username:   username,
		password:   password,
		httpClient: client,
		rb:         rb,
	}
}

func (c *Client) Login(ctx context.Context) error {
	params := url.Values{}
	params.Set("loginId", c.username)
	params.Set("passwd", c.password)
	params.Set("send", "1")
	params.Set("url", "")

	req, err := c.rb.NewRequest(ctx, http.MethodPost, "/login.do", httpc.WithForm(params))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type BodyComposition struct {
	Weight  float64
	BodyFat float64
}

func (c *Client) GetBodyComposition(ctx context.Context, date time.Time) (*BodyComposition, error) {
	params := url.Values{}
	params.Set("date", date.Format("20060102"))

	req, err := c.rb.NewRequest(ctx, http.MethodGet, "/innerscan.do", httpc.WithQueries(params))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get innnerscan")
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse response html")
	}

	weightStr, _ := doc.Find(`input[name="innerscanBean[0][0].keyData"]`).Attr("value")
	bodyFatStr, _ := doc.Find(`input[name="innerscanBean[0][1].keyData"]`).Attr("value")

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid weight: %s", weightStr)
	}

	bodyFat, err := strconv.ParseFloat(bodyFatStr, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid body fat: %s", bodyFatStr)
	}

	return &BodyComposition{
		Weight:  weight,
		BodyFat: bodyFat,
	}, nil
}
