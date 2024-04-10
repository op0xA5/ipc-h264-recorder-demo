package main

import (
	"encoding/json"
	"fmt"
	"github.com/grafov/m3u8"
	"ipc-player/schema"
	"net/http"
	"strconv"
	"time"
)

func main() {
	err := initStorage()
	if err != nil {
		panic(err)
	}

	http.Handle("/api/playback", http.HandlerFunc(playbackHandler))
	http.Handle("/api/realtime", http.HandlerFunc(realtimeHandler))
	http.Handle("/", http.FileServer(http.Dir("./web")))

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}

func playbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	startMs, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	endMs, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	res, err := QueryRecordByTimeRange("ipc01", "1", time.UnixMilli(startMs), time.UnixMilli(endMs))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	playlist, err := m3u8.NewMediaPlaylist(0, uint(len(res)))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	playlist.MediaType = m3u8.VOD
	playlist.TargetDuration = 10
	playlist.Closed = true
	for _, record := range res {
		err := playlist.Append("/"+record.FileURL, record.Interval, "")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	playlist.Encode().WriteTo(w)

}

func realtimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		realtime := &schema.Realtime{
			Device: "ipc01",
			Stream: "1",
			SeqNo:  1,
		}
		err := CreateRealtime(realtime)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		res, err := json.Marshal(realtime)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write(res)
		return
	}
	if r.Method == http.MethodGet {
		var id = r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		realtime, err := GetRealtime(id)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if realtime == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		winsize := 3
		maxWaitDuration := 60 * time.Second
		waitEnd := time.Now().Add(maxWaitDuration)

		var res []*schema.Record
		for {
			if realtime.LastStartAt == nil {
				res, err = QueryRecordLast(realtime.Device, realtime.Stream, winsize)
			} else {
				res, err = QueryRecordLastStartAt(realtime.Device, realtime.Stream, *realtime.LastStartAt, winsize)
			}
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if len(res) == 0 {
				if time.Now().After(waitEnd) {
					break
				}
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		if len(res) == 0 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if len(res) > winsize {
			res = res[:winsize]
		}

		playlist, err := m3u8.NewMediaPlaylist(uint(len(res)), uint(len(res)))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		playlist.SeqNo = uint64(realtime.SeqNo)
		playlist.TargetDuration = 10
		for _, record := range res {
			err := playlist.Append("/"+record.FileURL, record.Interval, "")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			fmt.Println(record.ID)
		}

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		playlist.Encode().WriteTo(w)

		realtime.LastStartAt = &res[len(res)-1].StartAt
		realtime.SeqNo++
		SaveRealtime(realtime)
	}
}
