package main

import (
	"ppe/internal/domain"

	"github.com/google/uuid"
)

func main() {
	var state domain.Payment = domain.Payment{
		ID: uuid.New(),
	}

	state.TransitionTo(domain.StatusRefunded)
}
