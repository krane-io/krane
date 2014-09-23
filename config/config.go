package config

import(
	"os"
	"os/user"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v1"
	"github.com/krane-io/krane/types"
)


// Function that return the path of Krane Configuration Files
func configPath()(path string) {

	user, err := user.Current()
	if err != nil {
		 log.Fatalln(err)
	}
	if user.Uid == "0" {
		path = "/etc/krane"
	} else {
		path = user.HomeDir+"/.krane"
	}
	//fmt.Printf("%#v", user)
	return path
}


// Returns Krane configuraton
func LoadConfigurationFile()(config types.KraneConfiguration) {
	ymlFile, err := os.Open(configPath()+"/server.yml")
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer ymlFile.Close()
	b, _ := ioutil.ReadAll(ymlFile) 
	yaml.Unmarshal(b, &config)
	return config
}