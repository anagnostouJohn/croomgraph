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
	"strings"
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

// var Database *mongo.Database
var uri string

var wg sync.WaitGroup
var setEND bool = false

var db *mongo.Database

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
	db = ConnectToMongo()
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
	// router.POST("/change", ChangeData)
	router.Run(":8080")
	wg.Wait()
	db.Client().Disconnect(ctx)
}

func GetRooms(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	// db := ConnectToMongo()
	col := db.Collection("sensors")
	// col := db.Collection("sensors")
	// defer col.Database().Client().Disconnect(ctx)
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

// func ChangeData(c *gin.Context) {
// 	c.Header("Content-Type", "text/html; charset=utf-8")
// 	c.Header("Access-Control-Allow-Origin", "*")

// 	var data vars.Data

// 	if err := c.ShouldBindJSON(&data); err != nil {

// 		// If there's an error in binding, return a bad request
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	fmt.Println(data)
// 	switch d := data.Value; d {
// 	case "1":
// 		check.SetNew(1, 100) // LIVE
// 	case "2":
// 		check.SetNew(180, 8640) //INFO  DAY 8640 180 RECORDS
// 	case "3":
// 		check.SetNew(180, 8640) // MONTH
// 	case "4":
// 		check.SetNew(262800, 3153600) //INFO YEAR 3153600 RECORDS 262800 EVERY MONTH
// 	}
// 	c.JSON(http.StatusOK, gin.H{"received_value": data.Value})
// }

func SendData(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	room := c.Query("Room")

	// db := ConnectToMongo()
	col := db.Collection(room)
	ctx := context.Background()

	/////FIND ONE TO GET DATA////////////////
	findOptsForone := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	filter := bson.D{}
	var last vars.RoomData
	err := col.FindOne(ctx, filter, findOptsForone).Decode(&last)
	if err != nil {
		fmt.Println(err)
	}

	///////////////////////FIND ALL//////////////////////////
	findOpts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(100)
	cur, err := col.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		c.JSON(500, gin.H{"data": "Error Connected DB"})
		return
	}
	allData := vars.AllData{}
	allData.Room = room

	///////////////////////////////////////////
	sensorOrdered := make([]vars.SensorsRoomData, len(last.Sensors)) //Data of every sensor of the room
	for cur.Next(ctx) {                                              // Runs through Every line of the 100 last lines from the DB. it has a list of the
		//Sensors of the room
		raw := vars.RoomData{}
		if err := cur.Decode(&raw); err != nil {
			log.Fatal(err)
		}

		for i, j := range raw.Sensors { //Runs Through every sensor of the room
			/////// Convert string values to float64 becaouse frontend wants float not string
			/////// Check if is empty string if is emptty sensor is not pressent
			sensorOrdered[i].Lab = j.Lab
			if j.Tc != "" {
				if s, err := strconv.ParseFloat(j.Tc, 32); err == nil {
					roundedValueT := math.Round(s*100) / 100
					sensorOrdered[i].Tc = append(sensorOrdered[i].Tc, roundedValueT)
				}
			}
			if j.Tf != "" {
				if s, err := strconv.ParseFloat(j.Tf, 32); err == nil {
					roundedValueTf := math.Round(s*100) / 100
					sensorOrdered[i].Tf = append(sensorOrdered[i].Tf, roundedValueTf)
				}
			}

			if j.Hf != "" {
				if s, err := strconv.ParseFloat(j.Hf, 32); err == nil {
					roundedValueHf := math.Round(s*100) / 100
					sensorOrdered[i].Hf = append(sensorOrdered[i].Hf, roundedValueHf)
				}
			}

			if j.Hc != "" {
				if s, err := strconv.ParseFloat(j.Hc, 32); err == nil {
					roundedValueHc := math.Round(s*100) / 100
					sensorOrdered[i].Hc = append(sensorOrdered[i].Hc, roundedValueHc)
				}
			}

			if j.Lf != "" {
				if s, err := strconv.ParseFloat(j.Lf, 32); err == nil {
					roundedValueLf := math.Round(s*100) / 100
					sensorOrdered[i].Lf = append(sensorOrdered[i].Lf, roundedValueLf)
				}
			}

			if j.Lc != "" {
				if s, err := strconv.ParseFloat(j.Lc, 32); err == nil {
					roundedValueLc := math.Round(s*100) / 100
					sensorOrdered[i].Lc = append(sensorOrdered[i].Lc, roundedValueLc)
				}
			}
			if j.H != "" {
				if s, err := strconv.ParseFloat(j.H, 32); err == nil {
					roundedValueH := math.Round(s*100) / 100
					sensorOrdered[i].H = append(sensorOrdered[i].H, roundedValueH)
				}
			}
		}

	} ///<<<<<<<<<<<<<<<<<
	///Check if sensor has same values. If yes then has dummy values and then discurt it
	elementsToDelete := make([]vars.ElementsToDelete, len(sensorOrdered))
	for i := len(sensorOrdered) - 1; i >= 0; i-- {
		elementsToDelete[i].Entry = i
		// for i := range sensorOrdered {
		if CheckIfhasSameValues(sensorOrdered[i].Hc) {
			slices.Reverse(sensorOrdered[i].Hc)
		} else {
			fmt.Println("A")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 0)
			// elementsToDelete = append(elementsToDelete, i)
		}
		if CheckIfhasSameValues(sensorOrdered[i].Hf) {
			slices.Reverse(sensorOrdered[i].Hf)
		} else {
			fmt.Println("B")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 1)
			// elementsToDelete = append(elementsToDelete, i)
		}
		if CheckIfhasSameValues(sensorOrdered[i].Lc) {
			slices.Reverse(sensorOrdered[i].Lc)
		} else {
			fmt.Println("C")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 2)
			// elementsToDelete = append(elementsToDelete, i)
		}
		if CheckIfhasSameValues(sensorOrdered[i].Lf) {
			slices.Reverse(sensorOrdered[i].Lf)
		} else {
			fmt.Println("D")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 3)
			// elementsToDelete = append(elementsToDelete, i)
		}
		if CheckIfhasSameValues(sensorOrdered[i].Tc) {
			slices.Reverse(sensorOrdered[i].Tc)
		} else {
			fmt.Println("E")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 4)
			// elementsToDelete = append(elementsToDelete, i)
		}
		if CheckIfhasSameValues(sensorOrdered[i].Tf) {
			slices.Reverse(sensorOrdered[i].Tf)
		} else {
			fmt.Println("F")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 5)
			// elementsToDelete = append(elementsToDelete, i)
		}

		if CheckIfhasSameValues(sensorOrdered[i].H) {
			slices.Reverse(sensorOrdered[i].H)
		} else {
			fmt.Println("G")
			elementsToDelete[i].List = append(elementsToDelete[i].List, 6)
			// elementsToDelete = append(elementsToDelete, i)
		}
		fmt.Println(elementsToDelete)

		// for _, j := range elementsToDelete {
		// 	if len(j.List) > 0 {
		// 		slices.Reverse(j.List)
		// 		for _, k := range j.List {
		// 			if k == 0 {
		// 				sensorOrdered[j.Entry].Hc = sensorOrdered[j.Entry].Hc[:0]
		// 			} else if k == 1 {
		// 				sensorOrdered[j.Entry].Hf = sensorOrdered[j.Entry].Hf[:0]
		// 			} else if k == 2 {
		// 				sensorOrdered[j.Entry].Lc = sensorOrdered[j.Entry].Lc[:0]
		// 			} else if k == 3 {
		// 				sensorOrdered[j.Entry].Lf = sensorOrdered[j.Entry].Lf[:0]
		// 			} else if k == 4 {
		// 				sensorOrdered[j.Entry].Tc = sensorOrdered[j.Entry].Tc[:0]
		// 			} else if k == 5 {
		// 				sensorOrdered[j.Entry].Tf = sensorOrdered[j.Entry].Tf[:0]
		// 			} else if k == 6 {
		// 				sensorOrdered[j.Entry].H = sensorOrdered[j.Entry].H[:0]
		// 			}

		// 		}
		// 	}
		// }

	}
	// fmt.Println(elementsToDelete, "AAAAAAAAAAAAAAAAAAAAAa")
	// allData.SensorsData = sensorOrdered
	allData.SensorsData = sensorOrdered
	for _, j := range allData.SensorsData {
		fmt.Println(j.H, "AAAAAAAAAAAAAAAAaa")
	}

	c.JSON(http.StatusOK, gin.H{
		"data": allData,
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

func CheckIfhasSameValues(slice []float64) bool {
	for _, j := range slice {
		if j != slice[0] {
			return true
		}
	}
	return false
}

// nextSize := 0
// prevsize := 0
// Rooms := []string{}
// finalRoomSensors := []vars.AllData{}
// paramValue := c.Query("valueCalendar")

// db := ConnectToMongo()
// col := db.Collection("sensors")
// // defer col.Database().Client().Disconnect(ctx)

// cur, err := col.Find(ctx, bson.D{})
// if err != nil {
// 	fmt.Println(err)
// }

// for cur.Next(ctx) {
// 	var result bson.M
// 	if err := cur.Decode(&result); err != nil {
// 		log.Fatal(err)
// 	}
// 	// Print the result (each document)
// 	Rooms = append(Rooms, result["name"].(string))
// }
// cur.Close(ctx)
// findOptions := options.Find()
// findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
// findOptions.SetLimit(check.DataToRetreave)

// for _, room := range Rooms {
// 	roomSensorColl := db.Collection(room)
// 	cur, err := roomSensorColl.Find(ctx, bson.M{}, findOptions)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	var ResultsOfRoom []vars.RoomData
// 	if err = cur.All(ctx, &ResultsOfRoom); err != nil {
// 		log.Fatal(err, "FATAL ERROR")
// 	}
// 	RoomSensors := make([]vars.SensorOrdered, len(ResultsOfRoom[0].Sensors))

// 	TempTemp := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 	TempHum := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 	TempHiC := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 	// var TempHiC []float64
// 	for i, j := range ResultsOfRoom[0].Sensors {
// 		RoomSensors[i].Sensor = j.Lab
// 	}
// 	// for _, lineOfEveryResult := range ResultsOfRoom {
// 	var tempRoomData = []vars.RoomData{}
// 	for {

// 		// tempRoomData := []vars.RoomData{}
// 		// TempTemp := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 		// TempHum := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 		// TempHiC := make([][]float64, len(ResultsOfRoom[0].Sensors))
// 		if nextSize > len(ResultsOfRoom) {
// 			fmt.Println(prevsize, "------", check.Step)
// 			prevsize = prevsize - check.Step
// 			tempRoomData = ResultsOfRoom[prevsize:]
// 			nextSize = check.Step
// 			prevsize = 0
// 			setEND = true
// 		} else if nextSize == len(ResultsOfRoom) {

// 			setEND = true
// 			break
// 		} else {
// 			nextSize = nextSize + check.Step
// 			tempRoomData = ResultsOfRoom[prevsize:nextSize]
// 			prevsize = prevsize + check.Step
// 			// setEND = true
// 		}
// 		fmt.Println("EDO2")
// 		for _, j := range tempRoomData { //K

// 			for sensnum, sens := range j.Sensors { //I
// 				if sens.Tc != "" {
// 					if s, err := strconv.ParseFloat(sens.Tc, 32); err == nil {
// 						roundedValueT := math.Round(s*100) / 100
// 						TempTemp[sensnum] = append(TempTemp[sensnum], roundedValueT)
// 					}
// 				}

// 				if sens.Hic != "" {
// 					if s, err := strconv.ParseFloat(sens.Hic, 32); err == nil {
// 						roundedValueHiC := math.Round(s*100) / 100
// 						TempHiC[sensnum] = append(TempHiC[sensnum], roundedValueHiC)
// 					}
// 				}

// 				if sens.H != "" {
// 					if s, err := strconv.ParseFloat(sens.H, 32); err == nil {
// 						roundedValueH := math.Round(s*100) / 100
// 						TempHum[sensnum] = append(TempHum[sensnum], roundedValueH)

// 					}
// 				}
// 			}
// 		}
// 		fmt.Println("EDO1")
// 		for i := 0; i < len(TempTemp); i++ {
// 			var t float64
// 			for _, j := range TempTemp[i] {
// 				t = t + j
// 			}
// 			t = t / float64(check.Step)
// 			RoomSensors[i].Temperature = append(RoomSensors[i].Temperature, t)
// 			TempTemp[i] = TempTemp[i][:0]
// 		}

// 		for i := 0; i < len(TempHum); i++ {
// 			var hu float64
// 			for _, j := range TempHum[i] {
// 				hu = hu + j
// 			}
// 			hu = hu / float64(check.Step)
// 			RoomSensors[i].Humidity = append(RoomSensors[i].Humidity, hu)
// 			TempHum[i] = TempHum[i][:0]
// 		}

// 		for i := 0; i < len(TempHiC); i++ {
// 			var hi float64
// 			for _, j := range TempHiC[i] {
// 				hi = hi + j
// 			}
// 			hi = hi / float64(check.Step)
// 			RoomSensors[i].HeatIndex = append(RoomSensors[i].HeatIndex, hi)
// 			TempHiC[i] = TempHiC[i][:0]
// 		}
// 		if setEND {
// 			tempRoomData = tempRoomData[:0]

// 			fmt.Println("INSIDE", tempRoomData)
// 			break

// 		}
// 	}
// 	fmt.Println("EDO")
// 	for i, rs := range RoomSensors {

// 		T := CheckIfZero(rs.Temperature)
// 		H := CheckIfZero(rs.HeatIndex)
// 		HIC := CheckIfZero(rs.Humidity)
// 		if T {
// 			RoomSensors[i].Temperature = []float64{}
// 		}
// 		if H {
// 			RoomSensors[i].Humidity = []float64{}
// 		}
// 		if HIC {
// 			RoomSensors[i].HeatIndex = []float64{}
// 		}
// 	}

// 	for i := range RoomSensors {
// 		slices.Reverse(RoomSensors[i].HeatIndex)
// 		slices.Reverse(RoomSensors[i].Humidity)
// 		slices.Reverse(RoomSensors[i].Temperature)

// 		fmt.Println(len(RoomSensors[i].Temperature), "<<<<<<<<<<<<<<<<")
// 	}

// 	Final := vars.AllData{Room: room, SensorsData: RoomSensors}
// 	finalRoomSensors = append(finalRoomSensors, Final)
// 	nextSize = 0
// 	prevsize = 0
// 	setEND = false
// }
// c.JSON(http.StatusOK, gin.H{
// 	"data": finalRoomSensors,
// })
// }

func GetData(urls []string) {
	defer wg.Done()
	// d := ConnectToMongo()

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
			col := db.Collection("sensors") ///<<<<<<<<<<<
			rdTrim := strings.TrimSpace(rd.Name)
			filter := bson.D{{Key: "name", Value: rdTrim}}
			upsert := true
			opts := options.UpdateOptions{
				Upsert: &upsert,
			}
			update := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: rdTrim}}}}
			_, err = col.UpdateOne(ctx, filter, update, &opts)

			if err != nil {
				fmt.Println(err)
			}
			currentTime := time.Now()

			// Get epoch time in seconds

			rd.Epoch = currentTime.Unix()
			c := db.Collection(rd.Name)

			if rdTrim != "" {
				id, err := c.InsertOne(ctx, rd)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(id)
			}
			resp.Body.Close()

			// fmt.Println(rd)
		}
		time.Sleep(10 * time.Second)
	}

}

func ConnectToMongo() *mongo.Database {
	// dbURL := os.Getenv("MONGO_URI")
	uri = fmt.Sprintf("mongodb://%s:%s@192.168.23.61:27017/", config.Database.User, config.Database.Password)
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
	Database := client.Database("cRoomData")

	return Database
}
