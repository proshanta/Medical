/*
Copyright IBM Corp. 2016 All Rights Reserved.

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
	"fmt"
	"strconv"
	"encoding/json"
	"bytes"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	
)

// Patient details
type PatientDetails struct {
	PatientId string `json:"PatientId"` //Adhar ID
	Gender string `json:"Gender"`    
	Age string `json:"Age"`
	TimeStamp string `json:"TimeStamp"`
}

// Prescription entry
type PrescriptionRequest struct {
	PatientId string `json:"PatientId"`
	TimeStamp string `json:"TimeStamp"`
	Issues string `json:"Issues"`
	Procedure string `json:"Procedure"`
	CarePlan string `json:"CarePlan"`
	Medicine string `json:"Medicine"`
	Doctore  string `json:"Doctore"`
	PrescriptionVersionNumber  string `json:"PrescriptionVersionNumber"`
	PrescriptionHash  string `json:"PrescriptionHash"`
	PrescriptionData  string `json:"PrescriptionData"`
}

// PrescriptionDetails entry
type MedicalHistory struct {
	MedicalRecordHistory []string `json:"MedicalRecordHistory"`
}


// Medical Chaincode implementation
type Medical struct {
}

//var myLogger = logging.MustGetLogger("medical_mgmt")


func (t *Medical) Init(stub shim.ChaincodeStubInterface) pb.Response {

	//myLogger.Debug("Init Chaincode...")
	_, args := stub.GetFunctionAndParameters()

	
	if len(args) < 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}
/*
    // Set the admin
	// The metadata will contain the certificate of the administrator
	adminCert, err := stub.GetCallerMetadata()
	if err != nil {
		//myLogger.Debug("Failed getting metadata")
		return shim.Error("Failed getting metadata.")
	}
	if len(adminCert) == 0 {
		//myLogger.Debug("Invalid admin certificate. Empty.")
		return shim.Error("Invalid admin certificate. Empty.")
	}

	//myLogger.Debug("The administrator is [%x]", adminCert)

	stub.PutState("admin", adminCert)

	stub.PutState("RequestINC", []byte("1"))

	//myLogger.Debug("Init Chaincode...done")
*/
	return shim.Success(nil)
}

func (t *Medical) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//myLogger.Debug("Invoke Chaincode...")
	function, args := stub.GetFunctionAndParameters()
	if function == "RegisterPatient" {
		// Insert Patient details
		return t.RegisterPatient(stub, args)
	} else if function == "PrescriptionEntry" {
		// Allocates slot for the asset
		return t.PrescriptionEntry(stub, args)
	} else if function == "queryDetails" {
		// the old "queryDetails" is now implemtned in invoke
		return t.queryDetails(stub, args)
	} else if function == "getPatientHistory" {
		// the old "queryDetails" is now implemtned in invoke
		return t.getPatientHistory(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Register Patient
func (t *Medical) RegisterPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	PatientId := args[0]
	Gender := args[1]
	Age := args[2]
	TimeStamp:=args[3]

	// ==== Check if Patient already exists ====
	PatientAsBytes, err := stub.GetState(PatientId)
	if err != nil {
		return shim.Error("Failed to get Patient: " + err.Error())
	} else if PatientAsBytes != nil {
		fmt.Println("This Patient already exists: " + PatientId)
		return shim.Error("This Patient already exists: " + PatientId)
	}

    Patient := &PatientDetails{PatientId, Gender, Age, TimeStamp}
	PatientNew,_:=json.Marshal(Patient)
	// === Save patient to state ===
	err = stub.PutState(PatientId, PatientNew)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("SUCCESS"))	
}

//PrescriptionEntry 
func (t *Medical) PrescriptionEntry(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	PatientId := args[0]
	TimeStamp := args[1]
	Issues := args[2]
	Procedure := args[3]
	CarePlan := args[4]
	Medicine := args[5]
	Doctore := args[6]
	PrescriptionVersionNumber := args[7]
	PrescriptionHash := args[8]
	PrescriptionData := args[9]

	Avalbytes, _ := stub.GetState("RequestINC")
	Aval, _ := strconv.ParseInt(string(Avalbytes), 10, 0)
	newAval:=int(Aval) + 1
	newREQincrement:= strconv.Itoa(newAval)
	stub.PutState("RequestINC", []byte(newREQincrement))
	
	ReqId:=string(Avalbytes)

	Patient := &PrescriptionRequest{PatientId, TimeStamp, Issues, Procedure, CarePlan, Medicine, Doctore, PrescriptionVersionNumber, PrescriptionHash, PrescriptionData }
	// === Save patient to state ===
	PatientNew,_:=json.Marshal(Patient)
	err = stub.PutState(ReqId, PatientNew)
	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "patientid~reqId"
	patientIdReqIndexKey, err := stub.CreateCompositeKey(indexName, []string{PatientId, ReqId})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(patientIdReqIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	//myLogger.Debug("- end init PrescriptionEntry")
	
	return shim.Success([]byte("SUCCESS"))
}

func (t *Medical) getPatientHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
/*
	// Verify the identity of the caller
	// Only an administrator can invoker assign
	adminCertificate, err := stub.GetState("admin")
	if err != nil {
		return shim.Error("Failed fetching admin identity")
	}

	ok, err := t.isCaller(stub, adminCertificate)
	if err != nil {
		return shim.Error("Failed checking admin identity")
	}
	if !ok {
		return shim.Error("The caller is not an administrator")
	}

*/	
	PatientId := args[0]

	// Query the patientid~reqId index by patientid
	// This will execute a key range query on all keys starting with 'patientid'
	patientidResultsIterator, err := stub.GetStateByPartialCompositeKey("patientid~reqId", []string{PatientId})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer patientidResultsIterator.Close()

	// Iterate through result set and for each Prescription found
	var i int
	var MedicalRecord MedicalHistory
	MedicalRecord.MedicalRecordHistory = make([]string, 0)

	for i = 0; patientidResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the prescription from the composite key
		PrescriptionNameKey, _, err := patientidResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(PrescriptionNameKey)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedPatientId := compositeKeyParts[0]
		returnedReqId := compositeKeyParts[1]
		fmt.Printf("- found a prescription from index:%s color:%s name:%s\n", objectType, returnedPatientId, returnedReqId)

		Prescription, err := stub.GetState(returnedReqId)

		MedicalRecord.MedicalRecordHistory = append(MedicalRecord.MedicalRecordHistory, string(Prescription))
		
	}
	
	NewMedi,_:=json.Marshal(MedicalRecord)
	
	return shim.Success(NewMedi)

}

//peer chaincode query -C myc1 -n patient -c '{"Args":["queryDetails","{\"selector\":{\"PatientId\":\"U123\"}}"]}'
//queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner)

func (t *Medical) queryDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

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
		queryResultKey, queryResultRecord, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResultKey)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResultRecord))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

/*
func (t *Medical) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	//myLogger.Debug("Check caller...")

	// In order to enforce access control, we require that the
	// metadata contains the signature under the signing key corresponding
	// to the verification key inside certificate of
	// the payload of the transaction (namely, function name and args) and
	// the transaction binding (to avoid copying attacks)

	// Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
	// \sigma is in the metadata

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed getting metadata")
	}
	payload, err := stub.GetArgsSlice()
	if err != nil {
		return false, errors.New("Failed getting payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed getting binding")
	}

	//myLogger.Debugf("passed certificate [% x]", certificate)
	//myLogger.Debugf("passed sigma [% x]", sigma)
	//myLogger.Debugf("passed payload [% x]", payload)
	//myLogger.Debugf("passed binding [% x]", binding)

	ok, err := impl.NewAccessControlShim(stub).VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		//myLogger.Errorf("Failed checking signature [%s]", err)
		return ok, err
	}
	if !ok {
		//myLogger.Error("Invalid signature")
		fmt.Println("Invalid signature")
	}

	//myLogger.Debug("Check caller...Verified!")

	return ok, err
}
*/
func main() {
	err := shim.Start(new(Medical))
	if err != nil {
		fmt.Printf("Error starting Medical: %s", err)
	}
}
