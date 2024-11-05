package main

import (
	"context"
	vars "croomgraph/VARS"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
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

var TempHum []float64
var TempTemp []float64
var TempHiC []float64

var check vars.ListCounter

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
	check.SetNew(1, 100)

}

func main() {

	wg.Add(1)
	go GetData(config.Sensors.Ips)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // You can specify allowed origins here
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	router.GET("/data", SendData)
	router.GET("/getrooms", GetRooms)
	router.POST("/change", ChangeData)
	router.Run(":8080")
	wg.Wait()
}

func GetRooms(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	db := ConnectToMongo()
	col := db.Collection("sensors")
	// col := db.Collection("sensors")
	defer col.Database().Client().Disconnect(ctx)
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println(err)
	}
	Rooms := []string{}
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Print the result (each document)
		Rooms = append(Rooms, result["name"].(string))
	}
	// fmt.Println(Rooms)

	c.JSON(http.StatusOK, gin.H{
		"data": Rooms,
	})
}

func ChangeData(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")

	var data vars.Data

	if err := c.ShouldBindJSON(&data); err != nil {

		// If there's an error in binding, return a bad request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(data)
	switch d := data.Value; d {
	case "1":
		check.SetNew(1, 100)
	case "2":
		check.SetNew(9, 9000)
	case "3":
		check.SetNew(61, 61000)
	case "4":
		check.SetNew(259, 259200)
	}
	c.JSON(http.StatusOK, gin.H{"received_value": data.Value})
}

func SendData(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	db := ConnectToMongo()

	col := db.Collection("sensors")
	defer col.Database().Client().Disconnect(ctx)
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println(err)
	}
	Rooms := []string{}
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		// Print the result (each document)
		Rooms = append(Rooms, result["name"].(string))
	}

	cur.Close(ctx)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(check.DataToRetreave)
	finalRoomSensors := []vars.AllData{}
	for _, r := range Rooms {
		SensorColl := db.Collection(r)
		cur, err := SensorColl.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			fmt.Println(err)
		}
		var ResultsOfRoom []vars.RoomData
		if err = cur.All(ctx, &ResultsOfRoom); err != nil {
			log.Fatal(err)
		}
		RoomSensors := make([]vars.SensorOrdered, len(ResultsOfRoom[0].Sensors))
		for i, j := range ResultsOfRoom[0].Sensors {
			RoomSensors[i].Sensor = j.Lab
		}
		for _, lineOfEveryResult := range ResultsOfRoom {
			for i, sens := range lineOfEveryResult.Sensors {

				fmt.Println(check.Av, check.CountForAv)
				time.Sleep(100 * time.Millisecond)
				if sens.Hic != "" {
					if s, err := strconv.ParseFloat(sens.Hic, 32); err == nil {
						roundedValue := math.Round(s*100) / 100
						TempHiC = append(TempHiC, roundedValue)
						if check.Check() {
							fmt.Println(len(TempHiC), "HIC")
							// fmt.Println(countForAv, av, "<<<<<<<<<<<<<<<<<<<")
							var x float64
							for _, hic := range TempHiC {
								x = x + hic
							}
							x = x / float64(len(TempHiC))
							TempHiC = TempHiC[:0]
							RoomSensors[i].HeatIndex = append(RoomSensors[i].HeatIndex, x)
						}

					}
				}
				if sens.Tc != "" {
					if s, err := strconv.ParseFloat(sens.Tc, 32); err == nil {
						roundedValue := math.Round(s*100) / 100
						TempTemp = append(TempTemp, roundedValue)
						// fmt.Println()
						if check.Check() {
							fmt.Println(TempTemp)
							fmt.Println(len(TempTemp), "TEMP")
							var x float64
							for _, temperature := range TempTemp {
								x = x + temperature
							}
							x = x / float64(len(TempTemp))
							TempTemp = TempTemp[:0]
							RoomSensors[i].Temperature = append(RoomSensors[i].Temperature, x)
						}
					}
				}
				if sens.H != "" {
					if s, err := strconv.ParseFloat(sens.H, 32); err == nil {
						roundedValue := math.Round(s*100) / 100
						TempHum = append(TempHum, roundedValue)
						if check.Check() {
							fmt.Println(len(TempHum), "HUM")
							var x float64
							for _, humidity := range TempHum {
								x = x + humidity
							}
							x = x / float64(len(TempHum))
							TempHum = TempHum[:0]
							RoomSensors[i].Humidity = append(RoomSensors[i].Humidity, x)
						}
					}
				}
				check.Increase()
				if check.Check() {
					fmt.Println("INSIDE")
					check.SetZero()
				}
			}
		}

		for i, rs := range RoomSensors {
			T := CheckIfZero(rs.Temperature)
			H := CheckIfZero(rs.HeatIndex)
			HIC := CheckIfZero(rs.Humidity)
			if T {
				RoomSensors[i].Temperature = []float64{}
			}
			if H {
				RoomSensors[i].Humidity = []float64{}
			}
			if HIC {
				RoomSensors[i].HeatIndex = []float64{}
			}
		}
		Final := vars.AllData{Room: r, SensorsData: RoomSensors}
		finalRoomSensors = append(finalRoomSensors, Final)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": finalRoomSensors,
	})
}

func CheckIfZero(list []float64) bool {

	for _, v := range list {
		if v != 0 {
			return false
		}
	}
	return true
}

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
