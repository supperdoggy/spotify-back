ffmpeg -i $1 -c:a libmp3lame -b:a 128k -map 0:0 -f segment -segment_time 10 -segment_list $2 -segment_format mpegts $3