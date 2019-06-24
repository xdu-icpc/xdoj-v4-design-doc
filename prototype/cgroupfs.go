package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"
)

func ensureCgroupV2(path string) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		log.Panicf("statfs: %v", err)
	}
	if stat.Type != 0x63677270 {
		log.Panicf("magic of fs %s is %d", path, stat.Type)
	}
}

func getOOMCount(dir string) (c int, err error) {
	fn := dir + "/memory.events";
	f, err := os.Open(fn)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var key, value string
	for {
		_, err = fmt.Fscanf(f, "%s%s", &key, &value)
		if err == io.EOF {
			return 0, errors.New("no OOM value in the file")
		}
		if err != nil {
			return 0, err
		}
		if key == "oom" {
			c, err = strconv.Atoi(value)
			return
		}
	}
}
