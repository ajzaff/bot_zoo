# bot_alpha_zoo

Bot Alpha Zoo is a strong deep learning RL Arimaa bot trained using a state of the art ResNet. It is a fairly faithful reproduction of the AlphaZero model. The bot uses self play and reinforcement by playing older versions of itself and did not use any expert knowledge besides the rules of the game, and board state representation passed to the network as input.

# Installing

Prebuilt binaries are available on the bot website and in the github releases artifacts section. Otherwise you may build the bot from source after completing required prerequisites.

## Required

* Golang

* Tensorflow

Annoyingly this doesn't work well with Go modules. `go get` the source and make sure `GOPATH` is set.
 
```
GO111MODULE=off go get github.com/tensorflow/tensorflow
```

As of writing, the Tensorflow authors have not intended some protos to be available in Go including the Example and Feature protos we require for building training data. Let's add this line to the script to generate those:

```diff
@@ -64,6 +64,7 @@ export PATH=$PATH:${GOPATH}/bin
 mkdir -p ../vendor
 for FILE in ${TF_DIR}/tensorflow/core/framework/*.proto \
     ${TF_DIR}/tensorflow/core/protobuf/*.proto \
+    ${TF_DIR}/tensorflow/core/example/*.proto \
     ${TF_DIR}/tensorflow/stream_executor/*.proto; do
   ${PROTOC} \
     -I ${TF_DIR} \
```

Now generate the Go wrapper ops and protos. 

```
$ go generate github.com/tensorflow/tensorflow/tensorflow/go/op
$ bash tensorflow/go/genop/generate.sh
```

The protos will have been outputted to a vendor directory. Since govendor is no longer recommended, simply copy the `example_protos_go_proto` files into the work tree. The protos should now be available to use.

If you are not using the recommended dockerized runner, and wish to have GPU support, you may need further steps to build Tensorflow from Source. Instructions are available here.

## Recommended

* Docker

## Install the Weights

Best known model weights are hosted on the bot homepage. Use the short script to install them to a local directory.

```
wget https://arimaa.ajz.dev/best-network
```

## GPU Support

The dockerized runner is the easiest way to get GPU inference and training. Otherwise there may be several more steps to install CUDA, cuDNN, and possibly Tensorflow shared object files from source. If you wish to do this instructions are available.

Follow the instructions to install CUDA for Docker support.

# See the games

Training data from the superepochs are available for download on the bot homepage in Protocol Buffer format.