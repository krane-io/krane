package aws

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/krane-io/krane/types"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Client struct {
	connection  *http.Client
	endpoint    string
	credentials aws.Auth
	region      map[string]*ec2.EC2
	cloud       types.Clouds
}

func (client *Client) Metal(accessKey string, secretKey string, plan string, name string) {
	garbageOutput, _ := regexp.Compile("(\\[[0-9]*-[0-9]*-[0-9T:+]*] [a-zA-Z]*: )")

	os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	os.Setenv("AWS_PLAN_ID", plan)
	os.Setenv("AWS_MACHINE_NAME", name)
	os.Setenv("AWS_KEY", "id_rsa")

	_, filename, _, _ := runtime.Caller(0)

	cmd := exec.Command("chef-client", "-z", path.Join(path.Dir(filename), "../../chef/aws.rb"))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("%s", err.Error())
	}

	ls := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		log.Errorf("%s", err.Error())
	}

	x := 0

	for {
		line, isPrefix, err := ls.ReadLine()
		if isPrefix {
			log.Errorf("%s", errors.New("isPrefix: true"))
		}
		if err != nil {
			if err != io.EOF {
				log.Errorf("%s", err.Error())
			}
			break
		}
		x = x + 1
		log.Infof("%s", garbageOutput.ReplaceAllString(string(line), ""))
	}
	err = cmd.Wait()
	if err != nil {
		log.Errorf("%s", err.Error())
	}
}

func (client *Client) Create(parameters url.Values) (string, error) {
	// var inspect_json map[string]interface{}
	if (parameters.Get("fqdn") == "") || (parameters.Get("name") == "") || (parameters.Get("plan") == "") {
		return "", errors.New("Error missing parameters to execute create ship")
	} else {
		go client.Metal(client.credentials.AccessKey, client.credentials.SecretKey, parameters.Get("plan"), parameters.Get("name"))
		time.Sleep(3 * time.Second)
		ship := client.FindShip(parameters.Get("name"))
		fmt.Printf("\n\n%#v\n", ship)
		if ship.Id == "" {
			time.Sleep(9 * time.Second)
			ship = client.FindShip(parameters.Get("name"))
		}
		return ship.Id, nil
	}

}

func (client *Client) Plan(parameters url.Values) ([]types.Plan, error) {
	var final []types.Plan

	if parameters.Get("name") == "all" {
		return client.cloud.Plans, nil
	} else if parameters.Get("name") != "" {
		name := strings.ToLower(parameters.Get("name"))
		for _, plan := range client.cloud.Plans {
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
		return client.cloud.Plans, nil
	}
}

func (client *Client) Name() string {
	return fmt.Sprintf("AWS driver")
}

func (client *Client) SearchTag(tags []ec2.Tag, key string) (value string) {
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (client *Client) FindShip(name string) types.Ship {
	var ship types.Ship

	filter := ec2.NewFilter()
	filter.Add("tag-key", "docker")
	filter.Add("tag:Name", name)

	for _, region := range client.region {
		response, _ := region.Instances(nil, filter)
		for _, pool := range response.Reservations {
			for _, instance := range pool.Instances {
				ship = types.Ship{
					instance.InstanceId,
					client.SearchTag(instance.Tags, "Name"),
					instance.DNSName,
					instance.PublicIpAddress,
					client.State(instance.State),
					"Ubuntu 14.04",
					instance.InstanceType,
					27017,
					"http",
					0,
					nil,
					false,
				}
			}
		}

	}

	return ship
}

func (client *Client) State(state ec2.InstanceState) (status string) {
	if state.Code == 16 {
		return "operational"
	} else if state.Code == 32 {
		return "shuting down"
	} else if state.Code == 48 {
		return "end"
	} else {
		return fmt.Sprintf("Unmapped for Code %d", state.Code)
	}
}
func (client *Client) ListRegion(c chan []types.Ship, parameters url.Values, region *ec2.EC2) {
	log.Infof("List Instances for Region %s", region.Name)
	var final []types.Ship
	filter := ec2.NewFilter()
	filter.Add("tag-key", "docker")

	resp, err := region.Instances(nil, filter)
	if err != nil {
		c <- final
	}
	for _, pool := range resp.Reservations {
		for _, ship := range pool.Instances {
			if client.State(ship.State) != "end" {
				final = append(final, types.Ship{ship.InstanceId, client.SearchTag(ship.Tags, "Name"), ship.DNSName, ship.PublicIpAddress, client.State(ship.State), "Ubuntu 14.04", ship.InstanceType, 27017, "http", 0, nil, false})
			}
		}
	}
	c <- final
}

func (client *Client) List(parameters url.Values) ([]types.Ship, error) {

	c := make(chan []types.Ship, len(client.region))
	var final []types.Ship

	for _, value := range client.region {
		go client.ListRegion(c, parameters, value)
	}

	for i := 1; i <= len(client.region); i++ {
		final = append(final, <-c...)
	}

	return final, nil
}

func (client *Client) Stop(args map[string]string) error {
	return nil
}

func (client *Client) Destroy(parameters url.Values) (string, error) {

	id := make([]string, 1)
	var err error

	if client.ValidateId(parameters.Get("idOrName")) {
		id[0] = parameters.Get("idOrName")
	} else {
		id[0] = client.FindShip(parameters.Get("idOrName")).Id
	}

	for _, region := range client.region {
		_, err = region.TerminateInstances(id)
		if err == nil {
			return id[0], nil
		}
	}
	return "", err
}

func (client *Client) ValidateId(text string) bool {
	r, _ := regexp.Compile("^([a-z]-[0-9a-z]{8})$")
	return r.MatchString(text)
}

func NewDriver() *Client {
	newClient := Client{}
	auth, err := aws.SharedAuth()
	if err != nil {
		log.Fatal(err)
	}
	newClient.credentials = auth
	newClient.region = make(map[string]*ec2.EC2)
	newClient.region["apnortheast"] = ec2.New(auth, aws.APNortheast)
	newClient.region["apsoutheast"] = ec2.New(auth, aws.APSoutheast)
	newClient.region["apsoutheast2"] = ec2.New(auth, aws.APSoutheast2)
	newClient.region["euwest"] = ec2.New(auth, aws.EUWest)
	newClient.region["saeast"] = ec2.New(auth, aws.SAEast)
	newClient.region["useast"] = ec2.New(auth, aws.USEast)
	newClient.region["uswest"] = ec2.New(auth, aws.USWest)
	newClient.region["uswest2"] = ec2.New(auth, aws.USWest2)

	newClient.cloud.Plans = []types.Plan{
		types.Plan{"53f0f0ecd8a5975a1c000162", "AWS", "USA", "Virginia A", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c000163", "AWS", "USA", "Virginia A", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000164", "AWS", "USA", "Virginia A", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c000165", "AWS", "USA", "Virginia A", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000166", "AWS", "USA", "Virginia A", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c000167", "AWS", "USA", "Virginia A", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000168", "AWS", "USA", "Virginia A", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000169", "AWS", "USA", "Virginia A", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00016a", "AWS", "USA", "Virginia A", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c00016b", "AWS", "USA", "Virginia A", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00016c", "AWS", "USA", "Virginia A", "m3.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00016d", "AWS", "USA", "Virginia A", "m3.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00016e", "AWS", "USA", "Virginia B", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c00016f", "AWS", "USA", "Virginia B", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000170", "AWS", "USA", "Virginia B", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c000171", "AWS", "USA", "Virginia B", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000172", "AWS", "USA", "Virginia B", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c000173", "AWS", "USA", "Virginia B", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000174", "AWS", "USA", "Virginia B", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000175", "AWS", "USA", "Virginia B", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000176", "AWS", "USA", "Virginia B", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000177", "AWS", "USA", "Virginia B", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000178", "AWS", "USA", "Virginia B", "m3.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000179", "AWS", "USA", "Virginia B", "m3.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00017a", "AWS", "USA", "Virginia C", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c00017b", "AWS", "USA", "Virginia C", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c00017c", "AWS", "USA", "Virginia C", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c00017d", "AWS", "USA", "Virginia C", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00017e", "AWS", "USA", "Virginia C", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c00017f", "AWS", "USA", "Virginia C", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000180", "AWS", "USA", "Virginia C", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000181", "AWS", "USA", "Virginia C", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000182", "AWS", "USA", "Virginia C", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000183", "AWS", "USA", "Virginia C", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000184", "AWS", "USA", "Virginia C", "m3.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000185", "AWS", "USA", "Virginia C", "m3.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000186", "AWS", "USA", "Virginia D", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c000187", "AWS", "USA", "Virginia D", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000188", "AWS", "USA", "Virginia D", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c000189", "AWS", "USA", "Virginia D", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00018a", "AWS", "USA", "Virginia D", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c00018b", "AWS", "USA", "Virginia D", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00018c", "AWS", "USA", "Virginia D", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00018d", "AWS", "USA", "Virginia D", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00018e", "AWS", "USA", "Virginia D", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c00018f", "AWS", "USA", "Virginia D", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000190", "AWS", "USA", "Virginia D", "m3.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000191", "AWS", "USA", "Virginia D", "m3.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000192", "AWS", "USA", "Virginia E", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c000193", "AWS", "USA", "Virginia E", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c000194", "AWS", "USA", "Virginia E", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c000195", "AWS", "USA", "Virginia E", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000196", "AWS", "USA", "Virginia E", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c000197", "AWS", "USA", "Virginia E", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000198", "AWS", "USA", "Virginia E", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c000199", "AWS", "USA", "Virginia E", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00019a", "AWS", "USA", "Virginia E", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c00019b", "AWS", "USA", "Virginia E", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00019c", "AWS", "USA", "Virginia E", "m3.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00019d", "AWS", "USA", "Virginia E", "m3.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c00019e", "AWS", "Europe", "Ireland A", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c00019f", "AWS", "Europe", "Ireland A", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001a0", "AWS", "Europe", "Ireland A", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c0001a1", "AWS", "Europe", "Ireland A", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001a2", "AWS", "Europe", "Ireland A", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c0001a3", "AWS", "Europe", "Ireland A", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001a4", "AWS", "Europe", "Ireland A", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001a5", "AWS", "Europe", "Ireland A", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001a6", "AWS", "Europe", "Ireland A", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001a7", "AWS", "Europe", "Ireland A", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001a8", "AWS", "Europe", "Ireland B", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c0001a9", "AWS", "Europe", "Ireland B", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001aa", "AWS", "Europe", "Ireland B", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c0001ab", "AWS", "Europe", "Ireland B", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001ac", "AWS", "Europe", "Ireland B", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c0001ad", "AWS", "Europe", "Ireland B", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001ae", "AWS", "Europe", "Ireland B", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001af", "AWS", "Europe", "Ireland B", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001b0", "AWS", "Europe", "Ireland B", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001b1", "AWS", "Europe", "Ireland B", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001b2", "AWS", "Europe", "Ireland C", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c0001b3", "AWS", "Europe", "Ireland C", "m1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001b4", "AWS", "Europe", "Ireland C", "m1.large"},
		types.Plan{"53f0f0ecd8a5975a1c0001b5", "AWS", "Europe", "Ireland C", "m1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001b6", "AWS", "Europe", "Ireland C", "t1.micro"},
		types.Plan{"53f0f0ecd8a5975a1c0001b7", "AWS", "Europe", "Ireland C", "m2.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001b8", "AWS", "Europe", "Ireland C", "m2.2xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001b9", "AWS", "Europe", "Ireland C", "m2.4xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001ba", "AWS", "Europe", "Ireland C", "c1.medium"},
		types.Plan{"53f0f0ecd8a5975a1c0001bb", "AWS", "Europe", "Ireland C", "c1.xlarge"},
		types.Plan{"53f0f0ecd8a5975a1c0001bc", "AWS", "USA", "N. California A", "m1.small"},
		types.Plan{"53f0f0ecd8a5975a1c0001bd", "AWS", "USA", "N. California A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001be", "AWS", "USA", "N. California A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001bf", "AWS", "USA", "N. California A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001c0", "AWS", "USA", "N. California A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001c1", "AWS", "USA", "N. California A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001c2", "AWS", "USA", "N. California A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001c3", "AWS", "USA", "N. California A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001c4", "AWS", "USA", "N. California A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001c5", "AWS", "USA", "N. California A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001c6", "AWS", "USA", "N. California B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001c7", "AWS", "USA", "N. California B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001c8", "AWS", "USA", "N. California B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001c9", "AWS", "USA", "N. California B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ca", "AWS", "USA", "N. California B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001cb", "AWS", "USA", "N. California B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001cc", "AWS", "USA", "N. California B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001cd", "AWS", "USA", "N. California B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ce", "AWS", "USA", "N. California B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001cf", "AWS", "USA", "N. California B", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001d0", "AWS", "USA", "N. California C", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001d1", "AWS", "USA", "N. California C", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001d2", "AWS", "USA", "N. California C", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001d3", "AWS", "USA", "N. California C", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001d4", "AWS", "USA", "N. California C", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001d5", "AWS", "USA", "N. California C", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001d6", "AWS", "USA", "N. California C", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001d7", "AWS", "USA", "N. California C", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001d8", "AWS", "USA", "N. California C", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001d9", "AWS", "USA", "N. California C", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001da", "AWS", "Asia Pacific", "Singapore A", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001db", "AWS", "Asia Pacific", "Singapore A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001dc", "AWS", "Asia Pacific", "Singapore A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001dd", "AWS", "Asia Pacific", "Singapore A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001de", "AWS", "Asia Pacific", "Singapore A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001df", "AWS", "Asia Pacific", "Singapore A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001e0", "AWS", "Asia Pacific", "Singapore A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001e1", "AWS", "Asia Pacific", "Singapore A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001e2", "AWS", "Asia Pacific", "Singapore A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001e3", "AWS", "Asia Pacific", "Singapore A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001e4", "AWS", "Asia Pacific", "Singapore B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001e5", "AWS", "Asia Pacific", "Singapore B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001e6", "AWS", "Asia Pacific", "Singapore B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001e7", "AWS", "Asia Pacific", "Singapore B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001e8", "AWS", "Asia Pacific", "Singapore B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001e9", "AWS", "Asia Pacific", "Singapore B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ea", "AWS", "Asia Pacific", "Singapore B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001eb", "AWS", "Asia Pacific", "Singapore B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ec", "AWS", "Asia Pacific", "Singapore B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001ed", "AWS", "Asia Pacific", "Singapore B", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ee", "AWS", "Asia Pacific", "Tokyo A", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001ef", "AWS", "Asia Pacific", "Tokyo A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001f0", "AWS", "Asia Pacific", "Tokyo A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001f1", "AWS", "Asia Pacific", "Tokyo A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001f2", "AWS", "Asia Pacific", "Tokyo A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001f3", "AWS", "Asia Pacific", "Tokyo A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001f4", "AWS", "Asia Pacific", "Tokyo A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001f5", "AWS", "Asia Pacific", "Tokyo A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001f6", "AWS", "Asia Pacific", "Tokyo A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001f7", "AWS", "Asia Pacific", "Tokyo A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001f8", "AWS", "Asia Pacific", "Tokyo B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c0001f9", "AWS", "Asia Pacific", "Tokyo B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c0001fa", "AWS", "Asia Pacific", "Tokyo B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c0001fb", "AWS", "Asia Pacific", "Tokyo B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001fc", "AWS", "Asia Pacific", "Tokyo B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c0001fd", "AWS", "Asia Pacific", "Tokyo B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001fe", "AWS", "Asia Pacific", "Tokyo B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c0001ff", "AWS", "Asia Pacific", "Tokyo B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000200", "AWS", "Asia Pacific", "Tokyo B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000201", "AWS", "Asia Pacific", "Tokyo B", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000202", "AWS", "Asia Pacific", "Tokyo C", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c000203", "AWS", "Asia Pacific", "Tokyo C", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000204", "AWS", "Asia Pacific", "Tokyo C", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c000205", "AWS", "Asia Pacific", "Tokyo C", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000206", "AWS", "Asia Pacific", "Tokyo C", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c000207", "AWS", "Asia Pacific", "Tokyo C", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000208", "AWS", "Asia Pacific", "Tokyo C", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000209", "AWS", "Asia Pacific", "Tokyo C", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00020a", "AWS", "Asia Pacific", "Tokyo C", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00020b", "AWS", "Asia Pacific", "Tokyo C", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00020c", "AWS", "USA", "Oregon A", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c00020d", "AWS", "USA", "Oregon A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00020e", "AWS", "USA", "Oregon A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c00020f", "AWS", "USA", "Oregon A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000210", "AWS", "USA", "Oregon A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c000211", "AWS", "USA", "Oregon A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000212", "AWS", "USA", "Oregon A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000213", "AWS", "USA", "Oregon A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000214", "AWS", "USA", "Oregon A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000215", "AWS", "USA", "Oregon A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000216", "AWS", "USA", "Oregon B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c000217", "AWS", "USA", "Oregon B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000218", "AWS", "USA", "Oregon B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c000219", "AWS", "USA", "Oregon B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00021a", "AWS", "USA", "Oregon B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c00021b", "AWS", "USA", "Oregon B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00021c", "AWS", "USA", "Oregon B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00021d", "AWS", "USA", "Oregon B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00021e", "AWS", "USA", "Oregon B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00021f", "AWS", "USA", "Oregon B", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000220", "AWS", "USA", "Oregon C", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c000221", "AWS", "USA", "Oregon C", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000222", "AWS", "USA", "Oregon C", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c000223", "AWS", "USA", "Oregon C", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000224", "AWS", "USA", "Oregon C", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c000225", "AWS", "USA", "Oregon C", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000226", "AWS", "USA", "Oregon C", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000227", "AWS", "USA", "Oregon C", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000228", "AWS", "USA", "Oregon C", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000229", "AWS", "USA", "Oregon C", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00022a", "AWS", "South America", "Sao Paulo A", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c00022b", "AWS", "South America", "Sao Paulo A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00022c", "AWS", "South America", "Sao Paulo A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c00022d", "AWS", "South America", "Sao Paulo A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00022e", "AWS", "South America", "Sao Paulo A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c00022f", "AWS", "South America", "Sao Paulo A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000230", "AWS", "South America", "Sao Paulo A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000231", "AWS", "South America", "Sao Paulo A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000232", "AWS", "South America", "Sao Paulo A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000233", "AWS", "South America", "Sao Paulo A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000234", "AWS", "South America", "Sao Paulo B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c000235", "AWS", "South America", "Sao Paulo B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000236", "AWS", "South America", "Sao Paulo B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c000237", "AWS", "South America", "Sao Paulo B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000238", "AWS", "South America", "Sao Paulo B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c000239", "AWS", "South America", "Sao Paulo B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00023a", "AWS", "South America", "Sao Paulo B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00023b", "AWS", "South America", "Sao Paulo B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00023c", "AWS", "South America", "Sao Paulo B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00023d", "AWS", "South America", "Sao Paulo B", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00023e", "AWS", "Asia Pacific", "Sydney A", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c00023f", "AWS", "Asia Pacific", "Sydney A", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000240", "AWS", "Asia Pacific", "Sydney A", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c000241", "AWS", "Asia Pacific", "Sydney A", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000242", "AWS", "Asia Pacific", "Sydney A", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c000243", "AWS", "Asia Pacific", "Sydney A", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000244", "AWS", "Asia Pacific", "Sydney A", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000245", "AWS", "Asia Pacific", "Sydney A", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000246", "AWS", "Asia Pacific", "Sydney A", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000247", "AWS", "Asia Pacific", "Sydney A", "c1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000248", "AWS", "Asia Pacific", "Sydney B", "m1.small"},
		types.Plan{"53f0f0edd8a5975a1c000249", "AWS", "Asia Pacific", "Sydney B", "m1.medium"},
		types.Plan{"53f0f0edd8a5975a1c00024a", "AWS", "Asia Pacific", "Sydney B", "m1.large"},
		types.Plan{"53f0f0edd8a5975a1c00024b", "AWS", "Asia Pacific", "Sydney B", "m1.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00024c", "AWS", "Asia Pacific", "Sydney B", "t1.micro"},
		types.Plan{"53f0f0edd8a5975a1c00024d", "AWS", "Asia Pacific", "Sydney B", "m2.xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00024e", "AWS", "Asia Pacific", "Sydney B", "m2.2xlarge"},
		types.Plan{"53f0f0edd8a5975a1c00024f", "AWS", "Asia Pacific", "Sydney B", "m2.4xlarge"},
		types.Plan{"53f0f0edd8a5975a1c000250", "AWS", "Asia Pacific", "Sydney B", "c1.medium"},
		types.Plan{"53f0f0edd8a5975a1c000251", "AWS", "Asia Pacific", "Sydney B", "c1.xlarge"},
	}
	return &newClient
}
