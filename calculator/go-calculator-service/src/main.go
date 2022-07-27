package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	logging "github.com/adalkiran/go-colorful-logging"
	"github.com/adalkiran/go-inventa"
)

var (
	SelfDescriptor         inventa.ServiceDescriptor
	OrchestratorDescriptor inventa.ServiceDescriptor

	RPCCommandFnRegistry map[string]inventa.RPCCommandFn

	inventaObj *inventa.Inventa
)

func main() {
	var err error
	hostname := os.Getenv("HOSTNAME")
	// ServiceType: "calc"
	// ServiceId:   "{HOSTNAME}"
	SelfDescriptor, err = inventa.ParseServiceFullId(fmt.Sprintf("svc:calc:%s", hostname))
	if err != nil {
		panic(err)
	}
	OrchestratorDescriptor, err = inventa.ParseServiceFullId("svc:orc:")
	if err != nil {
		panic(err)
	}

	//See: https://codewithyury.com/golang-wait-for-all-goroutines-to-finish/
	//See: https://www.geeksforgeeks.org/using-waitgroup-in-golang/
	waitGroup := new(sync.WaitGroup)

	logging.Freef("", "Welcome to Calculator Calculate Service in Go!")
	logging.Freef("", "=================================")
	logging.Freef("", "This module acts as calculator service server.")
	logging.LineSpacer(3)

	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Invalid REDIS_PORT value: \"%s\"", os.Getenv("REDIS_PORT"))
		return
	}

	RPCCommandFnRegistry = map[string]inventa.RPCCommandFn{
		"calculate-sum":      rpcCommandCalculateSum,
		"calculate-subtract": rpcCommandCalculateSubtract,
	}

	waitGroup.Add(1)
	inventaObj = inventa.NewInventa(os.Getenv("REDIS_HOST"), redisPort, os.Getenv("REDIS_PASSWORD"), SelfDescriptor.ServiceType, SelfDescriptor.ServiceId, inventa.InventaRoleService, RPCCommandFnRegistry)

	_, err = inventaObj.Start()
	if err != nil {
		panic(err)
	}

	err = inventaObj.TryRegisterToOrchestrator(OrchestratorDescriptor.Encode(), 30, 3*time.Second)
	if err == nil {
		logging.Infof(logging.ProtoAPP, "Registered to orchestration service as <u>%s</u>", SelfDescriptor.Encode())
	} else {
		logging.Errorf(logging.ProtoAPP, "Registration to orchestration service was failed! Breaking down! %s", err)
		return
	}

	waitGroup.Wait()

}

func rpcCommandCalculateSum(req *inventa.RPCCallRequest) []string {
	number1, err := strconv.Atoi(req.Args[0])
	if err != nil {
		return req.ErrorResponse(err)
	}
	number2, err := strconv.Atoi(req.Args[1])
	if err != nil {
		return req.ErrorResponse(err)
	}

	result := number1 + number2

	//Sending result value with this service's programming language name, just to show at orchestrator log.
	return []string{"go", string(strconv.Itoa(result))}
}

func rpcCommandCalculateSubtract(req *inventa.RPCCallRequest) []string {
	number1, err := strconv.Atoi(req.Args[0])
	if err != nil {
		return req.ErrorResponse(err)
	}
	number2, err := strconv.Atoi(req.Args[1])
	if err != nil {
		return req.ErrorResponse(err)
	}

	result := number1 - number2

	//Sending result value with this service's programming language name, just to show at orchestrator log.
	return []string{"go", string(strconv.Itoa(result))}
}
