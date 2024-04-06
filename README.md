ipc-recorder:

- Connect to RTSP server (like a IPCamera)
- fetch and decode video stream
- split into fragment every 10 seconds
- save to files on disk
- save records with start time, end time on database(SQLite on demo), so they can be playback in correct sequence

ipc-player:

- host a HTTP server
- static file server for browser to load web page or video file
- query database for video fragments and generate m3u8 file for playback
- when playing realtime:
    - create a realtime playback session
    - query last 3 fragments as start position
    - wait for new fragments when browser continue load m3u8 file

References:
- https://github.com/bluenviron/gortsplib
- https://github.com/grafov/m3u8
- https://github.com/video-dev/hls.js
- https://developer.apple.com/documentation/http-live-streaming/live-playlist-sliding-window-construction
