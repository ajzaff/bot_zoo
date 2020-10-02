import tensorflow as tf

from tensorflow import keras
from tensorflow.keras import layers


class AlphaConvolutionLayer(object):

    def __init__(self, filters, kernel_size, padding='valid', activation=None, name_prefix=None):
        self.conv = layers.Conv2D(
            filters, kernel_size, padding=padding, data_format='channels_last', name='{prefix}conv'.format(prefix=name_prefix))
        self.norm = layers.BatchNormalization(
            axis=3, name='{prefix}batch_normalization'.format(prefix=name_prefix))
        self.activation = None
        if activation is not None:
            self.activation = layers.Activation(
                activation, name='{prefix}activation'.format(prefix=name_prefix))

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.norm(x)
        if self.activation is None:
            return x
        return self.activation(x)


class AlphaResidualLayer(object):

    def __init__(self, depth=None):
        self.conv1 = AlphaConvolutionLayer(
            256, (3, 3), padding='same', activation='relu', name_prefix='layer_{depth}_0_'.format(depth=depth))
        self.conv2 = AlphaConvolutionLayer(
            256, (3, 3), padding='same', name_prefix='layer_{depth}_1_'.format(depth=depth))
        self.add = layers.Add(
            name='layer_{depth}_shortcut'.format(depth=depth))
        self.activation = layers.Activation(
            'relu', name='layer_{depth}_output'.format(depth=depth))

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv1(inputs)
        x = self.conv2(x)
        x = self.add([x, inputs])
        return self.activation(x)


class AlphaValueHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(
            1, (1, 1), activation='relu', name_prefix='value_conv_')
        self.flatten = layers.Flatten(
            data_format='channels_last', name='value_flatten')
        self.dense = layers.Dense(256, activation='relu', name='value_dense')
        self.output = layers.Dense(1, activation='tanh', name='value')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.flatten(x)
        x = self.dense(x)
        return self.output(x)


class AlphaPolicyHead(object):

    def __init__(self):
        self.conv = AlphaConvolutionLayer(
            2, (1, 1), activation='relu', name_prefix='policy_conv_')
        self.flatten = layers.Flatten(
            data_format='channels_last', name='policy_flatten')
        self.dense = layers.Dense(256, activation='relu', name='policy_dense')
        self.output = layers.Dense(230, activation='linear', name='policy')

    def __call__(self, inputs):
        return self.call(inputs)

    def call(self, inputs):
        x = self.conv(inputs)
        x = self.flatten(x)
        x = self.dense(x)
        return self.output(x)


model_depth = 16

N = 1000
N_validation = 100
bs = 100
epochs = 10
steps_per_epoch = N/bs

# Input
x = tf.keras.Input(shape=(8, 8, 21), name='input')
y = AlphaConvolutionLayer(
    256, (3, 3), activation='relu', name_prefix='input_layer_')(x)

# Hidden layers
for i in range(model_depth):
    y = AlphaResidualLayer(depth=1+i)(y)

# Value head
y1 = AlphaValueHead()(y)

# Policy head
y2 = AlphaPolicyHead()(y)

model = keras.Model(inputs=x, outputs=(y1, y2),
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
_y2 = tf.nn.softmax(tf.random.uniform(
    (N+N_validation, 230,),  dtype=tf.float16))

model.fit(x=_x[:N], y=(_y1[:N], _y2[:N]),
          batch_size=bs,
          epochs=epochs,
          steps_per_epoch=steps_per_epoch,
          validation_steps=3,
          validation_data=(_x[N:], (_y1[N:], _y2[N:])),
          callbacks=[model_checkpoint_callback])

# tf.saved_model.save(model,  saved_model_filepath)
