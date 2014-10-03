package concerto

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/krane-io/krane/config"
	"github.com/krane-io/krane/types"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/tabwriter"
)

type Client struct {
	certificate tls.Certificate
	connection  *http.Client
	endpoint    string
}

type Clouds struct {
	Plans []Plan
}

type Plan struct {
	Id        string `json:"id"`
	Provider  string `json:"cloud_provider"`
	Continent string `json:"continent"`
	Region    string `json:"region"`
	Plan      string `json:"plan"`
}

func (client *Client) loadCertificates() {
	/**
	 * Loads Clients Certificates and creates and 509KeyPair
	 */
	var err error
	client.certificate, err = tls.LoadX509KeyPair(fmt.Sprintf("%s/concerto/public.pem", config.ConfigPath()), fmt.Sprintf("%s/concerto/private.pem", config.ConfigPath()))
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.connection.Do(request)
	defer response.Body.Close()

	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.connection.Do(request)
	defer response.Body.Close()

	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(parameters.Encode())))

	response, err := client.connection.Do(request)

	if err != nil {
		log.Fatalln(err)
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

func (client *Client) GetCloudPlans() []Plan {
	output := client.Get(client.endpoint + "/krane/clouds")

	var cloud Clouds
	json.Unmarshal(output, &cloud)

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tPROVIDER\tCONTINENT\tREGION\tPLAN\n")

	for _, plan := range cloud.Plans {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", plan.Id, plan.Provider, plan.Continent, plan.Region, plan.Plan)
	}
	w.Flush()

	return cloud.Plans
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

	// w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	// fmt.Fprint(w, "ID\tNAME\tFQDN\tIP\tSTATE\tOS\tPLAN\n")

	// for _, ship := range fleet.Ships {
	// 	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", ship.Id, ship.Name, ship.Fqdn, ship.Ip, ship.State, ship.Os, ship.Plan)
	// }
	// w.Flush()

	return final, nil
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

// func main() {

// 	// start_krane_ship PUT    /krane/ships/:id/start(.:format)                                       api/krane/ships#start
// 	//      krane_ships GET    /krane/ships(.:format)                                                 api/krane/ships#index
// 	//     krane_clouds GET    /krane/clouds(.:format)
// 	concerto := NewDriver()

// 	// concerto.CreateShip("cadvisor5.concerto.io", "53f0f10fd8a5975a1c00038f")
// 	concerto.GetShips()
// 	//concerto.GetCloudPlans()
// 	//concerto.StopShip("241a7c8e8d88")
// 	//concerto.StartShip("da33c245fdc3")
// 	// fmt.Printf("\n\n\n")
// }
