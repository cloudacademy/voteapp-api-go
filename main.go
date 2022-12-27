package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	//including gorilla mux and handlers packages for HTTP routing and CORS support
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	//connections to mongo
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/shirou/gopsutil/cpu"
	"runtime"
)

const mongo_db = "langdb"
const mongo_collection = "languages"
const mongo_default_conn_str = "mongodb://mongo-0.mongo,mongo-1.mongo,mongo-2.mongo:27017/langdb"
const mongo_default_username = "admin"
const mongo_default_password = "password"

type codedetail struct {
	Usecase  string `json:"usecase,omitempty" bson:"usecase"`
	Rank     int    `json:"rank,omitempty" bson:"rank"`
	Compiled bool   `json:"compiled" bson:"compiled"`
	Homepage string `json:"homepage,omitempty" bson:"homepage"`
	Download string `json:"download,omitempty" bson:"download"`
	Votes    int64  `json:"votes" bson:"votes"`
}

type language struct {
	Name   string     `json:"name,omitempty" bson:"name"`
	Detail codedetail `json:"codedetail,omitempty" bson:"codedetail"`
}

type voteresult struct {
	Name   string `json:"name"`
	Votes  int64  `json:"votes"`
}

var c *mongo.Client

func createlanguage(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var detail codedetail
	_ = json.NewDecoder(req.Body).Decode(&detail)
	name := strings.ToLower(params["name"])

	fmt.Printf("POST api call made to /languages/%s\n", name)

	lang := language{name, detail}

	id := insertNewLanguage(c, lang)

	if id == nil {
		_ = json.NewEncoder(w).Encode("{'result' : 'insert failed!'}")
	} else {
		err := json.NewEncoder(w).Encode(detail)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func getlanguages(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("GET api call made to /languages")

	var langmap = make(map[string]*codedetail)
	langs, err := returnAllLanguages(c, bson.M{})

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	for _, lang := range langs {
		langmap[lang.Name] = &lang.Detail
	}

	err = json.NewEncoder(w).Encode(langmap)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func getlanguagebyname(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	name := strings.ToLower(params["name"])

	fmt.Printf("GET api call made to /languages/%s\n", name)

	lang, _ := returnOneLanguage(c, bson.M{"name": name})
	if lang == nil {
		_ = json.NewEncoder(w).Encode("{'result' : 'language not found'}")
	} else {
		err := json.NewEncoder(w).Encode(*lang)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func deletelanguagebyname(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	name := strings.ToLower(params["name"])

	fmt.Printf("DELETE api call made to /languages/%s\n", name)

	languagesRemoved := removeOneLanguage(c, bson.M{"name": name})

	_ = json.NewEncoder(w).Encode(fmt.Sprintf("{'count' : %d}", languagesRemoved))
}

func voteonlanguage(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	name := strings.ToLower(params["name"])

	fmt.Printf("GET api call made to /languages/%s/vote\n", name)

	//example using go funcs + channels
	//votesUpdated := updateVote(c, bson.M{"name": name})
	vchan := voteChannel()
	vchan <- name
	voteCount, _ := strconv.ParseInt(<-vchan, 10, 64)
	close(vchan)

	w.Header().Set("Content-Type", "application/json")

	voteresult := voteresult{
		Name: name,
		Votes: voteCount,
	}
	_ = json.NewEncoder(w).Encode(voteresult)
}

func voteChannel() (vchan chan string) {
	//example using go funcs + channels
	vchan = make(chan string)
	go func() {
		name := <-vchan
		//fmt.Println(fmt.Sprintf("name is %s", name))
		voteUpdated := strconv.FormatInt((updateVote(c, bson.M{"name": name})), 10)
		vchan <- voteUpdated
	}()
	return vchan
}

func returnAllLanguages(client *mongo.Client, filter bson.M) ([]*language, error) {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)

	var langs []*language
	collection := client.Database(mongo_db).Collection(mongo_collection)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("error querying documents from database")
	}
	for cur.Next(context.TODO()) {
		var lang language
		err = cur.Decode(&lang)
		if err != nil {
			return nil, errors.New("error on decoding the document")
		}
		langs = append(langs, &lang)
	}
	return langs, nil
}

func returnOneLanguage(client *mongo.Client, filter bson.M) (*language, error) {
	var lang language
	collection := client.Database(mongo_db).Collection(mongo_collection)
	singleResult := collection.FindOne(context.TODO(), filter)
	if singleResult.Err() == mongo.ErrNoDocuments {
		return nil, errors.New("no documents found")
	}
	if singleResult.Err() != nil {
		log.Println("Find error: ", singleResult.Err())
		return nil, singleResult.Err()
	}
	singleResult.Decode(&lang)
	return &lang, nil
}

func insertNewLanguage(client *mongo.Client, lang language) interface{} {
	collection := client.Database(mongo_db).Collection(mongo_collection)
	insertResult, err := collection.InsertOne(context.TODO(), lang)
	if err != nil {
		log.Fatalln("Error on inserting new language", err)
		return nil
	}
	return insertResult.InsertedID
}

func removeOneLanguage(client *mongo.Client, filter bson.M) int64 {
	collection := client.Database(mongo_db).Collection(mongo_collection)
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal("Error on deleting one Hero", err)
	}
	return deleteResult.DeletedCount
}

func updateVote(client *mongo.Client, filter bson.M) int64 {
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	collection := client.Database(mongo_db).Collection(mongo_collection)
	updatedData := bson.M{"$inc": bson.M{"codedetail.votes": 1}}
	updatedResult := collection.FindOneAndUpdate(context.TODO(), filter, updatedData, &opt)
	if updatedResult.Err() != nil {
		log.Fatal("Error on updating language vote count", updatedResult.Err())
	}
	lang := language{}
	_ = updatedResult.Decode(&lang)
	return lang.Detail.Votes
}

//getClient returns a MongoDB Client
func getClient() *mongo.Client {
	mongoconnstr := getEnv("MONGO_CONN_STR", mongo_default_conn_str)
	mongousername := getEnv("MONGO_USERNAME", mongo_default_username)
	mongopassword := getEnv("MONGO_PASSWORD", mongo_default_password)

	fmt.Println("MongoDB connection details:")
	fmt.Println("MONGO_CONN_STR:" + mongoconnstr)
	fmt.Println("MONGO_USERNAME:" + mongousername)
	fmt.Println("MONGO_PASSWORD:")
	fmt.Println("attempting mongodb backend connection...")

	clientOptions := options.Client().ApplyURI(mongoconnstr)

	//test if auth is enabled or expected,
	//for demo purposes when we setup mongo as a replica set using a StatefulSet resource in K8s auth is disabled
	if clientOptions.Auth != nil {
		clientOptions.Auth.Username = mongousername
		clientOptions.Auth.Password = mongopassword
	}

	options.Client().SetMaxConnIdleTime(60000)
	options.Client().SetHeartbeatInterval(5 * time.Second)

	client, err := mongo.NewClient(clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func init() {
	c = getClient()
	err := c.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("couldn't connect to the database", err)
	} else {
		log.Println("connected!!")
	}
}

func ok(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK!\n")
}

func version(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "API vTOKEN_VERSION\n")
}

func cpuDetails(w http.ResponseWriter, req *http.Request) {
	// cpu - get CPU number of cores and speed
	cpuStat, err := cpu.Info()
	if err != nil {
		fmt.Println(err)
	}
	
	runtimeOS := runtime.GOOS

	fmt.Fprintf(w, "OS: %s\n", runtimeOS)
	fmt.Fprintf(w, "CPU index number: %s\n", strconv.FormatInt(int64(cpuStat[0].CPU), 10))
	fmt.Fprintf(w, "CPU index number: %s\n", strconv.FormatInt(int64(cpuStat[0].CPU), 10))
	fmt.Fprintf(w, "VendorID: %s\n", cpuStat[0].VendorID)
	fmt.Fprintf(w, "Family: %s\n", cpuStat[0].Family)
	fmt.Fprintf(w, "Number of cores: %s\n", strconv.FormatInt(int64(cpuStat[0].Cores), 10))
	fmt.Fprintf(w, "Model Name: %s\n", cpuStat[0].ModelName)
	fmt.Fprintf(w, "Speed: %s\n", strconv.FormatFloat(cpuStat[0].Mhz, 'f', 2, 64))
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {
	fmt.Println("version 1.01")
	fmt.Println("serving on port 8080...")
	fmt.Println("tests:")
	fmt.Println("curl -s localhost:8080/ok")
	fmt.Println("curl -s localhost:8080/cpu")
	fmt.Println("curl -s localhost:8080/version")
	fmt.Println("curl -s localhost:8080/languages")
	fmt.Println("curl -s localhost:8080/languages | jq .")

	router := mux.NewRouter()

	//setup routes
	router.HandleFunc("/languages/{name}", createlanguage).Methods("POST")
	router.HandleFunc("/languages", getlanguages).Methods("GET")
	router.HandleFunc("/languages/{name}", getlanguagebyname).Methods("GET")
	router.HandleFunc("/languages/{name}", deletelanguagebyname).Methods("DELETE")
	router.HandleFunc("/languages/{name}/vote", voteonlanguage).Methods("GET")
	router.HandleFunc("/ok", ok).Methods("GET")
	router.HandleFunc("/cpu", cpuDetails).Methods("GET")
	router.HandleFunc("/version", version).Methods("GET")

	//required for CORS - ajax API requests originating from the react browser vote app
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

	//listen on port 8080
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
