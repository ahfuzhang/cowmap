
init:
	go mod init github.com/ahfuzhang/cowmap

benchmark:
	go test -benchmem -run=^$ -bench ^Benchmark_all$ github.com/ahfuzhang/cowmap
		