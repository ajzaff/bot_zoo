import tensorflow as tf

from tensorflow import keras
from tensorflow.keras import layers


class AlphaConvolutionLayer(object):

    def __init__(self, filters, kernel_size, padding='valid', activation='relu'):
        self.conv = layers.Conv2D(
            filters, kernel_size, padding=padding, data_format='channels_last')
        self.norm = layers.BatchNormalization(axis=3)
        self.activation = layers.Activation(activation)

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.norm(x)
        return self.activation(x)


class AlphaResidualLayer(object):

    def __init__(self):
        self.conv1 = AlphaConvolutionLayer(
            256, (3, 3), padding='same', activation='relu')
        self.conv2 = AlphaConvolutionLayer(256, (3, 3), padding='same')
        self.add = layers.Add()

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv1(inputs)
        x = self.conv2(x)
        x = self.add([x, inputs])
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
        x = self.flatten(x)
        x = self.dense(x)
        return self.output(x)


model_depth = 16

N = 5000
N_validation = 100
bs = 10
epochs = 20
steps_per_epoch = N/bs

# Input
x = tf.keras.Input(shape=(8, 8, 26))
y = AlphaConvolutionLayer(256, (3, 3), activation='relu')(x)

# Hidden layers
for i in range(model_depth):
    y = AlphaResidualLayer()(y)

# Value head
y1 = AlphaValueHead()(y)

# Policy head
y2 = AlphaPolicyHead()(y)

model = keras.Model(inputs=x, outputs=(y1, y2),
                    name='bot_alpha_zoo-{depth}'.format(depth=model_depth))

model.compile(
    optimizer=tf.keras.optimizers.SGD(
        learning_rate=0.04, nesterov=True, name='nesterov'),
    loss=('mse', 'categorical_crossentropy'))

model.summary()


checkpoint_filepath = './data/checkpoint/{name}'.format(name=model.name)
saved_model_filepath = './data/saved_models/{name}'.format(name=model.name)

model_checkpoint_callback = tf.keras.callbacks.ModelCheckpoint(
    filepath=checkpoint_filepath,
    save_weights_only=True,
    monitor='val_loss',
    mode='auto',
    save_best_only=True,
    load_weights_on_restart=True,
)

model.fit(x=tf.random.uniform((N, 8, 8, 26), dtype=tf.float16),
          y=(tf.random.uniform((N, 1,), dtype=tf.float16),
              tf.random.uniform((N, 225,), dtype=tf.float16)),
          batch_size=bs,
          epochs=epochs,
          steps_per_epoch=steps_per_epoch,
          validation_steps=1,
          validation_data=(tf.random.uniform((N_validation, 8, 8, 26), dtype=tf.float16),
                           (tf.random.uniform((N_validation, 1,), dtype=tf.float16),
                            tf.random.uniform((N_validation, 225,), dtype=tf.float16))),
          callbacks=[model_checkpoint_callback])

# tf.saved_model.save(model,  saved_model_filepath)
