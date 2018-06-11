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
	"os/exec"
	"strings"

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

// TlsCert path
var TlsCert = os.Getenv("TLS_CERT_PATH")

// TlsKey path
var TlsKey = os.Getenv("TLS_KEY_PATH")

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
			ClusterCommand    string
			ClusterResource   string
			ApplicationName   string
			EnvironmentName   string
			Formatting        string
			Label             string
			Replicas          string
			Version           string
			ApplicationCanary string
			ApplicationIstio  string
		}
	}
}

type args struct {
	command         string //	"get"
	resource        string //	"deployment"
	applicationName string //	m.QueryResult.Parameters.ApplicationName
	jsonPath        string // 	"-o=jsonpath='{$.spec.template.spec.containers[:1].image}'"
	environment     string //	"-n" m.QueryResult.Parameters.EnvironmentName
}

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
	log.Fatal(http.ListenAndServeTLS(":"+AppPort, CERT, KEY, router))
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
		var m message
		var resp response
		var arg []string
		//var result string

		b, _ := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err := json.Unmarshal(b, &m); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(422) // unprocessable entity
			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}
		}
		switch m.QueryResult.Parameters.ClusterCommand {
		case "get":

			arg = []string{"get",
				"deployment",
				"front-v1",
				"-o=jsonpath='{$.spec.template.spec.containers[:1].image}'",
				"-n" + "demo"}
			result, _ := kubeCtl(m, arg)
			result = fmt.Sprintf(currentImageVersion, m.QueryResult.Parameters.ApplicationName, strings.Split(strings.Split(result, "/")[len(strings.Split(result, "/"))-1], ":")[1])

			log.Print(result, m)

			resp.FulfillmentText = result
			speechText, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(speechText))

		case "getDeployment":
			m.QueryResult.Parameters.ClusterResource = "deployment"
			arg = []string{"get",
				m.QueryResult.Parameters.ClusterResource,
				"-n" + m.QueryResult.Parameters.EnvironmentName,
				m.QueryResult.Parameters.Formatting, "--no-headers",
				m.QueryResult.Parameters.Label}
			result, _ := kubeCtl(m, arg)
			log.Print(result, m)

			resp.FulfillmentText = result
			speechText, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(speechText))

		case "scale":
			m.QueryResult.Parameters.ApplicationName = "front-v1"
			arg = []string{"scale",
				"deployment",
				m.QueryResult.Parameters.ApplicationName,
				"-n", m.QueryResult.Parameters.EnvironmentName,
				"--replicas", m.QueryResult.Parameters.Replicas}

			result, err := kubeCtl(m, arg)
			//result, err := exec.Command("/usr/local/bin/kubectl", "scale", "deployment", m.QueryResult.Parameters.ApplicationName, "--replicas", m.QueryResult.Parameters.Replicas, "-n", m.QueryResult.Parameters.EnvironmentName).Output()

			log.Print("result: ", string(result), arg, err)

			resp.FulfillmentText = "Frontend scaled up to " + m.QueryResult.Parameters.Replicas + " replicas"
			speechText, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(speechText))
		//k -n demo get svc front-canary -o yaml --export|sed 's/weight:.[0-9]$/weight: 50/'

		case "canary":

			result, err := exec.Command("/bin/bash", "/tmp/canary.sh", m.QueryResult.Parameters.ApplicationCanary).Output()

			log.Print("result: ", string(result), arg, err)

			resp.FulfillmentText = "Canary applyed to " + m.QueryResult.Parameters.ApplicationCanary + " percent"
			speechText, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(speechText))

		case "istio":
			result, err := exec.Command("/bin/bash", "/tmp/istio.sh", m.QueryResult.Parameters.ApplicationIstio).Output()

			log.Print("result: ", string(result), arg, err)

			resp.FulfillmentText = "Traffic Policy applyed to " + m.QueryResult.Parameters.ApplicationCanary + " percent"
			speechText, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(speechText))
		}
	default:
		var resp response
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp.FulfillmentText = "Sorry, I can't do that."
		speechText, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(speechText))
	}

}

const (
	invalidAmountOfParams = "Invalid amount of parameters"
	invalidParams         = "Invalid parameters"

	noNamespaceNorDeployment = "No Namespace nor Deployment found in argument"
	currentImageVersion      = "Current deployed image and version for %s: %s"
)

func kubeCtl(m message, arg []string) (string, error) {

	//return fmt.Sprintf(currentImageVersion, "qrem-deploy", args), nil

	k, err := kubectl.NewKubectl("default", arg)
	if err != nil {
		// NOTE: Implement general logging later.
		return "Something went wrong. Ask kube for help: NewKubectl", err
	}

	output, err := k.Exec()
	if err != nil {
		// NOTE: Implement general logging later.
		return "Something went wrong. Ask kube for help:", err
	}

	return output, nil
}

func k(m message, arg string) (string, error) {

	out, err := exec.Command("/usr/local/bin/kubectl", "scale", "deployment", "front-v1", "--replicas=1", "-n", "demo").Output()
	if err != nil {
		log.Print(err)
	}
	return string(out), err
}
