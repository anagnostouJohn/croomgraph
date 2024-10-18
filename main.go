package main

import (
	"context"
	vars "croomgraph/VARS"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.TODO()
var config vars.RoomConfig
var Database *mongo.Database
var uri string

var wg sync.WaitGroup

func init() {

	file, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	// Create a variable to hold the configuration
	// Unmarshal the TOML data into the config struct
	err = toml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	wg.Add(1)
	go GetData(config.Sensors.Ips)

	router := gin.Default()

	router.GET("/data", SendData)
	router.Run(":8080")
	wg.Wait()
}

func SendData(c *gin.Context) {
	d := ConnectToMongo()
	col := d.Collection("sensors")
	// filter :=
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println(err)
	}
	sensorName := []string{}
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Print the result (each document)
		sensorName = append(sensorName, result["name"].(string))
	}
	cur.Close(ctx)
	// col := d.Collection("sensors")
	// retrData := []vars.RoomData{}
	fmt.Println(sensorName)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(100)
	results := [][]bson.M{}

	for _, s := range sensorName {
		SensorColl := d.Collection(s)
		cur, err := SensorColl.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			fmt.Println(err)
		}

		// Iterate through the results
		var sensor []bson.M

		if err = cur.All(ctx, &sensor); err != nil {
			log.Fatal(err)
		}
		cur.Close(ctx)
		results = append(results, sensor)
		// fmt.Println(results)

	}
	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
	// cur, err := collection.Find(ctx, bson.D{{}}, findOptions)
}

// findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
// 	findOptions.SetLimit(100)

func GetData(urls []string) {
	defer wg.Done()
	d := ConnectToMongo()
	for {
		for _, url := range urls {

			fmt.Println(config.Sensors)
			resp, err := http.Get(fmt.Sprintf("http://%s/getData.json", url))
			if err != nil {
				log.Fatalf("Error occurred while fetching data from API: %v", err)
			}

			// Read the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("Error reading response body: %v", err)
			}
			rd := vars.RoomData{}
			err = json.Unmarshal(body, &rd)
			if err != nil {
				fmt.Println(err)
			}
			col := d.Collection("sensors")
			filter := bson.D{{Key: "name", Value: rd.Name}}
			upsert := true
			opts := options.UpdateOptions{
				Upsert: &upsert,
			}
			update := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: rd.Name}}}}
			_, err = col.UpdateOne(ctx, filter, update, &opts)

			if err != nil {
				fmt.Println(err)
			}

			c := d.Collection(rd.Name)
			id, err := c.InsertOne(ctx, rd)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(id)
			resp.Body.Close()

			// fmt.Println(rd)
		}
		time.Sleep(10 * time.Second)
	}

}

func ConnectToMongo() *mongo.Database {
	uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", config.Database.User, config.Database.Password, config.Database.Server, config.Database.Port)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	var result bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	Database = client.Database("cRoomData")
	return Database
}
