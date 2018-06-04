package main

import (
	//	"encoding/json"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	_ "github.com/signavio/signa/ext/kubernetes/deployment"
	_ "github.com/signavio/signa/ext/kubernetes/get"
	_ "github.com/signavio/signa/ext/kubernetes/jobs"
	"github.com/signavio/signa/pkg/kubectl"
	//	"github.com/signavio/signa/pkg/slack"
)

// Version app
var Version = "version"

// BuildInfo app
var BuildInfo = "commit"

// Revision app
var Revision = Version + "+" + BuildInfo

// AppPort app
var AppPort = "443"

type response struct {
	FulfillmentText string `json:"fulfillmentText"`
}
type responseTmp struct {
	FulfillmentText     string `json:"fulfillmentText"`
	FulfillmentMessages struct {
		Card struct {
			Title    string
			Subtitle string
			ImageURI string
			Buttons  struct {
				Text     string
				Postback string
			}
		}
		Source  string
		Payload struct {
			Google struct {
				ExpectUserResponse bool
				RichResponse       struct {
					Items struct {
						SimpleResponse struct {
							TextToSpeech string
						}
					}
				}
			}
			Slack struct {
				Text string
			}
		}
		OutputContexts struct {
			Name          string
			LifespanCount int
			Parameters    struct {
				Param string
			}
		}
		FollowupEventInput struct {
			Name         string
			LanguageCode string
			Parameters   struct {
				Param string
			}
		}
	}
}

type message struct {
	ResponseID  string
	Session     string
	QueryResult struct {
		QueryText  string
		Parameters struct {
			Param0 string
			Param1 string
		}
	}
}

var m message
var resp response

func main() {
	//	configFile := flag.String(
	//		"config", "/etc/signa.yaml", "Path to the configuration file.")
	//	flag.Parse()
	//
	//	c := loadConfig(*configFile)
	//	//slack.Run(*configFile, c["slack-token"].(string))
	//
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/version", versionHandler)
	router.HandleFunc("/healthz", healthzHandler)

	log.Printf("Version: %v", Revision)
	//
	if os.Getenv("APP_PORT") != "" {
		AppPort = os.Getenv("APP_PORT")
	}
	if os.Getenv("APP_KIND") == "demo" {
		log.Printf("Mode: demo Server on: " + AppPort)
		router.HandleFunc("/", demoHandler)
		log.Fatal(http.ListenAndServe(":"+AppPort, router))
	}
	log.Printf("Server listen on " + AppPort)
	router.HandleFunc("/", tomHandler)
	log.Fatal(http.ListenAndServeTLS(":"+AppPort, "cert.pem", "key.pem", router))
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

func versionHandler(w http.ResponseWriter, r *http.Request) {
	var b []byte
	b = append([]byte("Version: "), Revision...)
	w.Write(b)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Healthz: alive!"))
}

func demoHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Welcome to DevOps Career Day!"))
}

func tomHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		log.Printf("Get GET Request!")
		w.Write([]byte("Alive!"))
		//fmt.Fprintf(w, "Hello, %q", html.EscapeString(" i have bought the tickets to the theatre"))
		//w.Write([]byte(response))
	case "POST":
		log.Printf("Get POST Request!")
		b, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err := json.Unmarshal(b, &m); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		fmt.Fprintf(w, m.QueryResult.Parameters.Param0)
		//result, _ := info(m)
		//mm := Message{"Alice", "Hello", 1294706395881547000}

		resp.FulfillmentText = "Alice"
		result, _ := json.Marshal(&resp)
		log.Print(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
		// Decode the JSON in the body and overwrite 'tom' with it
	/*	d := json.NewDecoder(r.Body).
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

func info(m message) (string, error) {

	args := []string{
		"get",
		"deployment",
		m.QueryResult.Parameters.Param0,
		"-o=jsonpath='{$.spec.template.spec.containers[:1].image}'",
		"-n",
		m.QueryResult.Parameters.Param1,
	}

	return fmt.Sprintf(currentImageVersion, "qrem-deploy", args), nil

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
