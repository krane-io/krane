## Requisites

Please install and download GOPM package manager. If you have issues please visit https://github.com/gpmgo/gopm for a more details process.

	go get -u github.com/gpmgo/gopm

## Installation

First lets download all the code

	go get -u github.com/krane-io/krane

Then we install all the package dependencies

    cd $GOPATH/src/github.com/krane-io/krane
    gopm get
    
Then we compile the project
    
    gopm build
    
Voilà you should have a Krane binary ready to use.

## Configuration

Firstly create a folder in your home called *.krane*

    mkdir -p ~/.krane/
    
Then create a configuration file called *config.yml*

```
--- 
production: 
  server: 
    driver: concerto
    host: 
      fqdn: 127.0.0.1
      name: localhost
      port: 1979
      schema: tcp
```
## Configuration AWS

In case you want to use AWS as your default cloud provider please indicate it in krane/config.yml in the driver part. Once indicated Krane will automatically use AWS tools to get your cloud credentials. You can configure you aws tools set executing

    aws configure
    
After doing this you need to install chef-metal so that krane can install docker uniformly across multiple machines.

    gem install chef-metal-fog chef-metal
    
## Configuration Concerto

Log into concerto and go to *Settings -> Your Account* and download you api key and decompress it in *~/.krane/concerto*

    mkdir -p ~/.krane/concerto
    unzip api-key.zip
