package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	peer "github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
	"bytes"
	"time"
)

type VersityChaincode struct {
}

type record struct {
	ObjectType string `json:"docType"` 		//docType is used to distinguish the various types of objects in state database
	RecordID   int    `json:"recordId"`		//Id of this record
	FirstName  string `json:"firstName"`    //Name of the student
	LastName   string `json:"lastName"`    	//Name of the student
	ID         string `json:"id"`			//Student ID
	University string `json:"university"`	//University
	Degree     string `json:"degree"`		//Degree type. ex. "Bachelor of Science in Computer Science"
	GPA        string `json:"gpa"`  		//GPA. ex. 4.0, 2.0
	MajorGPA   string `json:"majorGpa"`	    //Major GPA. ex. 4.0, 2.0
	Owner      string `json:"owner"`		//Owner email of the record
}

var numArgs int = 8

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(VersityChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *VersityChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *VersityChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	//fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initRecord" { //create a new record
		return t.initRecord(stub, args)
	} else if function == "readRecord" { //read a record
		return t.readRecord(stub, args)
	} else if function == "queryRecordsByOwner" { //find records for owner X using rich query
		return t.queryRecordsByOwner(stub, args)
	} else if function == "queryRecords" { //find records based on an ad hoc rich query
		return t.queryRecords(stub, args)
	} else if function == "getHistoryForRecord" { //get history of values for a record
		return t.getHistoryForRecord(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initRecord - create a new record, store into chaincode state
// ============================================================
func (t *VersityChaincode) initRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//type record struct {
	//	ObjectType string `json:"docType"` 		//docType is used to distinguish the various types of objects in state database
	//	FirstName  string `json:"firstName"`    //Name of the student
	//	LastName   string `json:"lastName"`    	//Name of the student
	//	ID         string `json:"id"`			//Student ID
	//	University string `json:"university"`	//University
	//	Degree     string `json:"degree"`		//Degree type. ex. "Bachelor of Science in Computer Science"
	//	GPA        string `json:"gpa"`  		//GPA stored as float. ex. 4.0, 2.0
	//	MajorGPA   string `json:"majorGpa"`	//Major GPA stored as float. ex. 4.0, 2.0
	//	Owner      string `json:"owner"`		//Owner email of the record
	//}


	//  0      1        2       3                         4                                     5                         6      7          8
	// "32", "dylan", "bryan", "200049641", "North Carolina State University", "Bachelor of Science in Computer Science", "4.0", "4.0", "dbryan@ncsu.edu"
	if len(args) != numArgs {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	// ==== Input sanitation ====
	//fmt.Println("- start init record")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	recordId, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	firstName := strings.ToLower(args[1])
	lastName := strings.ToLower(args[2])
	id := strings.ToLower(args[3])
	university := strings.ToLower(args[4])
	degree := strings.ToLower(args[5])
	gpa := strings.ToLower(args[6])
	majorGpa := strings.ToLower(args[7])
	owner := strings.ToLower(args[8])

	// ==== Check if record already exists ====
	recordAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get record: " + err.Error())
	} else if recordAsBytes != nil {
		fmt.Println("This record already exists: " + args[0])
		return shim.Error("This record already exists: " + args[0])
	}

	// ==== Create record object and marshal to JSON ====
	objectType := "record"
	record := &record{objectType, recordId, firstName, lastName, id, university, degree, gpa, majorGpa, owner}
	recordJSONasBytes, err := json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save record to state ===
	err = stub.PutState(recordId, recordJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Record saved. Return success ====
	//fmt.Println("- end init record")
	return shim.Success(nil)
}

// ===============================================
// readRecord - read a record from chaincode state
// ===============================================
func (t *VersityChaincode) readRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var recordId, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the record to query")
	}

	recordId = args[0]
	valAsbytes, err := stub.GetState(recordId) //get the record from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + recordId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Record does not exist: " + recordId + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryRecordsByOwner queries for records based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *VersityChaincode) queryRecordsByOwner(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//   0
	// "dbryan@ncsu.edu"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"record\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryRecords uses a query string to perform a query for records.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryRecordsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	//fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	//fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recordId := args[0]

	//fmt.Printf("- start getHistoryForRecord: %s\n", recordId)

	resultsIterator, err := stub.GetHistoryForKey(recordId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the record
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")

		buffer.WriteString(string(response.Value))


		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")


		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	//fmt.Printf("- getHistoryForRecord returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}