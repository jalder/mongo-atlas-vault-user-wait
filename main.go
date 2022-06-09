package main

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type AtlasApi struct {
	PublicKey string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
	ProjectId string `json:"projectId"`
	ClusterName string `json:"clusterName"`
}

type AtlasStatus struct {
	ChangeStatus string `json:"changeStatus"`
}

func MD5Hex(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func AtlasApiConf(f string) AtlasApi {
	var api AtlasApi
	conf, _ := os.Open(f)
	defer conf.Close()
	p := json.NewDecoder(conf)
    p.Decode(&api)
    return api
}

func MongoPing(mongoUri string) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
	}()
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("MongoDB Atlas Authentication Succeeded and Primary Pinged.")
}

func main() {
	fmt.Println("Starting MongoDB Atlas x Vault DB User Liveness Check")

	//pull config from vault secrets mount
	confPath := ""
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}
	conf := AtlasApiConf(confPath)
	groupId := conf.ProjectId
	clusterName := conf.ClusterName
	publicKey := conf.PublicKey
	privateKey := conf.PrivateKey

	dbPath := ""
	if len(os.Args) > 2 {
		dbPath = os.Args[2]
	}
	dbFile, _ := ioutil.ReadFile(dbPath)
	mongoUri := string(dbFile)

	//environment variables take precedence
	if os.Getenv("MONGODB_ATLAS_PROJECT_ID") != "" {
		groupId = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	}
	if os.Getenv("MONGODB_ATLAS_CLUSTER_NAME") != "" {
		clusterName = os.Getenv("MONGODB_ATLAS_CLUSTER_NAME")
	}
	if os.Getenv("MONGODB_ATLAS_PUBLIC_API_KEY") != "" {
		publicKey = os.Getenv("MONGODB_ATLAS_PUBLIC_API_KEY")
	}
	if os.Getenv("MONGODB_ATLAS_PRIVATE_API_KEY") != "" {
		privateKey = os.Getenv("MONGODB_ATLAS_PRIVATE_API_KEY")
	}
	if os.Getenv("MONGODB_URI") != "" {
		mongoUri = os.Getenv("MONGODB_URI")
	}
	
	if groupId == "" || clusterName == "" || publicKey == "" || privateKey == "" || mongoUri == "" {
		fmt.Println("Missing required variables.")
		os.Exit(1)
	}
	for {
		fmt.Println("Checking Status of Cluster User Changes...")

		url := "https://cloud.mongodb.com/api/atlas/v1.0/groups/" + groupId + "/clusters/" + clusterName + "/status"

		response, err := http.Get(url)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		nonce := ""
		realm := ""
		qop := ""

		if response.StatusCode != http.StatusUnauthorized {
			fmt.Println("Unexpected response to digest auth request")
			fmt.Println(response)
			os.Exit(1)
		} else {
			authHeaders := strings.Split(response.Header["Www-Authenticate"][0], ",")
			for _, r := range authHeaders {
				if strings.Contains(r, "nonce") {
					nonce = strings.Split(r, `"`)[1]
				}
				if strings.Contains(r, "realm") {
					realm = strings.Split(r, `"`)[1]
				}
				if strings.Contains(r, "qop") {
					qop = strings.Split(r, `"`)[1]
				}
			}
		}

		if nonce == "" {
			fmt.Println("Error performing HTTP Digest Authentication")
			os.Exit(1)
		} else {
			ha1 := MD5Hex(publicKey + ":" + realm + ":" + privateKey)
			ha2 := MD5Hex("GET:" + url)
			nonceCount := 00000001
			b := make([]byte, 8)
			io.ReadFull(rand.Reader, b)
			cnonce := fmt.Sprintf("%x", b)[:16]
			authResponse := MD5Hex(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, nonce, nonceCount, cnonce, qop, ha2))
			auth := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc=%v, qop=%s, response="%s", algorithm="MD5"`, publicKey, realm, nonce, url, cnonce, nonceCount, qop, authResponse)
			//		fmt.Println("Built digest auth header...")
			//		fmt.Println(auth)
			client := &http.Client{}
			request, err := http.NewRequest("GET", url, nil)
			request.Header.Set("Authorization", auth)
			request.Header.Set("Accept", "application/json")
			response, err := client.Do(request)
			defer response.Body.Close()

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			status, err := ioutil.ReadAll(response.Body)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			var state AtlasStatus
			err = json.Unmarshal([]byte(status), &state)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			} else {
				if state.ChangeStatus == "APPLIED" {
					fmt.Println("Atlas reports changeStatus: " + string(state.ChangeStatus))
					fmt.Println("Confirming Vault Credentials and Atlas Access are Valid...")
					MongoPing(mongoUri)
					fmt.Println("Exiting")
					os.Exit(0)
				} else {
					fmt.Println("Atlas reports changeStatus: " + string(state.ChangeStatus))
					fmt.Println("Sleeping...")
					time.Sleep(5 * time.Second)
					//todo: exit 1 at some point...
				}
			}
		}
	}
}
