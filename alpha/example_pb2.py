# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: example.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='example.proto',
  package='zoo',
  syntax='proto3',
  serialized_options=_b('Z\037github.com/ajzaff/bot_zoo/proto'),
  serialized_pb=_b('\n\rexample.proto\x12\x03zoo\"\x8c\x02\n\x07\x45xample\x12*\n\x07\x62itsets\x18\x01 \x03(\x0b\x32\x19.zoo.Example.BitsetsEntry\x12(\n\x06policy\x18\x02 \x03(\x0b\x32\x18.zoo.Example.PolicyEntry\x12\r\n\x05value\x18\x03 \x01(\x02\x1a(\n\x06\x42itset\x12\x10\n\x08\x61ll_ones\x18\x02 \x01(\x08\x12\x0c\n\x04ones\x18\x03 \x03(\r\x1a\x43\n\x0c\x42itsetsEntry\x12\x0b\n\x03key\x18\x01 \x01(\r\x12\"\n\x05value\x18\x02 \x01(\x0b\x32\x13.zoo.Example.Bitset:\x02\x38\x01\x1a-\n\x0bPolicyEntry\x12\x0b\n\x03key\x18\x01 \x01(\r\x12\r\n\x05value\x18\x02 \x01(\x02:\x02\x38\x01\"*\n\x08\x45xamples\x12\x1e\n\x08\x65xamples\x18\x01 \x03(\x0b\x32\x0c.zoo.ExampleB!Z\x1fgithub.com/ajzaff/bot_zoo/protob\x06proto3')
)




_EXAMPLE_BITSET = _descriptor.Descriptor(
  name='Bitset',
  full_name='zoo.Example.Bitset',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='all_ones', full_name='zoo.Example.Bitset.all_ones', index=0,
      number=2, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='ones', full_name='zoo.Example.Bitset.ones', index=1,
      number=3, type=13, cpp_type=3, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=135,
  serialized_end=175,
)

_EXAMPLE_BITSETSENTRY = _descriptor.Descriptor(
  name='BitsetsEntry',
  full_name='zoo.Example.BitsetsEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='zoo.Example.BitsetsEntry.key', index=0,
      number=1, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='zoo.Example.BitsetsEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=_b('8\001'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=177,
  serialized_end=244,
)

_EXAMPLE_POLICYENTRY = _descriptor.Descriptor(
  name='PolicyEntry',
  full_name='zoo.Example.PolicyEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='zoo.Example.PolicyEntry.key', index=0,
      number=1, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='zoo.Example.PolicyEntry.value', index=1,
      number=2, type=2, cpp_type=6, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=_b('8\001'),
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=246,
  serialized_end=291,
)

_EXAMPLE = _descriptor.Descriptor(
  name='Example',
  full_name='zoo.Example',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='bitsets', full_name='zoo.Example.bitsets', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='policy', full_name='zoo.Example.policy', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='value', full_name='zoo.Example.value', index=2,
      number=3, type=2, cpp_type=6, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[_EXAMPLE_BITSET, _EXAMPLE_BITSETSENTRY, _EXAMPLE_POLICYENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=23,
  serialized_end=291,
)


_EXAMPLES = _descriptor.Descriptor(
  name='Examples',
  full_name='zoo.Examples',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='examples', full_name='zoo.Examples.examples', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=293,
  serialized_end=335,
)

_EXAMPLE_BITSET.containing_type = _EXAMPLE
_EXAMPLE_BITSETSENTRY.fields_by_name['value'].message_type = _EXAMPLE_BITSET
_EXAMPLE_BITSETSENTRY.containing_type = _EXAMPLE
_EXAMPLE_POLICYENTRY.containing_type = _EXAMPLE
_EXAMPLE.fields_by_name['bitsets'].message_type = _EXAMPLE_BITSETSENTRY
_EXAMPLE.fields_by_name['policy'].message_type = _EXAMPLE_POLICYENTRY
_EXAMPLES.fields_by_name['examples'].message_type = _EXAMPLE
DESCRIPTOR.message_types_by_name['Example'] = _EXAMPLE
DESCRIPTOR.message_types_by_name['Examples'] = _EXAMPLES
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Example = _reflection.GeneratedProtocolMessageType('Example', (_message.Message,), dict(

  Bitset = _reflection.GeneratedProtocolMessageType('Bitset', (_message.Message,), dict(
    DESCRIPTOR = _EXAMPLE_BITSET,
    __module__ = 'example_pb2'
    # @@protoc_insertion_point(class_scope:zoo.Example.Bitset)
    ))
  ,

  BitsetsEntry = _reflection.GeneratedProtocolMessageType('BitsetsEntry', (_message.Message,), dict(
    DESCRIPTOR = _EXAMPLE_BITSETSENTRY,
    __module__ = 'example_pb2'
    # @@protoc_insertion_point(class_scope:zoo.Example.BitsetsEntry)
    ))
  ,

  PolicyEntry = _reflection.GeneratedProtocolMessageType('PolicyEntry', (_message.Message,), dict(
    DESCRIPTOR = _EXAMPLE_POLICYENTRY,
    __module__ = 'example_pb2'
    # @@protoc_insertion_point(class_scope:zoo.Example.PolicyEntry)
    ))
  ,
  DESCRIPTOR = _EXAMPLE,
  __module__ = 'example_pb2'
  # @@protoc_insertion_point(class_scope:zoo.Example)
  ))
_sym_db.RegisterMessage(Example)
_sym_db.RegisterMessage(Example.Bitset)
_sym_db.RegisterMessage(Example.BitsetsEntry)
_sym_db.RegisterMessage(Example.PolicyEntry)

Examples = _reflection.GeneratedProtocolMessageType('Examples', (_message.Message,), dict(
  DESCRIPTOR = _EXAMPLES,
  __module__ = 'example_pb2'
  # @@protoc_insertion_point(class_scope:zoo.Examples)
  ))
_sym_db.RegisterMessage(Examples)


DESCRIPTOR._options = None
_EXAMPLE_BITSETSENTRY._options = None
_EXAMPLE_POLICYENTRY._options = None
# @@protoc_insertion_point(module_scope)