#   Copyright (c) 2022-present, Adil Alper DALKIRAN
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#   ==============================================================================

import asyncio
from os import environ
import string
import sys
import traceback
from typing import Tuple

# If DEVELOPMENT_SEARCH_PATH environment variable specified, search for py-inventa module 
# in the specified path. This required for development of py-inventa itself.
development_search_path = environ.get("DEVELOPMENT_SEARCH_PATH")
if development_search_path and len(development_search_path) > 0:
    sys.path.insert(0, f"{development_search_path}/py-inventa")

environ.update([("PYTHONASYNCIODEBUG", "1")])

from inventa import Inventa, InventaRole, ServiceDescriptor, RPCCallRequest
import numpy as np

hostname = environ.get("HOSTNAME")
SelfDescriptor = ServiceDescriptor.ParseServiceFullId(f"svc:linalg:{hostname}")
OrchestratorDescriptor = ServiceDescriptor.ParseServiceFullId("svc:orc:")

def readNumpyArrayBigEndianWithReshape(shapeStr: bytes, byteArray: bytes) -> np.ndarray:
    # IMPORTANT: ">i4" means int32 (4 bytes) big endian byte order. It's important to be in consensus with other parties,
    # so Golang side can parse and deserialize the data correctly.
    # See: https://numpy.org/doc/stable/reference/arrays.dtypes.html
    matrixShape = list(map(int, shapeStr.split(b',')))
    matrixRawArray = np.frombuffer(byteArray, dtype=">i4")
    # matrixRawArray is one-dimensional array. We should reshape it
    matrix = np.reshape(matrixRawArray, matrixShape)
    return matrix

def writeNumpyArrayBigEndianWitShape(matrix: np.ndarray) -> Tuple[bytes, str]:
    # IMPORTANT: ">i4" means int32 (4 bytes) big endian byte order. It's important to be in consensus with other parties,
    # so Golang side can parse and deserialize the data correctly.
    # See: https://numpy.org/doc/stable/reference/arrays.dtypes.html
    resultMatrix = matrix.astype(dtype=">i4")
    resultShape = ",".join(map(str, list(resultMatrix.shape)))
    encodedResultMatrix = resultMatrix.tobytes()
    return encodedResultMatrix, resultShape


def rpcCommandLinalgMatmul(req: RPCCallRequest) -> string:
    if len(req.Args) != 4:
        e = Exception(f"Argument count must be 4 but for linalg-matmul {len(req.Args)} found. Args: {req.Args}")
        print("Error:", e)
        return req.ErrorResponse(e)

    try:
        matrixA = readNumpyArrayBigEndianWithReshape(req.Args[0], req.Args[1])
        matrixB = readNumpyArrayBigEndianWithReshape(req.Args[2], req.Args[3])
        resultMatrix = np.matmul(matrixA, matrixB)
        encodedResultMatrix, resultShape = writeNumpyArrayBigEndianWitShape(resultMatrix)
        return [resultShape, encodedResultMatrix]
    except Exception as e:
        traceback_str = ''.join(traceback.format_tb(e.__traceback__))
        print("Error:", e, "\nwith stack trace:\n", traceback_str, "\nargs: ", req.Args)
        return req.ErrorResponse(e)


RPCCommandFnRegistry = {
    "linalg-matmul":  rpcCommandLinalgMatmul,
}

def connect_to_redis() -> Inventa:
    hostname = environ.get("REDIS_HOST", "localhost")
    port = environ.get("REDIS_PORT", 6379)
    password = environ.get("REDIS_PASSWORD", None)

    r = Inventa(hostname, port, password, SelfDescriptor.ServiceType, SelfDescriptor.ServiceId, InventaRole.Service, RPCCommandFnRegistry)
    r.Start()
    return r

async def try_register_to_orchestrator(inventa: Inventa):
    try:
        await inventa.TryRegisterToOrchestrator(OrchestratorDescriptor.Encode(), 30, 3000)
    except Exception as e:
        print(f"Registration to orchestration service was failed! Breaking down! {e}")
        raise e
    print(f"Registered to orchestration service as {SelfDescriptor.Encode()}")

def exception_handler(loop, context):
    if context["future"].get_name() == "try_register_to_orchestrator":
        loop.stop()
    else:
        traceback_str = ''.join(traceback.format_tb(context["exception"].__traceback__))
        print("Error:", context["exception"], "\nwith stack trace:\n", traceback_str)

def main():
    print("Welcome to Calculator Linear Algebra Service in Python!")
    print("=================================")
    print("This module acts as linear algebra service server.\n\n\n")

    event_loop = asyncio.get_event_loop()
    event_loop.set_exception_handler(exception_handler)
    inventa = connect_to_redis()
    event_loop.create_task(try_register_to_orchestrator(inventa), name="try_register_to_orchestrator")
    
    event_loop.run_forever()

if __name__ == '__main__':
    main()
