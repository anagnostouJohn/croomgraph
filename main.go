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
	"slices"
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
var setEND bool = false

// var prevsize = 0
// var nextSize = 0

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
		check.SetNew(30, 3000)
	case "3":
		check.SetNew(610, 61000)
	case "4":
		check.SetNew(2592, 259200)
	}
	c.JSON(http.StatusOK, gin.H{"received_value": data.Value})
}

func SendData(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	nextSize := 0
	prevsize := 0
	Rooms := []string{}
	finalRoomSensors := []vars.AllData{}
	fmt.Println("MESA MESA NEO")
	db := ConnectToMongo()
	col := db.Collection("sensors")
	defer col.Database().Client().Disconnect(ctx)
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println(err)
	}

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

	for _, room := range Rooms {
		roomSensorColl := db.Collection(room)
		cur, err := roomSensorColl.Find(ctx, bson.M{}, findOptions)
		if err != nil {
			fmt.Println(err)
		}
		var ResultsOfRoom []vars.RoomData
		if err = cur.All(ctx, &ResultsOfRoom); err != nil {
			log.Fatal(err)
		}
		fmt.Println(len(ResultsOfRoom), nextSize, "AAAAAAAAAAASSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS")
		RoomSensors := make([]vars.SensorOrdered, len(ResultsOfRoom[0].Sensors))

		TempTemp := make([][]float64, len(ResultsOfRoom[0].Sensors))
		TempHum := make([][]float64, len(ResultsOfRoom[0].Sensors))
		TempHiC := make([][]float64, len(ResultsOfRoom[0].Sensors))
		// var TempHiC []float64
		for i, j := range ResultsOfRoom[0].Sensors {
			RoomSensors[i].Sensor = j.Lab
		}
		// for _, lineOfEveryResult := range ResultsOfRoom {
		var tempRoomData = []vars.RoomData{}
		for {

			// tempRoomData := []vars.RoomData{}
			// TempTemp := make([][]float64, len(ResultsOfRoom[0].Sensors))
			// TempHum := make([][]float64, len(ResultsOfRoom[0].Sensors))
			// TempHiC := make([][]float64, len(ResultsOfRoom[0].Sensors))
			if nextSize > len(ResultsOfRoom) {
				fmt.Println(prevsize, "------", check.Step)
				prevsize = prevsize - check.Step
				tempRoomData = ResultsOfRoom[prevsize:]
				nextSize = check.Step
				prevsize = 0
				setEND = true
			} else if nextSize == len(ResultsOfRoom) {

				setEND = true
				break
			} else {
				nextSize = nextSize + check.Step
				tempRoomData = ResultsOfRoom[prevsize:nextSize]
				prevsize = prevsize + check.Step
				// setEND = true
			}

			for _, j := range tempRoomData { //K

				for sensnum, sens := range j.Sensors { //I
					if sens.Tc != "" {
						if s, err := strconv.ParseFloat(sens.Tc, 32); err == nil {
							roundedValueT := math.Round(s*100) / 100
							TempTemp[sensnum] = append(TempTemp[sensnum], roundedValueT)
						}
					}

					if sens.Hic != "" {
						if s, err := strconv.ParseFloat(sens.Hic, 32); err == nil {
							roundedValueHiC := math.Round(s*100) / 100
							TempHiC[sensnum] = append(TempHiC[sensnum], roundedValueHiC)
						}
					}

					if sens.H != "" {
						if s, err := strconv.ParseFloat(sens.H, 32); err == nil {
							roundedValueH := math.Round(s*100) / 100
							TempHum[sensnum] = append(TempHum[sensnum], roundedValueH)

						}
					}
				}
			}

			for i := 0; i < len(TempTemp); i++ {
				var t float64
				for _, j := range TempTemp[i] {
					t = t + j
				}
				t = t / float64(check.Step)
				RoomSensors[i].Temperature = append(RoomSensors[i].Temperature, t)
				TempTemp[i] = TempTemp[i][:0]
			}

			for i := 0; i < len(TempHum); i++ {
				var hu float64
				for _, j := range TempHum[i] {
					hu = hu + j
				}
				hu = hu / float64(check.Step)
				RoomSensors[i].Humidity = append(RoomSensors[i].Humidity, hu)
				TempHum[i] = TempHum[i][:0]
			}

			for i := 0; i < len(TempHiC); i++ {
				var hi float64
				for _, j := range TempHiC[i] {
					hi = hi + j
				}
				hi = hi / float64(check.Step)
				RoomSensors[i].HeatIndex = append(RoomSensors[i].HeatIndex, hi)
				TempHiC[i] = TempHiC[i][:0]
			}

			// for _, temperature := range TempTemp {
			// 	x = x + temperature
			// }
			if setEND {

				// setEND = false
				tempRoomData = tempRoomData[:0]

				fmt.Println("INSIDE", tempRoomData)
				break

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

		for i := range RoomSensors {
			slices.Reverse(RoomSensors[i].HeatIndex)
			slices.Reverse(RoomSensors[i].Humidity)
			slices.Reverse(RoomSensors[i].Temperature)
		}
		Final := vars.AllData{Room: room, SensorsData: RoomSensors}
		finalRoomSensors = append(finalRoomSensors, Final)
		nextSize = 0
		prevsize = 0
		setEND = false
	}
	// fmt.Println(finalRoomSensors, "FINAL FINAL")
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
	dbURL := os.Getenv("MONGO_URI")
	uri = fmt.Sprintf("mongodb://%s:%s@%s", config.Database.User, config.Database.Password, dbURL)
	fmt.Println(uri, "AAAAAAAAAAAAAAAAAAAAAAAAAADDDDDDDDDD")
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
