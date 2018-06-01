package main

import (
	//	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	_ "github.com/signavio/signa/ext/kubernetes/deployment"
	_ "github.com/signavio/signa/ext/kubernetes/get"
	_ "github.com/signavio/signa/ext/kubernetes/info"
	_ "github.com/signavio/signa/ext/kubernetes/jobs"
	"github.com/signavio/signa/pkg/kubectl"
	//	"github.com/signavio/signa/pkg/slack"
)

type resp struct {
	speech      string
	displayText string
}

func main() {
	configFile := flag.String(
		"config", "/etc/signa.yaml", "Path to the configuration file.")
	flag.Parse()

	c := loadConfig(*configFile)
	//slack.Run(*configFile, c["slack-token"].(string))

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", tomHandler)
	log.Printf("Go!%v", c)

	log.Fatal(http.ListenAndServe(":8080", router))

}

func loadConfig(file string) map[string]interface{} {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	var c map[string]interface{}
	d := yaml.NewDecoder(f)
	d.Decode(&c)

	return c
}

func tomHandler(w http.ResponseWriter, r *http.Request) {

	//response := resp{speech: " i have bought the tickets to the theatre"}

	switch r.Method {
	case "GET":
		// Just send out the JSON version of 'tom'
		//, _ := json.Marshal(tom)
		result, _ := info("version dev/qrem-backend")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
		//fmt.Fprintf(w, "Hello, %q", html.EscapeString(" i have bought the tickets to the theatre"))
		//w.Write([]byte(response))
	case "POST":
		result, _ := info("version dev/qrem-backend")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
		// Decode the JSON in the body and overwrite 'tom' with it
	/*	d := json.NewDecoder(r.Body)
		p := &person{}
		err := d.Decode(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		tom = p
	*/
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "I can't do that.")
	}
}

const (
	invalidAmountOfParams = "Invalid amount of parameters"
	invalidParams         = "Invalid parameters"

	noNamespaceNorDeployment = "No Namespace nor Deployment found in argument"
	currentImageVersion      = "Current deployed image and version for `%s`: ```%s```"
)

func info(c string) (string, error) {

	args := []string{
		"get",
		"deployment",
		"qrem-backend",
		"-o=jsonpath='{$.spec.template.spec.containers[:1].image}'",
		"-n",
		"dev",
	}
	k, err := kubectl.NewKubectl("default", args)
	if err != nil {
		// NOTE: Implement general logging later.
		return "", err
	}

	output, err := k.Exec()
	if err != nil {
		// NOTE: Implement general logging later.
		return "", err
	}

	return fmt.Sprintf(currentImageVersion, "qrem-deploy", output), nil
}
