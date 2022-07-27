/*
   Copyright (c) 2022-present, Adil Alper DALKIRAN

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	logging "github.com/adalkiran/go-colorful-logging"
	"github.com/adalkiran/go-inventa"
)

type CalculatorModule struct {
	inventa.ServiceConsumer
}

type LinalgModule struct {
	inventa.ServiceConsumer
}

var (
	SelfDescriptor       inventa.ServiceDescriptor
	RPCCommandFnRegistry map[string]inventa.RPCCommandFn

	CalculatorModules map[string]*CalculatorModule
	LinalgModules     map[string]*LinalgModule

	inventaObj *inventa.Inventa
	mu         sync.Mutex
)

func NewCalculatorModule(id string, inventaObj *inventa.Inventa) *CalculatorModule {
	return &CalculatorModule{
		ServiceConsumer: inventa.ServiceConsumer{
			SelfDescriptor: inventa.ServiceDescriptor{
				ServiceType: "calc",
				ServiceId:   id,
			},
			Inventa: inventaObj,
		},
	}
}

func NewLinalgModule(id string, inventaObj *inventa.Inventa) *LinalgModule {
	return &LinalgModule{
		ServiceConsumer: inventa.ServiceConsumer{
			SelfDescriptor: inventa.ServiceDescriptor{
				ServiceType: "linalg",
				ServiceId:   id,
			},
			Inventa: inventaObj,
		},
	}
}

func main() {
	var err error
	// ServiceType: "orc"
	// ServiceId:   "", empty for orchestrator, because there is ony one orchestrator
	SelfDescriptor, err = inventa.ParseServiceFullId("svc:orc:")
	if err != nil {
		panic(err)
	}

	//See: https://codewithyury.com/golang-wait-for-all-goroutines-to-finish/
	//See: https://www.geeksforgeeks.org/using-waitgroup-in-golang/
	waitGroup := new(sync.WaitGroup)

	logging.Freef("", "Welcome to Calculator Orchestrator Server!")
	logging.Freef("", "=================================")
	logging.Freef("", "This module acts as service registrar server and orchestrates other modules.")
	logging.Freef("", "In this demo, this project makes some random calculations in every 2 seconds.")
	logging.LineSpacer(3)

	redisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Invalid REDIS_PORT value: \"%s\"", os.Getenv("REDIS_PORT"))
		return
	}

	// Currently we shouldn't provide any RPC function at orchestrator side.
	RPCCommandFnRegistry = map[string]inventa.RPCCommandFn{}

	CalculatorModules = map[string]*CalculatorModule{}
	LinalgModules = map[string]*LinalgModule{}

	waitGroup.Add(1)
	inventaObj = inventa.NewInventa(os.Getenv("REDIS_HOST"), redisPort, os.Getenv("REDIS_PASSWORD"), SelfDescriptor.ServiceType, SelfDescriptor.ServiceId, inventa.InventaRoleOrchestrator, RPCCommandFnRegistry)

	inventaObj.OnServiceRegistering = serviceRegisteringHandler
	inventaObj.OnServiceUnregistering = serviceUnregisteringHandler

	_, err = inventaObj.Start()
	if err != nil {
		panic(err)
	}

	waitGroup.Add(1)
	go doRemoteCalculations(waitGroup)
	waitGroup.Wait()
}

func doRemoteCalculations(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	calcTicker := time.NewTicker(2 * time.Second)
	linalgTicker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-calcTicker.C:
			doOneRemoteCalculation()
		case <-linalgTicker.C:
			doOneRemoteLinalgConstantValid()
			doOneRemoteLinalgConstantInvalid()
			doOneRemoteLinalgRandom()
		}
	}
}

func doOneRemoteCalculation() {
	number1 := rand.Intn(1000)
	number2 := rand.Intn(1000)
	calculatorModule := selectOneCalculatorService()
	logging.Freef("", "")
	if calculatorModule == nil {
		logging.Errorf(logging.ProtoAPP, "There isn't any available and registered calculator service found")
		return
	}
	logging.Infof(logging.ProtoAPP, "Doing remote calculations with %d and %d to service: %s...", number1, number2, calculatorModule.SelfDescriptor.Encode())
	args := []string{strconv.Itoa(number1), strconv.Itoa(number2)}

	// Doing Summation

	sumResponse, err := calculatorModule.Inventa.CallSync(calculatorModule.SelfDescriptor.Encode(), "calculate-sum", args, 3*time.Second)
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote calculation failed (%s): %s", "calculate-sum", err)
		return
	}
	remoteServiceLanguage := sumResponse[0]

	sumResponseInt, err := strconv.Atoi(sumResponse[1])
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote response (%s) couldn't be converted to int: %s", sumResponse, err)
		return
	}

	logging.Infof(logging.ProtoAPP, "Remote calculation of %s(%d, %d) = %d  (service language: %s)", "calculate-sum", number1, number2, sumResponseInt, remoteServiceLanguage)

	// Doing Substraction

	substractResponse, err := calculatorModule.Inventa.CallSync(calculatorModule.SelfDescriptor.Encode(), "calculate-substract", args, 3*time.Second)
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote calculation failed (%s): %s", "calculate-substract", err)
		return
	}
	remoteServiceLanguage = substractResponse[0]

	substractResponseInt, err := strconv.Atoi(substractResponse[1])
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote response (%s) couldn't be converted to int: %s", substractResponse, err)
		return
	}
	logging.Infof(logging.ProtoAPP, "Remote calculation of %s(%d, %d) = %d  (service language: %s)", "calculate-substract", number1, number2, substractResponseInt, remoteServiceLanguage)
}

func selectOneCalculatorService() *CalculatorModule {
	if len(CalculatorModules) == 0 {
		return nil
	}
	idx := rand.Intn(len(CalculatorModules))
	if idx < 0 {
		return nil
	}
	//TODO: Should find other nicer way to select one of map items randomly. I didn't prefer use reflection here.
	i := 0
	for _, item := range CalculatorModules {
		if i == idx {
			return item
		}
		i++
	}
	return nil
}

func getMatrixShape(matrix [][]int32) []int {
	rowCount := len(matrix)
	colCount := 0
	for _, row := range matrix {
		colCount = int(math.Max(float64(colCount), float64(len(row))))
	}
	return []int{rowCount, colCount}
}

func encodeMatrixToBytesWithShape(matrix [][]int32) (string, string) {
	//IMPORTANT: Python side uses ">i4" as data type, it means int32 (4 bytes) big endian byte order.
	//It's important to be in consensus with other parties, so Python side can parse and deserialize the data correctly.

	//Determine matrix shape
	matrixShape := getMatrixShape(matrix)
	//See: https://pkg.go.dev/encoding/binary#Write
	buf := new(bytes.Buffer)
	for _, row := range matrix {
		for _, cell := range row {
			err := binary.Write(buf, binary.BigEndian, cell)
			if err != nil {
				logging.Errorf(logging.ProtoAPP, "Error occured while converting matrix to byte array: %s", err)
				return "0,0", ""
			}
		}
	}
	return strconv.Itoa(matrixShape[0]) + "," + strconv.Itoa(matrixShape[1]), buf.String()
}

func decodeMatrixFromBytesWithShape(shapeString string, encodedString string) ([][]int32, error) {
	//IMPORTANT: Python side uses ">i4" as data type, it means int32 (4 bytes) big endian byte order.
	//It's important to be in consensus with other parties, so Python side can parse and deserialize the data correctly.

	shapeParts := strings.Split(shapeString, ",")
	shape := make([]int, len(shapeParts))

	for i, s := range shapeParts {
		shape[i], _ = strconv.Atoi(s)
	}

	rowCount := shape[0]
	colCount := shape[1]

	buf := bytes.NewReader([]byte(encodedString))

	var cell int32
	var err error

	matrix := make([][]int32, rowCount)
	for rowIdx := 0; rowIdx < rowCount; rowIdx++ {
		row := make([]int32, colCount)
		for colIdx := 0; colIdx < colCount; colIdx++ {
			err = binary.Read(buf, binary.BigEndian, &cell)
			if err != nil {
				return nil, err
			}
			row[colIdx] = cell
		}
		matrix[rowIdx] = row
	}
	return matrix, nil

}

func generateRandomMatrix(rowCount int, colCount int) [][]int32 {
	//Generate a random colCount x rowCount matrix
	//See: https://stackoverflow.com/a/53575298
	matrix := make([][]int32, rowCount)
	for rowIdx := 0; rowIdx < rowCount; rowIdx++ {
		row := make([]int32, colCount)
		for colIdx := 0; colIdx < colCount; colIdx++ {
			row[colIdx] = rand.Int31n(1000)
		}
		matrix[rowIdx] = row
	}
	return matrix
}

func doOneRemoteLinalgRandom() {
	logging.Freef("", "")
	logging.Infof(logging.ProtoAPP, "Doing a valid matrix multiplication with <u>RANDOM matrices</u>, a <u>VALID response</u> is expected.")
	//Generate a row count value between 1 and 5 for matrixA
	matrixA_rowCount := rand.Intn(4) + 1
	//Generate a column count value between 1 and 5 for matrixA
	matrixA_colCount := rand.Intn(4) + 1

	//For matrix multiplication, column count of matrixA and row count of matrixB must be equal.
	matrixB_rowCount := matrixA_colCount
	//Generate a column count value between 1 and 5 for matrixB
	matrixB_colCount := rand.Intn(4) + 1

	matrixA := generateRandomMatrix(matrixA_rowCount, matrixA_colCount)
	matrixB := generateRandomMatrix(matrixB_rowCount, matrixB_colCount)
	doOneRemoteLinalg(matrixA, matrixB)
}

func doOneRemoteLinalgConstantValid() {
	logging.Freef("", "")
	logging.Infof(logging.ProtoAPP, "Doing a valid matrix multiplication with <u>CONSTANT matrices</u>, a <u>VALID response</u> is expected.")

	matrixA := [][]int32{{1, 2, 3}, {4, 5, 6}}
	matrixB := [][]int32{{2}, {2}, {2}}
	doOneRemoteLinalg(matrixA, matrixB)
}

func doOneRemoteLinalgConstantInvalid() {
	logging.Freef("", "")
	logging.Infof(logging.ProtoAPP, "Doing a valid matrix multiplication with <u>CONSTANT matrices</u>, an <u>ERROR response</u> is expected, because column count of matrix A is not equal with row count of matrix B.")
	matrixA := [][]int32{{1, 2, 3}, {4, 5, 6}}
	matrixB := [][]int32{{2}, {2}, {2}, {2}}
	doOneRemoteLinalg(matrixA, matrixB)
}

func doOneRemoteLinalg(matrixA [][]int32, matrixB [][]int32) {
	matrixAShape := getMatrixShape(matrixA)
	matrixBShape := getMatrixShape(matrixB)

	encodedMatrixAShape, encodedMatrixA := encodeMatrixToBytesWithShape(matrixA)
	encodedMatrixBShape, encodedMatrixB := encodeMatrixToBytesWithShape(matrixB)

	linalgModule := selectOneLinalgService()
	if linalgModule == nil {
		logging.Errorf(logging.ProtoAPP, "There isn't any available and registered linalg service found")
		return
	}
	logging.Infof(logging.ProtoAPP, "Doing remote linalg matrix multiplication calculation with \nmatrixA(%v) %v\nmatrixB(%v) %v\nto service: %s...", matrixAShape, matrixA, matrixBShape, matrixB, linalgModule.SelfDescriptor.Encode())
	args := []string{encodedMatrixAShape, encodedMatrixA, encodedMatrixBShape, encodedMatrixB}

	matmulResponse, err := linalgModule.Inventa.CallSync(linalgModule.SelfDescriptor.Encode(), "linalg-matmul", args, 3*time.Second)
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote linalg calculation failed (%s): %s", "linalg-matmul", err)
		return
	}

	encodedResponseMatrixShape := matmulResponse[0]
	encodedResponseMatrix := matmulResponse[1]
	responseMatrix, err := decodeMatrixFromBytesWithShape(encodedResponseMatrixShape, encodedResponseMatrix)
	responseMatrixShape := getMatrixShape(responseMatrix)
	if err != nil {
		logging.Errorf(logging.ProtoAPP, "Remote linalg calculation parsing failed (%s): %s", "linalg-matmul", err)
		return
	}

	logging.Infof(logging.ProtoAPP, "Remote linalg calculation of %s = shape(%v) %v", "linalg-matmul", responseMatrixShape, responseMatrix)

}

func selectOneLinalgService() *LinalgModule {
	if len(LinalgModules) == 0 {
		return nil
	}
	idx := rand.Intn(len(LinalgModules))
	if idx < 0 {
		return nil
	}
	//TODO: Should find other nicer way to select one of map items randomly. I didn't prefer use reflection here.
	i := 0
	for _, item := range LinalgModules {
		if i == idx {
			return item
		}
		i++
	}
	return nil
}

func serviceRegisteringHandler(serviceDescriptor inventa.ServiceDescriptor) error {
	mu.Lock()
	defer mu.Unlock()
	switch serviceDescriptor.ServiceType {
	case "calc":
		CalculatorModules[serviceDescriptor.ServiceId] = NewCalculatorModule(serviceDescriptor.ServiceId, inventaObj)
		logging.Infof(logging.ProtoAPP, "Calculator module has been registered as %s", serviceDescriptor.Encode())
		return nil
	case "linalg":
		LinalgModules[serviceDescriptor.ServiceId] = NewLinalgModule(serviceDescriptor.ServiceId, inventaObj)
		logging.Infof(logging.ProtoAPP, "Linalg module has been registered as %s", serviceDescriptor.Encode())
		return nil

	default:
		logging.Warningf(logging.ProtoAPP, "Unknown service type to register: %s. Raw args: <u>%s</u>", serviceDescriptor.ServiceType, serviceDescriptor.Encode())
		return fmt.Errorf("unknown service type to register: %s", serviceDescriptor.ServiceType)
	}
}

func serviceUnregisteringHandler(serviceDescriptor inventa.ServiceDescriptor, isZombie bool) error {
	mu.Lock()
	defer mu.Unlock()
	switch serviceDescriptor.ServiceType {
	case "calc":
		mb := CalculatorModules[serviceDescriptor.ServiceId]
		if mb == nil {
			return nil
		}
		delete(CalculatorModules, serviceDescriptor.ServiceId)
		var warningFormat string
		if isZombie {
			warningFormat = "Calculator module %s is not alive anymore, it has been unregistered."
		} else {
			warningFormat = "Calculator module has been unregistered: %s"
		}
		logging.Infof(logging.ProtoAPP, warningFormat, serviceDescriptor.Encode())
		return nil
	case "linalg":
		mb := LinalgModules[serviceDescriptor.ServiceId]
		if mb == nil {
			return nil
		}
		delete(LinalgModules, serviceDescriptor.ServiceId)
		var warningFormat string
		if isZombie {
			warningFormat = "Linalg module %s is not alive anymore, it has been unregistered."
		} else {
			warningFormat = "Linalg module has been unregistered: %s"
		}
		logging.Infof(logging.ProtoAPP, warningFormat, serviceDescriptor.Encode())
		return nil
	default:
		logging.Warningf(logging.ProtoAPP, "Unknown service type to unregister: %s. Raw args: %s", serviceDescriptor.ServiceType, serviceDescriptor.Encode())
		return fmt.Errorf("unknown service type to unregister: %s", serviceDescriptor.ServiceType)
	}
}
