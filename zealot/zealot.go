package main

import "probe"

func main() {
	uid, err := probe.ReadFavList()
	if err != nil {
		panic(err)
	}

	err = uid.Run()
	if err != nil {
		panic(err)
	}
}
