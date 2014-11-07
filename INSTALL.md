## Prerequisites

Before you start, you will need to install and download the GOPM package manager. You can install this by typing:

    go get -u github.com/gpmgo/gopm

If you have any issues, please visit https://github.com/gpmgo/gopm for a more detailed description of the install process for GOPM.

## Installation of Krane

First, download all the code:

    go get -u github.com/krane-io/krane

Next, install all the package dependencies:

    cd $GOPATH/src/github.com/krane-io/krane
    gopm get
    
Then compile the project:
    
    gopm build
    
Voilà, you should now have a Krane binary ready to use.

## Configuration

First, create a folder in your home called *.krane*

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
## Configuring AWS as your cloud driver

In case you want to use AWS as your default cloud driver, please indicate this in *krane/config.yml* by using the line:

    driver: aws

Once you have done this, Krane will automatically use AWS tools to get your cloud credentials. You can configure your AWS tools set executing:

    aws configure
    
After doing this, you need to install chef-metal so that Krane can install Docker uniformly across multiple machines.

    gem install chef-metal-fog chef-metal
    
## Configuring Flexiant Concerto as your cloud driver

First, if you don't already have a Flexiant Concerto account, log into *http://krane.io/* and create a free account.

Next, log into Concerto (use the *krane.io* link if that's how you got your account), go to *Settings -> Your Account* and download you api key and decompress it into *~/.krane/concerto* as follows:

    mkdir -p ~/.krane/concerto
    unzip api-key.zip

That's it. Now you can use Krane to provision Docker containers across multiple clouds.
