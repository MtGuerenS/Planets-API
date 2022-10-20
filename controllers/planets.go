package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"planetsAPI/configs"
)

// Returns the list of all the collections within the database
func GetCollections() gin.HandlerFunc {
	return func(c *gin.Context) {
		collections, err := configs.CLIENT.Database("sample_guides").ListCollectionNames(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(collections)
		c.IndentedJSON(http.StatusOK, collections)
	}
}

// Returns the list of all the elements within the collection
func GetAllDocuments() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection_id := c.Param("collection_id")

		collections := configs.GetCollection(configs.CLIENT, collection_id)
		
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
	}
}

/*	Returns MULTIPLE documents { field: 'value'}
	can use &compare=($gt, $lt, etc.) 				*/
func SearchesCollection() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection_id := c.Param("collection_id")
		
		// map of all query params: map[string][]string == /?name=__&orderFromSun=___
		queryParams := c.Request.URL.Query()

		// builds query key value pairs
		query := bson.M{}
		for key, _ := range queryParams {
			if (key == "compare" || key == "_id") { 	//special cases
				if (key == "_id") {
					// parse object id from id parameter
					id, err := primitive.ObjectIDFromHex(c.Query(key))
					if err != nil{
					c.JSON(http.StatusBadRequest, gin.H{"message": "Could not convert ObjectIDFromHex"})
						return
					}
					query[key] = id
				}
			} else {
				num, err := strconv.Atoi(c.Query(key)); if err != nil {		// conversion to type int, bool, or string
					bull, err1 := strconv.ParseBool(c.Query(key)); if err1 != nil {
						query[key] = c.Query(key)
					} else {
						query[key] = bull
					}
				} else {
					value, ok := queryParams["compare"]; if ok { 			// checks to see if compare is being used and if its being used its value is added to whatever number calls it
						s := regexp.MustCompile(" ").Split(value[0], 2)		// uses regex to split compare value into two --will be used to error check
						if (s[1] == key) {
							query[key] = bson.D{{s[0], num}} 	// bson.D{{compare operator, num}}
						} else {
							c.IndentedJSON(http.StatusConflict, gin.H{"message": "Double check that compare follows: compare=(operator) (field name)"})
							return
						}
					} else {
						query[key] = num
					}
				}
			}
		}

		collections := configs.GetCollection(configs.CLIENT, collection_id)

		cursor, err := collections.Find(context.TODO(), query)
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
	}
}

// Inserts one planet 
func InsertOnePlanet() gin.HandlerFunc {
	return func(c *gin.Context){
		collection_id := c.Param("collection_id")

		coll := configs.GetCollection(configs.CLIENT, collection_id)
		
		orderFromSun, err := strconv.Atoi(c.Query("orderFromSun"))
		hasRings, err := strconv.ParseBool(c.Query("hasRings"))

		newPlanet := bson.M{"name": c.Query("name"), "orderFromSun": orderFromSun, "hasRings": hasRings}
		
		result, err := coll.InsertOne(context.TODO(), newPlanet)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Double check the query for any errors"})
		}

		fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
		c.IndentedJSON(http.StatusOK, newPlanet)
	}
}

// Deletes any planets that match the conditions {field: value}
func DeletePlanets() gin.HandlerFunc {
	return func(c *gin.Context) {
		collection_id := c.Param("collection_id")

		// map of all query params: map[string][]string == /?name=__&orderFromSun=___
		queryParams := c.Request.URL.Query()

		// builds query key value pairs
		query := bson.M{}
		for key, _ := range queryParams {
			if (key == "compare") {		// special cases
				if (key == "_id") {
					// parse object id from id parameter
					id, err := primitive.ObjectIDFromHex(c.Query(key))
					if err != nil{
					c.JSON(http.StatusBadRequest, gin.H{"message": "Could not convert ObjectIDFromHex"})
						return
					}
					query[key] = id
				}
			} else {
				num, err := strconv.Atoi(c.Query(key)); if err != nil {		// conversion to type int, bool, or string
					bull, err1 := strconv.ParseBool(c.Query(key)); if err1 != nil {
						query[key] = c.Query(key)
					} else {
						query[key] = bull
					}
				} else {
					value, ok := queryParams["compare"]; if ok { 			// checks to see if compare is being used and if its being used its value is added to whatever number calls it
						s := regexp.MustCompile(" ").Split(value[0], 2)		// uses regex to split compare value into two --will be used to error check
						if (s[1] == key) {
							query[key] = bson.D{{s[0], num}} 	// bson.D{{compare operator, num}}
						} else {
							c.IndentedJSON(http.StatusConflict, gin.H{"message": "Double check that compare follows: compare=(operator) (field name)"})
							return
						}
					} else {
						query[key] = num
					}
				}
			}
		}

		collections := configs.GetCollection(configs.CLIENT, collection_id)

		result, err := collections.DeleteMany(context.TODO(), query)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error deleting ONE document from collections"})
		}

		c.IndentedJSON(http.StatusOK, result)
	}
}

/*	REQUIRES _id=<<planet id>> TO UPDATE
	Updates a SINGULAR value that matches with the initial conditions of the search --uses $set operator */
func UpdatePlanets() gin.HandlerFunc {
	return func(c *gin.Context){
		collection_id := c.Param("collection_id")

		collections := configs.GetCollection(configs.CLIENT, collection_id)

		// map of all query params: map[string][]string == /?name=__&orderFromSun=___
		queryParams := c.Request.URL.Query()

		// builds query key value pairs
		query := bson.M{}
		for key, _ := range queryParams {
			if (key == "compare" || key == "_id") {		//special cases
				if (key == "_id") {
					// parse OnjectID() from id parameter
					id, err := primitive.ObjectIDFromHex(c.Query(key))
					if err != nil{
					c.JSON(http.StatusBadRequest, gin.H{"message": "Could not convert ObjectIDFromHex"})
						return
					}
					query[key] = id
				}
			} else {
				num, err := strconv.Atoi(c.Query(key)); if err != nil {		// conversion to type int, bool, or string
					bull, err1 := strconv.ParseBool(c.Query(key)); if err1 != nil {
						query[key] = c.Query(key)
					} else {
						query[key] = bull
					}
				} else {
					value, ok := queryParams["compare"]; if ok { 			// checks to see if compare is being used and if its being used its value is added to whatever number calls it
						s := regexp.MustCompile(" ").Split(value[0], 2)		// uses regex to split compare value into two --will be used to error check
						if (s[1] == key) {
							query[key] = bson.D{{s[0], num}} 	// bson.D{{compare operator, num}}
						} else {
							c.IndentedJSON(http.StatusConflict, gin.H{"message": "Double check that compare follows: compare=(operator) (field name)"})
							return
						}
					} else {
						query[key] = num
					}
				}
			}
		}

		// uses the ObjectID() established
		filter := bson.M{"_id": query["_id"]}

		update := bson.M{"$set": query}

		result, err := collections.UpdateMany(context.TODO(), filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "Error occured updating"})
		}

		c.IndentedJSON(http.StatusOK, result)
	}
}