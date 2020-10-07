import tensorflow as tf

from tensorflow.keras import layers


class AlphaConvolutionLayer(layers.Layer):

    def __init__(self, filters, kernel_size, padding='valid', activation=None, name=None):
        super(AlphaConvolutionLayer, self).__init__(name=name)
        self.conv = layers.Conv2D(
            filters, kernel_size, padding=padding, data_format='channels_last')
        self.norm = layers.BatchNormalization(axis=3)
        self.activation = None
        if activation is not None:
            self.activation = layers.Activation(activation)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.norm(x)
        if self.activation is None:
            return x
        return self.activation(x)


class AlphaResidualLayer(layers.Layer):

    def __init__(self):
        super(AlphaResidualLayer, self).__init__()
        self.conv1 = AlphaConvolutionLayer(
            256, (3, 3), padding='same', activation='relu')
        self.conv2 = AlphaConvolutionLayer(256, (3, 3), padding='same')
        self.add = layers.Add()
        self.activation = layers.Activation('relu')

    def call(self, inputs):
        x = self.conv1(inputs)
        x = self.conv2(x)
        x = self.add([x, inputs])
        return self.activation(x)


class AlphaValueHead(layers.Layer):

    def __init__(self, name=None):
        super(AlphaValueHead, self).__init__(name=name)
        self.conv = AlphaConvolutionLayer(1, (1, 1), activation='relu')
        self.flatten = layers.Flatten(data_format='channels_last')
        self.hidden = layers.Dense(256, activation='relu')
        self.value_dense = layers.Dense(1)
        self.value_output = layers.Activation('tanh')

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.flatten(x)
        x = self.hidden(x)
        x = self.value_dense(x)
        return self.value_output(x)


class AlphaPolicyHead(layers.Layer):

    def __init__(self, name=None):
        super(AlphaPolicyHead, self).__init__(name=name)
        self.conv = AlphaConvolutionLayer(2, (1, 1), activation='relu')
        self.flatten = layers.Flatten(data_format='channels_last')
        self.dense = layers.Dense(256, activation='relu')
        self.policy_output = layers.Dense(232, activation='linear')

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.flatten(x)
        x = self.dense(x)
        return self.policy_output(x)


model_depth = 16

N = 100000
N_validation = 10000
bs = 200
epochs = 64
steps_per_epoch = N/bs

# Input
x = tf.keras.Input(shape=(8, 8, 21), name='x')
y = AlphaConvolutionLayer(256, (3, 3), activation='relu', name='input_')(x)

# Hidden layers
for i in range(model_depth):
    y = AlphaResidualLayer()(y)

# Value head
y1 = AlphaValueHead(name='value_')(y)

# Policy head
y2 = AlphaPolicyHead(name='policy_')(y)

model = tf.keras.Model(inputs=x, outputs=(y1, y2),
                       name='bot_alpha_zoo-{depth}'.format(depth=model_depth))

model.compile(
    optimizer=tf.keras.optimizers.SGD(
        learning_rate=0.01, momentum=0.9),
    loss=('mse', tf.keras.losses.CategoricalCrossentropy(from_logits=True)))

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

_x = tf.random.categorical(tf.math.log([[0.9, 0.1]]), (N+N_validation)*8*8*21)
_x = tf.reshape(_x, (N+N_validation, 8, 8, 21))

_y1 = 2*tf.random.uniform((N+N_validation, 1,), dtype=tf.float16)-1
_y2 = tf.nn.softmax(tf.random.normal(
    (N+N_validation, 232,),  dtype=tf.float16))

# model.fit(x=_x[:N], y=(_y1[:N], _y2[:N]),
#           batch_size=bs,
#           epochs=epochs,
#           steps_per_epoch=steps_per_epoch,
#           validation_steps=3,
#           validation_data=(_x[N:], (_y1[N:], _y2[N:])),
#           callbacks=[model_checkpoint_callback])

tf.saved_model.save(model, saved_model_filepath)
