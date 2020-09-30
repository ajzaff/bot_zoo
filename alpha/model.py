import tensorflow as tf

from tensorflow import keras


class AlphaValueHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(1, (1, 1))
        self.dense = tf.keras.layers.Dense(256)
        self.output = tf.keras.layers.Dense(1, activation='tanh')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = keras.layers.Activation('relu')(x)
        x = self.dense(x)
        return self.output(x)


class AlphaPolicyHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(2, (1, 1))
        self.dense = tf.keras.layers.Dense(256)
        self.output = tf.keras.layers.Dense(1, activation='softmax')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = keras.layers.Activation('relu')(x)
        x = self.dense(x)
        return self.output(x)


class AlphaConvolutionLayer(object):

    def __init__(self, filters, kernel_size, input_shape=None):
        self.conv = tf.keras.layers.Conv2D(
            filters, kernel_size, input_shape=input_shape, padding="same", data_format="channels_last")
        self.norm = tf.keras.layers.BatchNormalization(axis=3)

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        return self.norm(x)


class AlphaResidualLayer(object):

    def __init__(self):
        self.conv1 = AlphaConvolutionLayer(256, (3, 3))
        self.conv2 = AlphaConvolutionLayer(256, (3, 3))

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x_shortcut = inputs
        x = self.conv1(inputs)
        x = tf.keras.layers.Activation('relu')(x)
        x = self.conv2(x)
        return tf.keras.layers.Activation('relu')(x+x_shortcut)


class AlphaModel(keras.Model):

    def __init__(self, residual_layers):
        super(AlphaModel, self).__init__()
        self.conv = AlphaConvolutionLayer(256, (3, 3), input_shape=(8, 8, 26))
        self.residual_layers = [AlphaResidualLayer()
                                for i in range(residual_layers)]
        self.policy = AlphaPolicyHead()
        self.value = AlphaValueHead()

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = tf.keras.layers.Activation('relu')(x)
        for layer in self.residual_layers:
            x = layer(x)
        self.policy(x)
        self.value(x)
        return x


x = tf.keras.Input(shape=(8, 8, 26))
y = tf.keras.layers.Conv2D(
    256, (4, 4), input_shape=(8, 8, 26), padding="same", data_format="channels_last")(x)
y = tf.keras.layers.BatchNormalization(axis=3)(y)
y = tf.keras.layers.Activation('relu')(y)

for i in range(16):
    y_shortcut = y+0
    y = tf.keras.layers.Conv2D(
        256, (4, 4), padding="same", data_format="channels_last")(y)
    y = tf.keras.layers.BatchNormalization(axis=3)(y)
    y = tf.keras.layers.Activation('relu')(y)
    y = tf.keras.layers.Conv2D(
        256, (4, 4), padding="same", data_format="channels_last")(y)
    y = tf.keras.layers.BatchNormalization(axis=3)(y)
    y = tf.keras.layers.Add()([y, y_shortcut])
    y = tf.keras.layers.Activation('relu')(y)

# Value head
y_in = y+0
y1 = tf.keras.layers.Conv2D(
    256, (4, 4), padding="same", data_format="channels_last")(y_in)
y1 = tf.keras.layers.BatchNormalization(axis=3)(y1)
y1 = keras.layers.Activation('relu')(y1)
y1 = tf.keras.layers.Flatten(data_format="channels_last")(y1)
y1 = tf.keras.layers.Dense(256)(y1)
y1 = keras.layers.Activation('relu')(y1)

# Policy head
y2 = tf.keras.layers.Conv2D(
    256, (4, 4), padding="same", data_format="channels_last")(y)
y2 = tf.keras.layers.BatchNormalization(axis=3)(y2)
y2 = keras.layers.Activation('relu')(y2)
y2 = tf.keras.layers.Flatten(data_format="channels_last")(y2)
y2 = tf.keras.layers.Dense(256)(y2)
y2 = keras.layers.Activation('tanh')(y2)
