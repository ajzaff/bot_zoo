package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	zoopb "github.com/ajzaff/bot_zoo/proto"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
)

func main() {
	f, err := os.Open("data/training/games0.pb.snappy")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := snappy.NewReader(f)
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	match := &zoopb.Match{}
	if err := proto.Unmarshal(bs, match); err != nil {
		log.Fatal(err)
	}
	fmt.Println(proto.MarshalTextString(match))
}
