package main

import (
	"fmt"
	"ppe/internal/domain"

	"github.com/google/uuid"
)

func main() {
	state, err := domain.NewPayment("test", 10, "EUR", uuid.New(), uuid.New())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(state)
}
