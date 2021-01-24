package product

import (
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
)

type productNotification struct {
	Action  string  `json:"action"`
	Product Product `json:"product"`
}

type inMessage struct {
	Cmd        string `json:"command"`
	ProductIDs []int  `json:"productIDs"`
}

type chanSubscriptionCmd struct {
	Cmd        string
	ProductIDs []int
	CommChan   chan *Product
}

var prodChan = make(chan productNotification, 1024)
var addRemoveSubscription = make(chan chanSubscriptionCmd)

func handleDistributionGoroutine() {

	notifyUpdates := make(map[int]map[chan *Product]bool)

	for {

		select {

		case incomingProduct := <-prodChan:

			notifChans, exists := notifyUpdates[incomingProduct.Product.ProductID]
			if exists && notifChans != nil {
				for k, v := range notifChans {
					if v {
						// this will block all threads in case of a single slow reader
						// the chan will fill and it will not be possible to send other
						// notifications to other readers.
						// option is to do launch each as a separate goroutine
						// but it will not guarantee order at the receiving side
						go func() { k <- &incomingProduct.Product }()
					}
				}
			}

		case subscription := <-addRemoveSubscription:

			switch subscription.Cmd {

			case "subscribe":
				for _, prd := range subscription.ProductIDs {
					ret, ok := notifyUpdates[prd]
					if !ok {
						ret = make(map[chan *Product]bool)
						notifyUpdates[prd] = ret
					}
					ret[subscription.CommChan] = true
				}
			case "unsubscribe":
				for _, prd := range subscription.ProductIDs {
					delete(notifyUpdates, prd)
				}

			case "closeconn":

				// empty the rest
				emptyKeys := make([]int, 0, 100)

				for k, v := range notifyUpdates {
					delete(v, subscription.CommChan)
					if len(v) == 0 {
						emptyKeys = append(emptyKeys, k)
					}
				}
				// clear the map of empty keys
				for _, k := range emptyKeys {
					delete(notifyUpdates, k)
				}
				close(subscription.CommChan)

			default:
				log.Printf("Unhandled command %v", subscription.Cmd)
			}
		}
	}
}

func init() {
	go handleDistributionGoroutine()
}

func HandleChangeProductNotification(msg []byte) {
	log.Println("Product changed: ", string(msg))

	productNotify := productNotification{}

	if err := json.Unmarshal(msg, &productNotify); err != nil {
		log.Println(err)
		return
	}

	prodChan <- productNotify
}

func productChangeWSHandler(ws *websocket.Conn) {

	// make the chan buffered so we can receive more messages until we process them
	inMsgChan := make(chan inMessage, 1024)
	inProductsUpdated := make(chan *Product, 1024)

	defer func() {
		addRemoveSubscription <- chanSubscriptionCmd{
			Cmd:      "closeconn",
			CommChan: inProductsUpdated,
		}

		// drain the channel
		for range inProductsUpdated {
		}

	}()

	go func(ws *websocket.Conn) {
		for {
			ws.MaxPayloadBytes = 1024 * 256
			var msg inMessage

			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				log.Println(err)
				break
			}
			inMsgChan <- msg
		}
		close(inMsgChan)
	}(ws)

	for {
		select {
		case msg, ok := <-inMsgChan:
			// subscribe - unsubscribe
			if !ok {
				return // connection close
			} else {

				addRemoveSubscription <- chanSubscriptionCmd{
					Cmd:        msg.Cmd,
					ProductIDs: msg.ProductIDs,
					CommChan:   inProductsUpdated,
				}

			}
		case product, ok := <-inProductsUpdated:
			// updated products
			if !ok {
				return
			}

			if err := websocket.JSON.Send(ws, product); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
