package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
	"bytes"
	"time"
)

type VersityChaincode struct {
}

var initNumArgs = 9

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
}

type recordWithPermissions struct {
	Record     record
	Owner      string `json:"owner"`		//Owner of the record
	Validated  bool   `json:"validated"`    //Initially set to false until a university validates the record
	Viewers    string `json:"viewers"`		//Comma delimited list of who can view this record, including employers
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(VersityChaincode))
	if err != nil {
		fmt.Printf("Error starting Versity chaincode: %s", err)
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

	// Handle different functions
	if function == "initRecord" { //create a new record
		return t.initRecord(stub, args)
	} else if function == "readRecord" { //read a record
		return t.readRecord(stub, args)
	} else if function == "validateRecord" { //grant permission for an employer to view a record(s)
		return t.validateRecord(stub, args)
	} else if function == "addViewerToRecords" { //grant permission for an employer to view a record(s)
		return t.addViewerToRecords(stub, args)
	} else if function == "queryRecordsByOwner" { //find records for owner X using rich query
		return t.queryRecordsByOwner(stub, args)
	} else if function == "queryRecords" { //find records based on an ad hoc rich query
		return t.queryRecords(stub, args)
	} else if function == "getHistoryForRecord" { //get history of values for a record
		return t.getHistoryForRecord(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initRecord - create a new record, store into chaincode state
// ============================================================
func (t *VersityChaincode) initRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//  0      1        2       3                         4                                     5                         6      7          8
	// "32", "dylan", "bryan", "200049641", "North Carolina State University", "Bachelor of Science in Computer Science", "4.0", "4.0", "OwnerSignature"
	if len(args) != initNumArgs {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}

	// ==== Input sanitation ====
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
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}

	recordId, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("1st argument must be a numeric string")
	}

	firstName := strings.ToLower(args[1])
	lastName := strings.ToLower(args[2])
	id := strings.ToLower(args[3])
	university := strings.ToLower(args[4])
	degree := strings.ToLower(args[5])
	gpa := strings.ToLower(args[6])
	majorGpa := strings.ToLower(args[7])
	owner := strings.ToLower(args[8]) //owner unique signature

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
	//viewers are blank on creation, owners must grant users permission to view
	record := &record{objectType, recordId, firstName, lastName, id, university, degree, gpa, majorGpa}
	recordWithPermissions := &recordWithPermissions{*record, owner, false, ""}
	recordJSONasBytes, err := json.Marshal(recordWithPermissions)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save record to state ===
	err = stub.PutState(args[0], recordJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//TODO: add index to find records by owner

	// ==== Record saved. Return success ====
	return shim.Success(nil)
}

// ===============================================
// readRecord - read a record from chaincode state
// ===============================================
func (t *VersityChaincode) readRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var recordId, requester, jsonResp string
	var recordWrapper recordWithPermissions
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the record to query and the Requester's ID")
	}

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	recordId = args[0]
	_, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("1st argument must be a numeric string")
	}

	requester = strings.ToLower(args[1])

	valAsbytes, err := stub.GetState(recordId) //get the record from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + recordId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Record does not exist: " + recordId + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal(valAsbytes, &recordWrapper)
	if err != nil {
		return shim.Error("Invalid record type. Unable to get record from ledger.")
	}

	//check if owner requested
	if requester == recordWrapper.Owner {
		record, err := json.Marshal(recordWrapper.Record)
		if err != nil {
			shim.Error(err.Error())
		} else if record == nil {
			shim.Error(err.Error())
		}
		return shim.Success(record)
	}

	//Check viewers for permission
	if len(recordWrapper.Viewers) > 0 {
		viewersArray := strings.Split(recordWrapper.Viewers, ",")
		for i := range viewersArray {
			if viewersArray[i] == requester {
				record, err := json.Marshal(recordWrapper.Record)
				if err != nil {
					shim.Error(err.Error())
				} else if record == nil {
					shim.Error(err.Error())
				}
				return shim.Success(record)
			}
		}
	}

	return shim.Error("Invalid Requester: Owner is " + recordWrapper.Owner)
}

func (t *VersityChaincode) validateRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// args needed: recordId
	// in future this will need the signature of the validating official
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}

	recordId := args[0]

	valAsBytes, err := stub.GetState(recordId) // get the record from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for Record ID: " + recordId)
	} else if valAsBytes == nil {
		return shim.Error("Record does not exist")
	}

	var recordWrapper recordWithPermissions
	err = json.Unmarshal(valAsBytes, &recordWrapper)
	if err != nil {
		return shim.Error("Invalid record type. Unable to get record from ledger.")
	}

	if recordWrapper.Validated {
		return shim.Success([]byte("Record already validated!"))
	}

	recordWrapper.Validated = true
	recordJSONAsBytes, err := json.Marshal(recordWrapper)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(recordId, recordJSONAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//
// Function for adding a viewer to a record
//
func addViewer(stub shim.ChaincodeStubInterface, recordId, owner, viewer string) bool {
	var recordWrapper recordWithPermissions

	if len(owner) <= 0 || len(viewer) <= 0 {
		return false
	}

	valAsBytes, err := stub.GetState(recordId)
	if err != nil || valAsBytes == nil {
		return false
	}

	err = json.Unmarshal(valAsBytes, &recordWrapper)
	if err != nil {
		return false
	}

	if len(recordWrapper.Owner) > 0 && recordWrapper.Owner == owner {
		//if passed in owner is owner of this record
		viewers := recordWrapper.Viewers
		viewersArray := strings.Split(viewers, ",")
		for i := range viewersArray {
			if viewersArray[i] == viewer {
				//viewer already has access so return true
				return true
			}
		}
		//if viewer doesnt already have access then add, save and return true
		viewersArray = append(viewersArray, viewer)
 		recordWrapper.Viewers = strings.Join(viewersArray, ",")
 		recordJSONAsBytes, err := json.Marshal(recordWrapper)
 		if err != nil {
 			return false
		}
 		err = stub.PutState(recordId, recordJSONAsBytes)
		if err != nil {
			return false
		}
		return true
	}

	return false
}

func (t *VersityChaincode) addViewerToRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	records := args[0] //commalist of records to add viewers to
	owner := args[1]   //owner of records
	viewer := args[2]  // viewer to add

	recordIdArray := strings.Split(records, ",")

	response := ""
	for i := range recordIdArray {
		result := addViewer(stub, recordIdArray[i], owner, viewer)
		if result == false{
			response += ("Unable to add viewer to record: " + recordIdArray[i] + ". ")
		}
	}

	//TODO: add index for viewer so that they can get records by viewer

	if response == "" {
		return shim.Success(nil)
	} else {
		return shim.Error("Error(s): " + response)
	}

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
func (t *VersityChaincode) queryRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {

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

	return buffer.Bytes(), nil
}

func (t *VersityChaincode) getHistoryForRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recordId := args[0]

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

	return shim.Success(buffer.Bytes())
}