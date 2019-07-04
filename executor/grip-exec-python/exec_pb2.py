# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: exec.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='exec.proto',
  package='executor',
  syntax='proto3',
  serialized_options=None,
  serialized_pb=_b('\n\nexec.proto\x12\x08\x65xecutor\x1a\x1cgoogle/protobuf/struct.proto\"&\n\x04\x43ode\x12\x0c\n\x04\x63ode\x18\x01 \x01(\t\x12\x10\n\x08\x66unction\x18\x02 \x01(\t\"*\n\rCompileResult\x12\n\n\x02id\x18\x01 \x01(\r\x12\r\n\x05\x65rror\x18\x02 \x01(\t\";\n\x05Input\x12$\n\x04\x64\x61ta\x18\x01 \x03(\x0b\x32\x16.google.protobuf.Value\x12\x0c\n\x04\x63ode\x18\x02 \x01(\r\"=\n\x06Result\x12$\n\x04\x64\x61ta\x18\x01 \x01(\x0b\x32\x16.google.protobuf.Value\x12\r\n\x05\x65rror\x18\x02 \x01(\t2t\n\x08\x45xecutor\x12\x34\n\x07\x43ompile\x12\x0e.executor.Code\x1a\x17.executor.CompileResult\"\x00\x12\x32\n\x07Process\x12\x0f.executor.Input\x1a\x10.executor.Result\"\x00(\x01\x30\x01\x62\x06proto3')
  ,
  dependencies=[google_dot_protobuf_dot_struct__pb2.DESCRIPTOR,])




_CODE = _descriptor.Descriptor(
  name='Code',
  full_name='executor.Code',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='code', full_name='executor.Code.code', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='function', full_name='executor.Code.function', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
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
  serialized_start=54,
  serialized_end=92,
)


_COMPILERESULT = _descriptor.Descriptor(
  name='CompileResult',
  full_name='executor.CompileResult',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='executor.CompileResult.id', index=0,
      number=1, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='error', full_name='executor.CompileResult.error', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
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
  serialized_start=94,
  serialized_end=136,
)


_INPUT = _descriptor.Descriptor(
  name='Input',
  full_name='executor.Input',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='data', full_name='executor.Input.data', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='code', full_name='executor.Input.code', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
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
  serialized_start=138,
  serialized_end=197,
)


_RESULT = _descriptor.Descriptor(
  name='Result',
  full_name='executor.Result',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='data', full_name='executor.Result.data', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='error', full_name='executor.Result.error', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
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
  serialized_start=199,
  serialized_end=260,
)

_INPUT.fields_by_name['data'].message_type = google_dot_protobuf_dot_struct__pb2._VALUE
_RESULT.fields_by_name['data'].message_type = google_dot_protobuf_dot_struct__pb2._VALUE
DESCRIPTOR.message_types_by_name['Code'] = _CODE
DESCRIPTOR.message_types_by_name['CompileResult'] = _COMPILERESULT
DESCRIPTOR.message_types_by_name['Input'] = _INPUT
DESCRIPTOR.message_types_by_name['Result'] = _RESULT
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Code = _reflection.GeneratedProtocolMessageType('Code', (_message.Message,), dict(
  DESCRIPTOR = _CODE,
  __module__ = 'exec_pb2'
  # @@protoc_insertion_point(class_scope:executor.Code)
  ))
_sym_db.RegisterMessage(Code)

CompileResult = _reflection.GeneratedProtocolMessageType('CompileResult', (_message.Message,), dict(
  DESCRIPTOR = _COMPILERESULT,
  __module__ = 'exec_pb2'
  # @@protoc_insertion_point(class_scope:executor.CompileResult)
  ))
_sym_db.RegisterMessage(CompileResult)

Input = _reflection.GeneratedProtocolMessageType('Input', (_message.Message,), dict(
  DESCRIPTOR = _INPUT,
  __module__ = 'exec_pb2'
  # @@protoc_insertion_point(class_scope:executor.Input)
  ))
_sym_db.RegisterMessage(Input)

Result = _reflection.GeneratedProtocolMessageType('Result', (_message.Message,), dict(
  DESCRIPTOR = _RESULT,
  __module__ = 'exec_pb2'
  # @@protoc_insertion_point(class_scope:executor.Result)
  ))
_sym_db.RegisterMessage(Result)



_EXECUTOR = _descriptor.ServiceDescriptor(
  name='Executor',
  full_name='executor.Executor',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  serialized_start=262,
  serialized_end=378,
  methods=[
  _descriptor.MethodDescriptor(
    name='Compile',
    full_name='executor.Executor.Compile',
    index=0,
    containing_service=None,
    input_type=_CODE,
    output_type=_COMPILERESULT,
    serialized_options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Process',
    full_name='executor.Executor.Process',
    index=1,
    containing_service=None,
    input_type=_INPUT,
    output_type=_RESULT,
    serialized_options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_EXECUTOR)

DESCRIPTOR.services_by_name['Executor'] = _EXECUTOR

# @@protoc_insertion_point(module_scope)