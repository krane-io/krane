package concerto

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/krane-io/krane/config"
	"github.com/krane-io/krane/types"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Client struct {
	certificate tls.Certificate
	connection  *http.Client
	endpoint    string
}

func (client *Client) loadCertificates() {
	/**
	 * Loads Clients Certificates and creates and 509KeyPair
	 */
	var err error
	client.certificate, err = tls.LoadX509KeyPair(fmt.Sprintf("%s/concerto/cert.crt", config.ConfigPath()), fmt.Sprintf("%s/concerto/private/cert.key", config.ConfigPath()))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (client *Client) createConnection() {
	/**
	 * Creates a client with specific transport configurations
	 */
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{client.certificate}, InsecureSkipVerify: true},
	}
	client.connection = &http.Client{Transport: transport}
}

func (client *Client) Get(url string) []byte {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.connection.Do(request)
	defer response.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode >= 400 {
		log.Fatal(response.StatusCode)
		return nil
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return body
	}
}

func (client *Client) Put(url string) []byte {
	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.connection.Do(request)
	defer response.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode >= 400 {
		log.Fatal(response.StatusCode)
		return nil
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return body
	}
}

func (client *Client) Post(url string, parameters url.Values) []byte {
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(parameters.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(parameters.Encode())))

	response, err := client.connection.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		log.Fatal(response.StatusCode)
		return nil
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return body
	}
}

func (client *Client) Name() string {
	return fmt.Sprintf("Concerto driver for %s", client.endpoint)
}

func (client *Client) Create(parameters url.Values) (string, error) {
	var inspect_json map[string]interface{}
	if (parameters.Get("fqdn") == "") || (parameters.Get("name") == "") || (parameters.Get("plan") == "") {
		return "", errors.New("Error missing parameters to execute create ship")
	} else {
		output := client.Post(client.endpoint+"/krane/ships", parameters)

		err := json.Unmarshal(output, &inspect_json)

		if err != nil {
			return "", errors.New("Unable to marshal to json ship response")

		}
		return inspect_json["id"].(string), nil
	}

}

func (client *Client) Plan(parameters url.Values) ([]types.Plan, error) {
	output := client.Get(client.endpoint + "/krane/clouds")

	var cloud types.Clouds
	json.Unmarshal(output, &cloud)

	var final []types.Plan

	if parameters.Get("name") == "all" {
		return cloud.Plans, nil
	} else if parameters.Get("name") != "" {
		name := strings.ToLower(parameters.Get("name"))
		for _, plan := range cloud.Plans {
			if strings.Contains(strings.ToLower(plan.Id), name) {
				final = append(final, plan)
			} else if strings.Contains(strings.ToLower(plan.Provider), name) {
				final = append(final, plan)
			} else if strings.Contains(strings.ToLower(plan.Continent), name) {
				final = append(final, plan)
			} else if strings.Contains(strings.ToLower(plan.Region), name) {
				final = append(final, plan)
			} else if strings.Contains(strings.ToLower(plan.Plan), name) {
				final = append(final, plan)
			}
		}
		return final, nil
	} else {
		return cloud.Plans, nil
	}
}

func (client *Client) List(parameters url.Values) ([]types.Ship, error) {
	output := client.Get(client.endpoint + "/krane/ships")

	var fleet types.Fleet
	json.Unmarshal(output, &fleet)

	var final []types.Ship

	if parameters.Get("state") != "" {
		for _, ship := range fleet.Ships {
			if parameters.Get("state") == ship.State {
				final = append(final, ship)
			}
		}
	} else {
		final = fleet.Ships
	}

	return final, nil
}

func (client *Client) FindShip(name string) types.Ship {
	var final types.Ship
	ships, _ := client.List(nil)

	for _, ship := range ships {
		if ship.Name == name {
			return ship
		}
	}
	return final
}

func (client *Client) Destroy(parameters url.Values) (string, error) {
	return "", nil
}

func (client *Client) ValidateId(text string) bool {
	r, _ := regexp.Compile("^([0-9a-z]{24})$")
	return r.MatchString(text)
}

func NewDriver() *Client {
	newClient := Client{}
	newClient.endpoint = "https://clients.concerto.io:886"
	newClient.loadCertificates()
	newClient.createConnection()
	return &newClient
}

func (client *Client) StartShip(id string) {
	output := client.Put(client.endpoint + "/krane/ships/" + id + "/start")
	fmt.Printf("%s", string(output))
}

func (client *Client) Stop(args map[string]string) error {
	output := client.Put(client.endpoint + "/krane/ships/" + args["id"] + "/stop")
	fmt.Printf("%s", string(output))
	return nil
}
