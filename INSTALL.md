## Requisites

Please install and download GOPM package mangager. If you have issues please visit https://github.com/gpmgo/gopm for a more details process.

	go get -u github.com/gpmgo/gopm

## Instalation

First lets download all the code

	go get -u github.com/krane-io/krane

Then we install all the package dependencies

    cd $GOPATH/src/github.com/krane-io/krane
    gopm get
    
Then we compile the project
    
    gopm build
    
Voilà you should have a Krane binary ready to use.
