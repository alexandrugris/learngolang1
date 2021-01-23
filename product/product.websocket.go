package product

import "log"

func HandleChangeProductNotification(msg []byte) {

	log.Println(string(msg))

}
