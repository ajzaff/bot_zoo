package main

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	corepb "github.com/tensorflow/tensorflow/tensorflow/go/core/protobuf/for_core_protos_go_proto"
)

func main() {
	bs, err := ioutil.ReadFile("data/saved_models/bot_alpha_zoo-16/saved_model.pb")
	if err != nil {
	}

	model := &corepb.SavedModel{}
	if err := proto.Unmarshal(bs, model); err != nil {
		panic(err)
	}

	fmt.Println(proto.MarshalTextString(model))
}
