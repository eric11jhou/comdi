package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
)

type Option struct {
	Time string `json:"Time"`
	Put  One    `json:"Put"`
	Call One    `json:"Call"`
}

type One struct {
	Total []interface{}   `json:"Total"`
	Rank  [][]interface{} `json:"Rank"`
}

func main() {
	r := gin.Default()
	r.GET("/option", func(c *gin.Context) {
		jsonFile, err := os.Open("data.json")
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var op Option
		err = json.Unmarshal(byteValue, &op)
		if err != nil {
			fmt.Println(err)
			c.JSON(400, nil)
		}
		c.JSON(200, op)
	})
	r.Run(":8080")
}
