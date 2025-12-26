package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/dpt"
	"github.com/knx-go/knx-go/pgknx"
)

func (a api) V1APIHandler() (http.Handler, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/group", a.handleGroupList)
	mux.HandleFunc("GET /v1/group/{id}", a.handleGroupGet)
	mux.HandleFunc("GET /v1/group/{id}/latest", a.handleGroupLatest)
	mux.HandleFunc("GET /v1/group/{id}/history", a.handleGroupHistory)
	mux.HandleFunc("GET /v1/group/{id}/read", a.handleGroupRead)
	mux.HandleFunc("POST /v1/group/{id}/write", a.handleGroupWrite)

	return mux, nil
}

type groupInfo struct {
	Name    string   `json:"name"`
	Address string   `json:"address"`
	Path    []string `json:"path,omitempty"`
	DPTs    []string `json:"dpts,omitempty"`
}

func (a api) handleGroupList(w http.ResponseWriter, r *http.Request) {
	groups := a.catalog.Groups()
	infos := make([]groupInfo, 0, len(groups))
	for _, group := range groups {
		info := groupInfo{
			Name:    group.Name,
			Address: group.AddressString(),
		}
		if len(group.Path) > 0 {
			info.Path = append([]string(nil), group.Path...)
		}
		if len(group.DPTs) > 0 {
			info.DPTs = make([]string, len(group.DPTs))
			for i, dpt := range group.DPTs {
				info.DPTs[i] = string(dpt)
			}
		}
		infos = append(infos, info)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(infos); err != nil {
		log.Printf("failed to encode groups: %v", err)
	}
}

func (a api) handleGroupGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var found bool

	for _, group := range a.catalog.Groups() {
		if group.Name != id && group.AddressString() != id {
			continue
		}
		found = true
		info := groupInfo{
			Name:    group.Name,
			Address: group.AddressString(),
		}
		if len(group.Path) > 0 {
			info.Path = append([]string(nil), group.Path...)
		}
		if len(group.DPTs) > 0 {
			info.DPTs = make([]string, len(group.DPTs))
			for i, dpt := range group.DPTs {
				info.DPTs[i] = string(dpt)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			log.Printf("failed to encode groups: %v", err)
		}
	}
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "group not found"}); err != nil {
			log.Printf("failed to encode group error: %v", err)
		}
	}
}

func (a api) handleGroupLatest(w http.ResponseWriter, r *http.Request) {
	// Group
	id := r.PathValue("id")
	found := -1
	for i, group := range a.catalog.Groups() {
		if group.Name != id && group.AddressString() != id {
			continue
		}
		found = i
	}
	if found == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "group not found"}); err != nil {
			log.Printf("failed to encode group error: %v", err)
		}
		return
	}
	group := a.catalog.Groups()[found]

	// Reply
	event, err := a.store.GroupLastEvent(a.ctx, group.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if e := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)}); e != nil {
			log.Printf("failed to encode group error: %v", e)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		log.Printf("failed to encode groups: %v", err)
	}
}

func (a api) handleGroupHistory(w http.ResponseWriter, r *http.Request) {
	// Group
	id := r.PathValue("id")
	found := -1
	for i, group := range a.catalog.Groups() {
		if group.Name != id && group.AddressString() != id {
			continue
		}
		found = i
	}
	if found == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "group not found"}); err != nil {
			log.Printf("failed to encode group error: %v", err)
		}
		return
	}
	group := a.catalog.Groups()[found]

	// Options
	opts := pgknx.HistoryOptions{}
	if v, ok := r.URL.Query()["from"]; ok {
		t, err := time.Parse(time.RFC3339, v[0])
		if err != nil {
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "from is not a valid date"}); err != nil {
				log.Printf("from is not a valid date: %v", err)
			}
			return
		}
		opts.From = t
	}
	if v, ok := r.URL.Query()["to"]; ok {
		t, err := time.Parse(time.RFC3339, v[0])
		if err != nil {
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "to is not a valid date"}); err != nil {
				log.Printf("to is not a valid date: %v", err)
			}
			return
		}
		opts.To = t
	}
	if v, ok := r.URL.Query()["limit"]; ok {
		i, err := strconv.Atoi(v[0])
		if err != nil {
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "limit is not a valid int"}); err != nil {
				log.Printf("limit is not a valid int: %v", err)
			}
			return
		}
		opts.Limit = i
	}
	if v, ok := r.URL.Query()["offset"]; ok {
		i, err := strconv.Atoi(v[0])
		if err != nil {
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "offset is not a valid int"}); err != nil {
				log.Printf("offset is not a valid int: %v", err)
			}
			return
		}
		opts.Offset = i
	}
	if v, ok := r.URL.Query()["desc"]; ok {
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "desc is not a valid bool"}); err != nil {
				log.Printf("desc is not a valid bool: %v", err)
			}
			return
		}
		opts.Descending = b
	}

	// Reply
	events, err := a.store.GroupEventsHistory(a.ctx, group.Name, opts)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if e := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)}); e != nil {
			log.Printf("failed to encode group error: %v", e)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Printf("failed to encode groups: %v", err)
	}
}

func (a api) handleGroupWrite(w http.ResponseWriter, r *http.Request) {
	// Group
	id := r.PathValue("id")
	found := -1
	for i, group := range a.catalog.Groups() {
		if group.Name != id && group.AddressString() != id {
			continue
		}
		found = i
	}
	if found == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "group not found"}); err != nil {
			log.Printf("failed to encode group error: %v", err)
		}
		return
	}
	group := a.catalog.Groups()[found]

	// Decode value
	type Payload struct {
		Value string `json:"value"`
	}
	var p Payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Send event
	payload, err := dpt.EncodeDPTFromStringN(string(group.DPTs[0]), p.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if e := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)}); e != nil {
			log.Printf("the input string is not valid: %v", e)
		}
		return
	}

	event := knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: group.Address,
		Data:        payload,
	}
	if err := a.tunnel.Send(event); err != nil {
		fmt.Printf("Error while sending: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); e != nil {
		log.Printf("failed to return success")
	}
}

func (a api) handleGroupRead(w http.ResponseWriter, r *http.Request) {
	// Group
	id := r.PathValue("id")
	found := -1
	for i, group := range a.catalog.Groups() {
		if group.Name != id && group.AddressString() != id {
			continue
		}
		found = i
	}
	if found == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "group not found"}); err != nil {
			log.Printf("failed to encode group error: %v", err)
		}
		return
	}
	group := a.catalog.Groups()[found]

	t := time.Now()

	// Send read event
	re := knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: group.Address,
	}
	if err := a.tunnel.Send(re); err != nil {
		fmt.Printf("Error while sending: %v\n", err)
		return
	}

	// Read reply
	var event pgknx.Event
	var err error
	for range 10 {
		time.Sleep(time.Second)
		event, err = a.store.GroupLastEvent(a.ctx, group.Name)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if e := json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)}); e != nil {
				log.Printf("failed to read the return value with error: %v", e)
			}
			return
		}
		if event.Timestamp.After(t) {
			break
		}
	}
	if !event.Timestamp.After(t) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if e := json.NewEncoder(w).Encode(map[string]string{"error": "no value returned before the timeout"}); e != nil {
			log.Printf("no value returned before the timeout")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		log.Printf("failed to encode groups: %v", err)
	}
}
