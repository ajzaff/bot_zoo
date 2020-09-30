import tensorflow as tf

from tensorflow import keras
from tensorflow.keras import layers


class AlphaConvolutionLayer(object):

    def __init__(self, filters, kernel_size, padding='valid', activation=None):
        self.conv = layers.Conv2D(
            filters, kernel_size, padding=padding, data_format='channels_last')
        self.norm = layers.BatchNormalization(axis=3)

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.norm(x)
        return layers.Activation('relu')(x)


class AlphaResidualLayer(object):

    def __init__(self):
        self.conv1 = AlphaConvolutionLayer(
            256, (3, 3), padding='same', activation='relu')
        self.conv2 = AlphaConvolutionLayer(256, (3, 3), padding='same')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x_shortcut = inputs+0
        x = self.conv1(inputs)
        x = self.conv2(x)
        x = layers.Add()([x, x_shortcut])
        return layers.Activation('relu')(x)


class AlphaValueHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(1, (1, 1), activation='relu')
        self.flatten = layers.Flatten(data_format='channels_last')
        self.dense = layers.Dense(256, activation='relu')
        self.output = layers.Dense(1, activation='tanh')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.flatten(x)
        x = self.dense(x)
        return self.output(x)


class AlphaPolicyHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(2, (1, 1), activation='relu')
        self.flatten = layers.Flatten(data_format='channels_last')
        self.dense = layers.Dense(256, activation='relu')
        self.output = layers.Dense(225, activation='linear')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = AlphaConvolutionLayer(2, (1, 1))(y)
        x = self.flatten(x)
        return self.output(x)


class AlphaModel(keras.Model):

    def __init__(self, residual_layers):
        super(AlphaModel, self).__init__()
        self.conv = AlphaConvolutionLayer(256, (4, 4))
        self.residual_layers = [AlphaResidualLayer()
                                for i in range(residual_layers)]
        self.policy = AlphaPolicyHead()
        self.value = AlphaValueHead()

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = layers.Activation('relu')(x)
        for layer in self.residual_layers:
            x = layer(x)
        self.policy(x)
        self.value(x)
        return x


# Input
x = tf.keras.Input(shape=(8, 8, 26))
y = AlphaConvolutionLayer(256, (3, 3))(x)

# Hidden layers
for i in range(16):
    y = AlphaResidualLayer()(y)

# Value head
y1 = AlphaValueHead()(y+0)

# Policy head
y2 = AlphaPolicyHead()(y)

model = keras.Model(inputs=x, outputs=(y1, y2), name='bot_zoo-16')

model.compile(
    optimizer=tf.keras.optimizers.SGD(
        learning_rate=0.04, nesterov=True, name='nesterov'),
    loss=('mse', 'categorical_crossentropy'))

model.summary()

# N = 500
# bs = 10

# model.fit(x=tf.random.uniform((N, 8, 8, 26), dtype=tf.float16),
#           y=(tf.random.uniform((N, 1,), dtype=tf.float16),
#               tf.random.uniform((N, 225,), dtype=tf.float16)),
#           batch_size=bs,
#           epochs=10,
#           steps_per_epoch=N/bs)
