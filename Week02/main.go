package main

import (
	"fmt"

	"week02/service"
)

func main() {
	user, err := service.GetInfo(1)
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}

	fmt.Println(user)
}
