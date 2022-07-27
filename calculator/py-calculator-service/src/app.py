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
import traceback
from warnings import catch_warnings

import sys

# If DEVELOPMENT_SEARCH_PATH environment variable specified, search for py-inventa module 
# in the specified path. This required for development of py-inventa itself.
development_search_path = environ.get("DEVELOPMENT_SEARCH_PATH")
if development_search_path and len(development_search_path) > 0:
    sys.path.insert(0, f"{development_search_path}/py-inventa")

environ.update([("PYTHONASYNCIODEBUG", "1")])

from inventa import Inventa, InventaRole, ServiceDescriptor, RPCCallRequest

hostname = environ.get("HOSTNAME")
SelfDescriptor = ServiceDescriptor.ParseServiceFullId(f"svc:calc:{hostname}")
OrchestratorDescriptor = ServiceDescriptor.ParseServiceFullId("svc:orc:")


def rpcCommandCalculateSum(req: RPCCallRequest) -> list[bytes]:
    number1 = None
    number2 = None
    try:
        number1 = int(req.Args[0])
        number2 = int(req.Args[1])
    except Exception as e:
        return req.ErrorResponse(e)
    
    result = number1 + number2
    
    #Sending result value with this service's programming language name, just to show at orchestrator log.
    return ["python", str(result)]

def rpcCommandCalculateSubstract(req: RPCCallRequest) -> list[bytes]:
    number1 = None
    number2 = None
    try:
        number1 = int(req.Args[0])
        number2 = int(req.Args[1])
    except Exception as e:
        return req.ErrorResponse(e)
    
    result = number1 - number2
    
    #Sending result value with this service's programming language name, just to show at orchestrator log.
    return ["python", str(result)]


RPCCommandFnRegistry = {
    "calculate-sum":  rpcCommandCalculateSum,
    "calculate-substract": rpcCommandCalculateSubstract,
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
    print("Welcome to Calculator Calculate Service in Python!")
    print("=================================")
    print("This module acts as calculator service server.\n\n\n")

    event_loop = asyncio.get_event_loop()
    event_loop.set_exception_handler(exception_handler)
    inventa = connect_to_redis()
    event_loop.create_task(try_register_to_orchestrator(inventa), name="try_register_to_orchestrator")
    
    event_loop.run_forever()

if __name__ == '__main__':
    main()
