# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import notifier_um_pb2 as notifier__um__pb2


class NotifierUmStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.RequestEmail = channel.unary_unary(
                '/notifier_um.NotifierUm/RequestEmail',
                request_serializer=notifier__um__pb2.Request.SerializeToString,
                response_deserializer=notifier__um__pb2.Reply.FromString,
                )


class NotifierUmServicer(object):
    """Missing associated documentation comment in .proto file."""

    def RequestEmail(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_NotifierUmServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'RequestEmail': grpc.unary_unary_rpc_method_handler(
                    servicer.RequestEmail,
                    request_deserializer=notifier__um__pb2.Request.FromString,
                    response_serializer=notifier__um__pb2.Reply.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'notifier_um.NotifierUm', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class NotifierUm(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def RequestEmail(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/notifier_um.NotifierUm/RequestEmail',
            notifier__um__pb2.Request.SerializeToString,
            notifier__um__pb2.Reply.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
