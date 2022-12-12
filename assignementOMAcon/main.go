package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
)

var (
	listOrderRe   = regexp.MustCompile(`/orders/`)
	getOrderRe    = regexp.MustCompile(`/orders/:id`)
	createOrderRe = regexp.MustCompile(`/orders/`)
	updateOrderRe = regexp.MustCompile("order/orders/")
)

type order struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	OrderItems  string `json:"order_items,omitempty"`
	TotalItems  string `json:"total_items,omitempty"`
	Payment     string `json:"payment,omitempty"`
	TableNumber string `json:"table_number,omitempty"`
}

type datastore struct {
	m map[string]order
	*sync.RWMutex
}

type orderHandler struct {
	store *datastore
}

func (h *orderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && listOrderRe.MatchString(r.URL.Path):
		h.List(w, r)
		return
	case r.Method == http.MethodGet && getOrderRe.MatchString(r.URL.Path):
		h.Get(w, r)
		return
	case r.Method == http.MethodPost && createOrderRe.MatchString(r.URL.Path):
		h.Create(w, r)
		return
	case r.Method == http.MethodPut && updateOrderRe.MatchString(r.URL.Path):
		h.update(w, r)
		return

	default:
		notFound(w, r)
		return
	}
}

func (h *orderHandler) List(w http.ResponseWriter, r *http.Request) {
	h.store.RLock()
	users := make([]order, 0, len(h.store.m))
	for _, v := range h.store.m {
		users = append(users, v)
	}
	h.store.RUnlock()
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *orderHandler) Get(w http.ResponseWriter, r *http.Request) {
	matches := getOrderRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 1 {
		notFound(w, r)
		return
	}
	orders := make([]order, 0, len(h.store.m))
	h.store.Lock()

	u, ok := h.store.m[matches[1]]

	for _, orderID := range orders {
		if orderID.ID == r.URL.Query().Get("id") {
			_, err := json.Marshal(u)
			if err != nil {
				internalServerError(w, r)
				return
			}
		}
	}

	h.store.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *orderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var u order
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		internalServerError(w, r)
		return
	}
	h.store.Lock()
	h.store.m[u.ID] = u
	h.store.Unlock()
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *orderHandler) update(w http.ResponseWriter, r *http.Request) {
	var u order
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		internalServerError(w, r)
		return
	}

	h.store.Lock()
	for index, item := range h.store.m {
		if item.ID == u.ID {
			h.store.m[index] = u
		}
	}
	h.store.Unlock()

	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}

func main() {
	var wg = sync.WaitGroup{}
	wg.Add(1)

	mux := http.NewServeMux()

	orderH := &orderHandler{
		store: &datastore{
			m: map[string]order{
				"1": {
					ID:          "1",
					Name:        "Rahul",
					OrderItems:  "veg pulav, biryani",
					TotalItems:  "2",
					Payment:     "Done",
					TableNumber: "11",
				},
				"2": {
					ID:          "2",
					Name:        "Mayur",
					OrderItems:  "Pav Bhaji, manchurian",
					TotalItems:  "2",
					Payment:     "Done",
					TableNumber: "123",
				},
				"3": {
					ID:          "3",
					Name:        "Nikhil",
					OrderItems:  "veg pulav",
					TotalItems:  "1",
					Payment:     "Done",
					TableNumber: "12",
				},
				"4": {
					ID:          "4",
					Name:        "Sanajana",
					OrderItems:  "chicken khima,roti",
					TotalItems:  "2",
					Payment:     "pending",
					TableNumber: "1234",
				},
				"5": {
					ID:          "5",
					Name:        "rohit",
					OrderItems:  "pulav",
					TotalItems:  "1",
					Payment:     "pending",
					TableNumber: "1",
				},
			},
			RWMutex: &sync.RWMutex{},
		},
	}

	mux.Handle("/order/", orderH)        // list
	mux.Handle("/orders/", orderH)       // create order
	mux.Handle("/orders/:id", orderH)    // get order by id
	mux.Handle("/order/orders/", orderH) // modify order

	fmt.Println("server started......")

	http.ListenAndServe("localhost:8081", mux)

	wg.Wait()
}
