package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type planets struct {
	Name 				string 			`json:"name"`
	OrderFromSun 		int	 			`json:"orderfromsun"`
	HasRings 			bool			`json:"hasrings"`
}

func main() {
	if err := godotenv.Load("mongodb_uri.env"); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MANGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	//context.TODO() is kinda there for timeout processes
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri)) //connects to mongodb
	if err != nil {
		panic(err)
	}

	defer func() { //defer func waits until nearby func have returned
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}() //nuiance of defer

	router := gin.Default()

	//Returns the databases
	router.GET("/", func(c *gin.Context) {
		databases, err := client.ListDatabaseNames(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(databases)
		c.IndentedJSON(http.StatusOK, databases)
	})

	//Returns the list of all the collections within the database
	router.GET("/:database_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collections, err := client.Database(database_id).ListCollectionNames(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(collections)
		c.IndentedJSON(http.StatusOK, collections)
	})

	//Retrurns the fields used with in a collection
	router.GET("/:database_id/:collection_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")

		collections := client.Database(database_id).Collection(collection_id)

		var result planets
		err = collections.FindOne(context.TODO(), bson.D{}).Decode(&result)
        val := reflect.Indirect(reflect.ValueOf(&result))

		index := reflect.ValueOf(result).NumField()

		fmt.Println("Fields within " + collection_id + ":")

		for i:=0; i < index; i++ {
			field := val.Type().Field(i).Name
			fmt.Println(field)
			c.IndentedJSON(http.StatusOK, field)
		}
	})

	//Returns the list of all the elements within the collection
	router.GET("/:database_id/:collection_id/all", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")

		collections := client.Database(database_id).Collection(collection_id)
		
		cursor, err := collections.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var result bson.D
			if err = cursor.Decode(&result); err != nil {
				log.Fatal(err)
			}
			fmt.Println(result)
			c.IndentedJSON(http.StatusOK, result)
		}
	})

	//Returns all the values in document given a field id
	router.GET("/:database_id/:collection_id/:field_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")
		
		collections := client.Database(database_id).Collection(collection_id)
		
		cursor, err := collections.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var result planets
			if err = cursor.Decode(&result); err != nil {
				log.Fatal(err)
			}
			if (strings.EqualFold(field_id, "Name")) {
				fmt.Println(result.Name)
				c.IndentedJSON(http.StatusOK, result.Name)
			} else if (strings.EqualFold(field_id, "OrderFromSun")) {
				fmt.Println(result.OrderFromSun)
				c.IndentedJSON(http.StatusOK, result.OrderFromSun)
			} else if (strings.EqualFold(field_id, "HasRings")) {
				fmt.Println(result.HasRings)
				c.IndentedJSON(http.StatusOK, result.HasRings)
			} 
		}
	})

	//Returns ONE document { field: 'value'}
	router.GET("/:database_id/:collection_id/:field_id/one=:value_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")
		value_id  := c.Param("value_id")
		
		collections := client.Database(database_id).Collection(collection_id)

		var result bson.D
		filter := bson.D{{field_id, value_id}}
		err := collections.FindOne(context.TODO(), filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "nothing found, check for typos"})
				return
			}
			log.Fatal(err)
		}

		fmt.Println(result)
		c.IndentedJSON(http.StatusOK, result)
	})

	//Returns MULTIPLE documents { field: 'value'}
	router.GET("/:database_id/:collection_id/:field_id/:value_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")

		var num int; var bull bool; var count int

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		collections := client.Database(database_id).Collection(collection_id)

		//uses the conversions above does the appropriate filter
		var filter bson.D
		if (count == 1) {
			filter = bson.D{{field_id, num}}
		} else if (count == 2) {
			filter = bson.D{{field_id, bull}}
		} else {
			filter = bson.D{{field_id, c.Param("value_id")}}
		}

		cursor, err := collections.Find(context.TODO(), filter)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "nothing found, check for typos"})
				return
			}
			log.Fatal(err)
		}
		
		defer cursor.Close(context.TODO())

		var results []bson.D  //gets all the results in the cursor
		err = cursor.All(context.TODO(), &results); if err != nil {
			fmt.Println("Error occured during cursor.ALL")
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error occured during cursor.ALL"})
		}

		//error checking for results with no data returned
		if (results == nil) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "nothing found, check for typos"})
		}

		//loops through the cursor
		for _, result := range results {
			fmt.Println(result)
			c.IndentedJSON(http.StatusOK, result)
		}
	})

	//Returns MULTIPLE documents WITH comparisons ie. $gt $eq $lt
	router.GET("/:database_id/:collection_id/:field_id/:value_id/comparison=:comparison", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")
		comparison := c.Param("comparison")

		var num int;

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Use an integer for comparison"})
		} 

		collections := client.Database(database_id).Collection(collection_id)

		filter := bson.D{{field_id, bson.D{{comparison, num}}}}

		cursor, err := collections.Find(context.TODO(), filter)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "nothing found, check for typos"})
				return
			}
			log.Fatal(err)
		}
		
		defer cursor.Close(context.TODO())

		var results []bson.D  //gets all the results in the cursor
		err = cursor.All(context.TODO(), &results); if err != nil {
			fmt.Println("Error occured during cursor.ALL")
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error occured during cursor.ALL"})
		}

		//error checking for results with no data returned
		if (results == nil) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "nothing found, check for typos"})
		}

		//loops through the cursor
		for _, result := range results {
			fmt.Println(result)
			c.IndentedJSON(http.StatusOK, result)
		}
	})

	//Inserts one document 
	router.POST("/:database_id/:collection_id/add/name=:name/orderfromsun=:orderfromsun/hasrings=:hasrings", func(c *gin.Context){
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")

		var newPlanet planets
		newPlanet.Name = c.Param("name"); 
		newPlanet.OrderFromSun, err = strconv.Atoi(c.Param("orderfromsun"))
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "OrderFromSun is not type int"})
		}
		newPlanet.HasRings, err = strconv.ParseBool(c.Param("hasrings"))
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "HasRings is not type bool"})
		}

		coll := client.Database(database_id).Collection(collection_id)
		// doc := bson.D{{"title", "Record of a Shriveled Datum"}, {"text", "No bytes, no problem. Just insert a document, in MongoDB"}}
		result, err := coll.InsertOne(context.TODO(), newPlanet)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
		c.IndentedJSON(http.StatusOK, result)
	})

	//Deletes a ONE document
	router.DELETE("/:database_id/:collection_id/delete-one/:field_id/:value_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")

		var num int; var bull bool; var count int

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		collections := client.Database(database_id).Collection(collection_id)

		//uses the conversions above does the appropriate filter
		var filter bson.D
		if (count == 1) {
			filter = bson.D{{field_id, num}}
		} else if (count == 2) {
			filter = bson.D{{field_id, bull}}
		} else {
			filter = bson.D{{field_id, c.Param("value_id")}}
		}

		result, err := collections.DeleteOne(context.TODO(), filter)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error deleting ONE document from collections"})
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	//Deletes a MULTIPLE document
	router.DELETE("/:database_id/:collection_id/delete-many/:field_id/:value_id", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")

		var num int; var bull bool; var count int

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		collections := client.Database(database_id).Collection(collection_id)

		//uses the conversions above does the appropriate filter
		var filter bson.D
		if (count == 1) {
			filter = bson.D{{field_id, num}}
		} else if (count == 2) {
			filter = bson.D{{field_id, bull}}
		} else {
			filter = bson.D{{field_id, c.Param("value_id")}}
		}

		result, err := collections.DeleteMany(context.TODO(), filter)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error deleting ONE document from collections"})
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	//Deletes MANY documents WITH a comparison
	router.DELETE("/:database_id/:collection_id/delete-many/:field_id/:value_id/comparison=:comparison", func(c *gin.Context) {
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")
		comparison := c.Param("comparison")

		var num int;

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Use an integer for comparison"})
		} 

		collections := client.Database(database_id).Collection(collection_id)

		filter := bson.D{{field_id, bson.D{{comparison, num}}}}

		result, err := collections.DeleteMany(context.TODO(), filter)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error deleting ONE document from collections"})
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	//Updates ONE document using $set operator
	router.PATCH("/:database_id/:collection_id/:field_id/:value_id/update-one/:field_id2/:value_id2", func(c *gin.Context){
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")

		collections := client.Database(database_id).Collection(collection_id)
		var num int; var bull bool; var count int

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		//uses the conversions above does the appropriate filter -- filter
		var filter bson.D
		if (count == 1) {
			filter = bson.D{{field_id, num}}
		} else if (count == 2) {
			filter = bson.D{{field_id, bull}}
		} else {
			filter = bson.D{{field_id, c.Param("value_id")}}
		}

		field_id = c.Param("field_id2")

		count = 0
		num, err = strconv.Atoi(c.Param("value_id2")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id2")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		//uses the conversions above does the appropriate filter-- update 
		var update bson.D
		if (count == 1) {
			update = bson.D{{"$set", bson.D{{field_id, num}}}}
		} else if (count == 2) {
			update = bson.D{{"$set", bson.D{{field_id, bull}}}}
		} else {
			update = bson.D{{"$set", bson.D{{field_id, c.Param("value_id2")}}}}
		}

		result, err := collections.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error occured updating"})
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	//Updates MANY documents using $set operator
	router.PATCH("/:database_id/:collection_id/:field_id/:value_id/update-many/:field_id2/:value_id2", func(c *gin.Context){
		database_id := c.Param("database_id")
		collection_id := c.Param("collection_id")
		field_id := c.Param("field_id")

		collections := client.Database(database_id).Collection(collection_id)
		var num int; var bull bool; var count int

		//does conversions to int or bool if :value_id is int or bool
		num, err = strconv.Atoi(c.Param("value_id")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		//uses the conversions above does the appropriate filter -- filter
		var filter bson.D
		if (count == 1) {
			filter = bson.D{{field_id, num}}
		} else if (count == 2) {
			filter = bson.D{{field_id, bull}}
		} else {
			filter = bson.D{{field_id, c.Param("value_id")}}
		}

		field_id = c.Param("field_id2")

		count = 0
		num, err = strconv.Atoi(c.Param("value_id2")); if err != nil {
			bull, err = strconv.ParseBool(c.Param("value_id2")); if err == nil {
				count = 2
			}
		} else {
			count = 1
		}

		//uses the conversions above does the appropriate filter-- update 
		var update bson.D
		if (count == 1) {
			update = bson.D{{"$set", bson.D{{field_id, num}}}}
		} else if (count == 2) {
			update = bson.D{{"$set", bson.D{{field_id, bull}}}}
		} else {
			update = bson.D{{"$set", bson.D{{field_id, c.Param("value_id2")}}}}
		}

		result, err := collections.UpdateMany(context.TODO(), filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error occured updating"})
		}

		c.IndentedJSON(http.StatusOK, result)
	})

	
	router.Run("localhost:8888")
}