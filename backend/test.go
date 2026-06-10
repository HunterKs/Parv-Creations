package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	_, err := bson.ObjectIDFromHex("1234567890abcdef12345678")
	fmt.Println(err == nil)
}
