# bot_alpha_zoo

Bot Alpha Zoo is a strong deep learning RL Arimaa bot trained using a state of the art ResNet. It is a fairly faithful reproduction of the AlphaZero model. The bot uses self play and reinforcement by playing older versions of itself and did not use any expert knowledge besides the rules of the game, and board state representation passed to the network as input.

# Installing

Prebuilt binaries are available on the bot website and in the github releases artifacts section. Otherwise you may build the bot from source after completing required prerequisites.

## Required

* Golang

* Tensorflow (2.0.0b1)

* Protobuf

Annoyingly this doesn't work well with Go modules. `go get` the source and make sure `GOPATH` is set.
 
```
GO111MODULE=off go get github.com/tensorflow/tensorflow
```

Now generate the Go wrapper ops and protos. 

```
$ go generate github.com/tensorflow/tensorflow/tensorflow/go/op
```

If you still have issues you can running:

```bash
$ bash tensorflow/go/genop/generate.sh
```

and manually move the vendored protos into the working tree.

## Recommended

* Docker

## Install the Weights

Best known model weights are hosted on the bot homepage. Use the short script to install them to a local directory.

```
wget https://arimaa.ajz.dev/best-network
```

## GPU Support

Docker GPU runner may be an easy way to get GPU support.

Follow the instructions to install CUDA for Docker support.

Otherwise there may be several more steps to install CUDA, cuDNN, and possibly Tensorflow shared object files from source.

It's important to note each version of Tensorflow expects a rather narrow configuration of CUDA / cuDNN versions.
See the table of [tested configurations](https://www.tensorflow.org/install/source#gpu) first.

You may be able to avoid building Tensorflow with the right combination of library `.so`s. `LIBRARY_PATH` and `LD_LIBRARY_PATH` are your friend for installing multiple versions of CUDA.

# See the games

Training data from the superepochs are available for download on the bot homepage in Protocol Buffer format.