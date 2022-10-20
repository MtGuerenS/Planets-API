This is a basic rest api that uses mongodb as its database. It collects data about planets. 

The planet struct follows the following schema:

type planets struct {
    	_id                 ObjectID()      
	name 				string
	orderFromSun 		int
	hasRings 			bool
}

To start up the api make sure you have go installed. Once installed, make sure you're in the Planets-Api directory.
Then, to run it:

    go run main.go

Feel free to change the port this is run on. Now its all good to go!
