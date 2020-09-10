module github.com/vishal1132/cafebucks-order

go 1.14

replace github.com/segmentio/kafka-go => /Users/vishal/Desktop/kafka-go
replace github.com/vishal1132/cafebucks => /Users/vishal/work/src/github.com/vishal1132/cafebucks

require (
	github.com/gorilla/mux v1.8.0
	github.com/klauspost/compress v1.11.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/rs/zerolog v1.19.0
	github.com/segmentio/kafka-go v0.4.2
	github.com/valyala/fastjson v1.5.4
	github.com/vishal1132/cafebucks v0.0.0-20200904165619-149e23dbeccb
)
