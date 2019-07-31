package main

func compileAddressWithStreet(state, street, houseNumber string) (address string) {
	if state == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + state
	} else if houseNumber == "" {
		address = "Vilnius, " + state + ", " + street
	} else {
		address = "Vilnius, " + state + ", " + street + " " + houseNumber
	}
	return
}

func compileAddress(state, street string) (address string) {
	if state == "" {
		address = "Vilnius"
	} else if street == "" {
		address = "Vilnius, " + state
	} else {
		address = "Vilnius, " + state + ", " + street
	}
	return
}
