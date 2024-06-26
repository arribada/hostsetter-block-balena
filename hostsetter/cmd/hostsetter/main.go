// Copyright (c) The Arribada initiative.
// Licensed under the MIT License.

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	flagVerbose  bool
	flagHostname string
	flagForce    bool
)

func init() {
	flag.BoolVar(&flagVerbose, "v", false, "be verbose")
	flag.StringVar(&flagHostname, "hostname", "", "the hostname we want to set")
	flag.BoolVar(&flagForce, "force", false, "force setting the hostname even if the check finds it already set")
}

var (
	superAddr string
	superKey  string
	theURL    string
)

func main() {
	flag.Parse()

	// env var exists and overrides flag, because env vars are easier with docker-compose...
	envHostname := os.Getenv("HST_HOSTNAME")
	if envHostname != "" {
		flagHostname = envHostname
	}
	if flagHostname == "" {
		log.Fatalf("mandatory -hostname was not provided")
	}
	if flagVerbose {
		log.Printf("called with -hostname: %v", flagHostname)
	}

	superAddr = os.Getenv("BALENA_SUPERVISOR_ADDRESS")
	superKey = os.Getenv("BALENA_SUPERVISOR_API_KEY")
	theURL = fmt.Sprintf("%s/v1/device/host-config?apikey=%s", superAddr, superKey)

	if superKey == "" {
		log.Fatal("BALENA_SUPERVISOR_API_KEY not set")
	}

	for {
		err := checkAndSet()
		if err == nil {
			// Mission accomplished, terminating.
			break
		}
		log.Printf("%v", err)
		time.Sleep(time.Minute)
	}
}

type config struct {
	Network Net `json:"network"`
}

type Net struct {
	Hostname string `json:"hostname"`
}

func getHostname() (string, error) {
	res, err := http.Get(theURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var conf config
	if err := json.Unmarshal(data, &conf); err != nil {
		return "", err
	}

	return conf.Network.Hostname, nil
}

func setHostname(val string) error {
	patch := config{
		Network: Net{
			Hostname: val,
		},
	}
	data, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", theURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil
}

func checkAndSet() error {
	curHostname, err := getHostname()
	if err != nil {
		return fmt.Errorf("while getting hostname: %v", err)
	}

	// TODO: that check may not be enough is someone other than us decided
	// to take the same hostname afterwards. So do we need to check for that,
	// and fight for it?
	if curHostname == flagHostname {
		if !flagForce {
			if flagVerbose {
				log.Printf("%s already set as hostname, nothing to do", flagHostname)
			}
			return nil
		}
		// We'll get a 423 if we try to force set the hostname if it is already set to
		// the same value. So we first "unset" it, but setting it to another value.
		if err := setHostname(genID()); err != nil {
			return fmt.Errorf("while (un)setting hostname: %v", err)
		}
	}

	if err := setHostname(flagHostname); err != nil {
		return fmt.Errorf("while setting hostname: %v", err)
	}

	return nil
}

func genID() string {
	h := sha1.New()
	h.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", h.Sum(nil))[:20]
}
