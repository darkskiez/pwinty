// Package pwinty is an incomplete implementation of Pwinty API service in Go
package pwinty

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	EndpointProduction = "https://api.pwinty.com"
	EndpointSandbox    = "https://sandbox.pwinty.com"

	StatusNotYetSubmitted = "NotYetSubmitted"
	StatusSubmitted       = "Submitted"
	StatusComplete        = "Complete"
	StatusCancelled       = "Cancelled"

	StatusAwaitingURLOrData = "AwaitingUrlOrData"
	StatusNotYetDownloaded  = "NotYetDownloaded"
	StatusOk                = "Ok"
	StatusFileNotFoundAtURL = "FileNotFoundAtUrl"
	StatusInvalid           = "Invalid"

	SizingCrop             = "Crop"
	SizingShrinkToFit      = "ShrinkToFit"
	SizingShrinkToExactFit = "ShrinkToExactFit"
)

type Order struct {
	ID                int      `json:"id"`
	IsValid           bool     `json:"isValid"`
	GeneralErrors     []string `json:"generalErrors"`
	RecipientName     string   `json:"recipientName"`
	Address1          string   `json:"address1"`
	Address2          string   `json:"address2"`
	AddressTownOrCity string   `json:"addressTownOrCity"`
	StateOrCounty     string   `json:"stateOrCounty"`
	PostalOrZipCode   string   `json:"postalOrZipCode"`
	Country           string   `json:"country"`
	Status            string   `json:"status"`
	Photos            []Photo  `json:"photos"`
}

type Photo struct {
	ID       int      `json:"id"`
	Type     string   `json:"type"`
	URL      string   `json:"url"`
	Status   string   `json:"status"`
	Copies   int      `json:"copies"`
	Sizing   string   `json:"sizing"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type Client struct {
	httpClient *http.Client
	merchantID string
	apiKey     string
	endpoint   string
}

// NewClient instantiates a new Client that will be able to query the API. if httpClient is nil, http.DefaultClient will be used.
func NewClient(merchantID, apiKey, endpoint string, httpClient *http.Client) *Client {
	c := httpClient
	if c == nil {
		c = http.DefaultClient
	}

	return &Client{httpClient: c, merchantID: merchantID, apiKey: apiKey, endpoint: endpoint}
}

// setAuthenticationHeaders set the correct headers for an authenticated request on the API
func (c *Client) setAuthenticationHeaders(req *http.Request) {
	req.Header.Set("X-Pwinty-MerchantId", c.merchantID)
	req.Header.Set("X-Pwinty-REST-API-Key", c.apiKey)
}

func (c *Client) prepareGetRequest(path string, values url.Values) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.endpoint, path)
	if values != nil {
		url = fmt.Sprintf("%s?%s", url, values.Encode())
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthenticationHeaders(req)

	return req, nil
}

func (c *Client) prepareFormRequest(method, path string, values url.Values) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.endpoint, path), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setAuthenticationHeaders(req)

	return req, nil
}

func (c *Client) GetOrder(id int) (*Order, error) {
	req, err := c.prepareGetRequest("/Orders", url.Values{"id": {strconv.Itoa(id)}})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	order := &Order{}
	if err := json.Unmarshal(body, order); err != nil {
		return nil, err
	}

	return order, nil
}

// PostOrder creates a new order on the pwinty API
func (c *Client) PostOrder(recipientName, address1, address2, town, state, zip, country string) (*Order, error) {
	req, err := c.prepareFormRequest("POST", "/Orders", url.Values{
		"recipientName":     {recipientName},
		"address1":          {address1},
		"address2":          {address2},
		"addressTownOrCity": {town},
		"stateOrCounty":     {state},
		"postalOrZipCode":   {zip},
		"country":           {country},
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	order := &Order{}
	if err := json.Unmarshal(body, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (c *Client) OrderStatus(id int, status string) error {
	req, err := c.prepareFormRequest("POST", "/Orders/Status", url.Values{
		"id":     {string(id)},
		"status": {status},
	})
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

func (c *Client) SubmissionStatus(id int) (*Order, error) {
	req, err := c.prepareGetRequest("/Orders/SubmissionStatus", url.Values{"id": {strconv.Itoa(id)}})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	order := &Order{}
	if err := json.Unmarshal(body, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (c *Client) PostPhotoURL(orderId int, photoType, photoUrl string, copies int, sizing string) (*Photo, error) {
	req, err := c.prepareFormRequest("POST", "/Photos", url.Values{
		"orderId": {string(orderId)},
		"type":    {photoType},
		"url":     {photoUrl},
		"copies":  {string(copies)},
		"sizing":  {sizing},
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	photo := &Photo{}
	if err := json.Unmarshal(body, photo); err != nil {
		return nil, err
	}

	return photo, nil
}
