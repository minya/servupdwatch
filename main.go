package main

import (
	"encoding/json"
	"fmt"
	"github.com/minya/gopushover"
	"github.com/minya/goutils/config"
	"github.com/minya/goutils/web"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
)

func main() {
	settings, errSettings := getSettings()
	if errSettings != nil {
		fmt.Fprintf(os.Stderr, "can't get settings \n")
		os.Exit(-1)
	}
	fmt.Printf("settings read\n")

	state, errState := getState()
	if errState != nil {
		state = State{make(map[string]string)}
	}
	fmt.Printf("state read\n")

	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	for _, service := range settings.Services {
		url := service.VersionUrl
		fmt.Printf("process %v\n", url)
		response, getErr := client.Get(url)
		if getErr != nil {
			fmt.Fprintf(os.Stderr, "can't fetch %v\n", url)
			continue
		}

		bodyBin, bodyErr := ioutil.ReadAll(response.Body)
		if bodyErr != nil {
			fmt.Fprintf(os.Stderr, "can't read body %v\n", url)
			continue
		}
		var version ServiceVersion
		errUnmarshal := json.Unmarshal(bodyBin, &version)
		if errUnmarshal != nil {
			fmt.Fprintf(os.Stderr, "can't unmarshal json %v\n", url)
			continue
		}

		var stateChanged bool
		if state.Map[service.Id] != version.CommitHash {
			state.Map[service.Id] = version.CommitHash
			stateChanged = true
			fmt.Printf("version changed for %v\n", service.Name)
			Push(service.Name, version.CommitHash)
		} else {
			fmt.Printf("No changes for %v\n", service.Name)
		}

		if stateChanged {
			setState(state)
		}
	}

}

func Push(srv string, hash string) {
	var poSettings PushoverSettings
	err := config.UnmarshalJson(&poSettings, ".servupdwatch/pushover.json")
	if err != nil {
		fmt.Printf("can't read pushover settings\n")
		return
	}
	fmt.Printf("Pushover settings: %v\n", poSettings)
	message := srv + " updated to " + hash
	result, err := gopushover.SendMessage(poSettings.Token, poSettings.User, message)
	if err != nil {
		fmt.Printf("error while push: %v\n", err)
	} else {
		fmt.Printf("%v\n", result)
	}
}

func getSettings() (*Settings, error) {
	var settings Settings
	err := config.UnmarshalJson(&settings, ".servupdwatch/config.json")
	return &settings, err
}

func getState() (State, error) {
	var state State
	err := config.UnmarshalJson(&state, ".servupdwatch/state.json")
	return state, err
}

func setState(state State) {
	user, _ := user.Current()
	settingsRoot := path.Join(user.HomeDir, ".servupdwatch")
	statePath := path.Join(settingsRoot, "state.json")
	newStateBin, _ := json.Marshal(state)
	ioutil.WriteFile(statePath, newStateBin, 0660)
}

type ServiceInfo struct {
	Id         string
	Name       string
	VersionUrl string
}

type Settings struct {
	Services []ServiceInfo
}

type ServiceVersion struct {
	CommitHash string
}

type State struct {
	Map map[string]string
}

type PushoverSettings struct {
	User  string
	Token string
}
