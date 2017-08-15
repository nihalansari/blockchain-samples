/*
Copyright (c) 2016 IBM Corporation and other Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.

Contributors:
Kim Letkeman - Initial Contribution
*/

// v0.1 KL -- new iot chaincode platform

package iotcontractplatform

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ChaincodeRoute stores a route for an asset class or event
type ChaincodeRoute struct {
	FunctionName string
	Method       string
	Class        AssetClass
	Function     func(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)
}

// SimpleChaincode is the receiver for all shim API
type SimpleChaincode struct{}

// ChaincodeFunc is the signature for all functions that eminate from deploy, invoke or query messages to the contract
type ChaincodeFunc func(stub shim.ChaincodeStubInterface, args []string) ([]byte, error)


// Structure to parse response JSON BEGIN
type ResponseStruct struct {

		Assetclass struct {
				name 	string	`json:"name"`
				prefix 	string	`json:"prefix"`
				assetIDpath 	string	`json:"assetIDpath"`	
		}

		AssetKey string `json:"assetkey"`

		AssetState struct {
//							Asset struct {	

			Asset []struct {	
							
							TransactionType 	string	`json:"transactionType"`
							OwnerId				string	`json:"ownerId"`
							AssetId				string	`json:"assetID"`		//BLOCKCHAIN KEY
																				//NOTE: assetId changed to assetID 
							MatnrAf				string	`json:"matnrAf"`
							PoDma				string	`json:"poDma"`
							PoSupp				string	`json:"poSupp"`
							DmaDelDate			string	`json:"dmaDelDate"`
							AfDelDate			string	`json:"afDelDate"`
							TruckMod			string	`json:"truckMod"`
							TruckPDate			string	`json:"truckPdate"`
							TruckChnum			string	`json:"truckChnum"`
							TruckEnnum			string	`json:"truckEnnum"`
							SuppTest			string	`json:"suppTest"`
							GrDma				string	`json:"grDma"`
							GrAf				string	`json:"grAf"`
							DmaMasdat			string	`json:"dmaMasdat"`
							AfDmaTest			string	`json:"afDmaTest"`
							DmaCert			{ Cert string	`json:"cert"` }
							AfDoc				string	`json:"afDoc"`
							Caller				string  `json:"caller"`		//the UI/person who fired the transaction
							V5cid           string `json:"v5cID"`
										} 
							}
		EventPayload struct {

					Asset struct { 
								AssetId		string	`json:"assetID"` 
								}
					eventFunction 	string	`json:"eventfunction"`
					txnId 	string	`json:"txnid"`
					txnTs 	string	`json:"txnts"`
		}
		EventOut struct		{
			name 	string	`json:"name"`
			payload struct { }
			compliant string `json:"compliant"`
		}

}

//END

//body of Input POST request
type InRequest struct {
		Asset struct {	
		
		TransactionType 	string	`json:"transactionType"`
		OwnerId				string	`json:"ownerId"`
		AssetId				string	`json:"assetID"`		//BLOCKCHAIN KEY
															//NOTE: assetId changed to assetID 
		MatnrAf				string	`json:"matnrAf"`
		PoDma				string	`json:"poDma"`
		PoSupp				string	`json:"poSupp"`
		DmaDelDate			string	`json:"dmaDelDate"`
		AfDelDate			string	`json:"afDelDate"`
		TruckMod			string	`json:"truckMod"`
		TruckPDate			string	`json:"truckPdate"`
		TruckChnum			string	`json:"truckChnum"`
		TruckEnnum			string	`json:"truckEnnum"`
		SuppTest			string	`json:"suppTest"`
		GrDma				string	`json:"grDma"`
		GrAf				string	`json:"grAf"`
		DmaMasdat			string	`json:"dmaMasdat"`
		AfDmaTest			string	`json:"afDmaTest"`
		DmaCert			{ Cert string	`json:"cert"` }
		AfDoc				string	`json:"afDoc"`
		Caller				string  `json:"caller"`		//the UI/person who fired the transaction
		V5cid           string `json:"v5cID"`
		} 
}

var router = make(map[string]ChaincodeRoute, 0)

// AddRoute allows a class definition to register its payload API, one route at a time
// functionName is the function that will appear in a rest or gRPC message
// method is one of deploy, invoke or query
// class is the asset class that created the route
// function is the actual function to be executed when the router is triggered
func AddRoute(functionName string, method string, class AssetClass, function ChaincodeFunc) error {
	if r, found := router[functionName]; found {
		err := fmt.Errorf("AddRoute: function name %s attempt to register against class %s as method %s but is already registered against class %s as method %s", class.Name, method, r.FunctionName, r.Class.Name, r.Method)
		log.Error(err)
		return err
	}
	r := ChaincodeRoute{
		FunctionName: functionName,
		Method:       method,
		Class:        class,
		Function:     function,
	}
	router[functionName] = r
	log.Debugf("Class %s added route with function name %s as method %s", r.Class.Name, r.FunctionName, r.Method)
	return nil
}

func getDeployFunctions() []ChaincodeFunc {
	var results = make([]ChaincodeFunc, 0)
	for _, r := range router {
		if r.Method == "deploy" {
			results = append(results, r.Function)
		}
	}
	return results
}

// EVTCCINVRESULT is a chaincode event ID to be emitted always at the end of an invoke
// The platform defines this as an event with a payload that is an array of objects that
// can be added to along the way. If an error occurs, the array is wiped and only the
// error appears in order to avoid confusion
// TODO: What about using it as a debugging mechanism? COOL!!!
const EVTCCINVRESULT string = "EVT.IOTCP.INVOKE.RESULT"

func setStubEvent(stub shim.ChaincodeStubInterface, err error, info map[string]interface{}) {
	log.Debugf("SetStubEvent called with err %+v and info %+v", err, info)
	var ire InvokeResultEvent
	if info == nil {
		ire = InvokeResultEvent{EVTCCINVRESULT, make(map[string]interface{}, 0)}
	} else {
		ire = InvokeResultEvent{EVTCCINVRESULT, info}
	}
	log.Debugf("SetStubEvent after deepmergemap %+v", ire)
	if err == nil {
		ire.Payload["status"] = "OK"
	} else {
		ire.Payload["status"] = "ERROR"
		ire.Payload["message"] = err.Error()
	}
	log.Debugf("SetStubEvent after err check %+v", ire)
	evbytes, err := json.Marshal(ire.Payload)
	_ = stub.SetEvent(EVTCCINVRESULT, evbytes)
}

// Init is called by deploy messages
func Init(stub shim.ChaincodeStubInterface, function string, args []string, ContractVersion string) ([]byte, error) {
	var iargs = make([]string, 2)
	if len(args) == 0 {
		err := fmt.Errorf("Init received no args, expecting a json object in args[0]")
		log.Error(err)
		setStubEvent(stub, err, nil)
		return nil, err
	}
	iargs[0] = args[0]
	iargs[1] = ContractVersion
	fs := getDeployFunctions()
	if len(fs) == 0 {
		err := fmt.Errorf("Init found no registered functions '%s'", function)
		log.Error(err)
		setStubEvent(stub, err, nil)
		return nil, err
	}
	for _, f := range fs {
		_, err := f(stub, iargs)
		if err != nil {
			err := fmt.Errorf("Init (%s) failed with error %s", function, err)
			log.Error(err)
			setStubEvent(stub, err, nil)
			return nil, err
		}
	}
	setStubEvent(stub, nil, nil)
	return nil, nil
}

// Invoke is called when an invoke message is received
func Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var r ChaincodeRoute
	r, found := router[function]
	if !found {
		err := fmt.Errorf("$NIHAL$Invoke did not find registered invoke function %s", function)
		log.Error(err)
		setStubEvent(stub, err, nil)
		return nil, err
	}
	eventToReportBytes, err := r.Function(stub, args)
	if err != nil {
		err := fmt.Errorf("Invoke (%s) failed with error %s", function, err)
		log.Error(err)
		setStubEvent(stub, err, nil)
		return nil, err
	}
	if len(eventToReportBytes) == 0 {
		setStubEvent(stub, nil, nil)
	} else {
		var eventMap map[string]interface{}
		err := json.Unmarshal(eventToReportBytes, &eventMap)
		if err != nil {
			err := fmt.Errorf("Invoke (%s) failed to marshal returned event to report with error %s, remember that chaincode events should be maps", function, err)
			log.Error(err)
			setStubEvent(stub, err, nil)
			return nil, err
		}
		setStubEvent(stub, nil, eventMap)
	}
	return nil, nil
}

// Query is called when a query message is received
func Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//First of get the input payload and read who the caller is
	var inreq InRequest
	Args := stub.GetStringArgs()
    	err2 := json.Unmarshal([]byte(Args[1]), &inreq)
	if err2 != nil {
		fmt.Println("Payload Unmarshal Error in Query error:", err2)
	}
	fmt.Println("$NIHAL$ Inside Query The caller is recorded as:", inreq.Asset.Caller)
	
	//Now call the requested method and get response data
	var r ChaincodeRoute
	r, found := router[function]
	if !found {
		err := fmt.Errorf("Query did not find registered query function %s", function)
		log.Error(err)
		return nil, err
	}
	result, err := r.Function(stub, args)
	if err != nil {
		err := fmt.Errorf("Query (%s) failed with error %s", function, err)
		log.Error(err)
		return nil, err
	}
	
	var respObj ResponseStruct
	err5 := json.Unmarshal([]byte(result), &respObj)
	if err5 != nil {
		fmt.Println("$NIHAL$ error while unmarshalling response structure:", err2)
	}
	
	//Now from the response object filter out the restricted fields
	//restriction will depend on the caller
	filteredResp, err3 := filterQueryResponse(respObj,inreq.Asset.Caller)
	if err3 != nil { 
		err3 := fmt.Errorf("filterQueryResponse returned Error")
		log.Error(err3)
		return nil, err3
	}
	
	//Now marshal filteredResp so that it can be sent back as string
	resbytes, err4 := json.Marshal(filteredResp) 
	if err4 != nil { 
		err4 := fmt.Errorf("Marshal ERROR just before sending back response in method Query")
		log.Error(err4)
		return nil, err4
	}
	
	return resbytes, nil
	
}

// readAllRoutes shows all registered routes
var readAllRoutes = func(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	type RoutesOut struct {
		FunctionName string     `json:"functionname"`
		Method       string     `json:"method"`
		Class        AssetClass `json:"class"`
	}
	var r = make([]RoutesOut, 0, len(router))
	for _, route := range router {
		ro := RoutesOut{
			route.FunctionName,
			route.Method,
			route.Class,
		}
		r = append(r, ro)
	}
	return json.Marshal(r)
}

func init() {
	AddRoute("readAllRoutes", "query", SystemClass, readAllRoutes)
}


// This function is called to filter out fields from the response
// based on the value of caller. 
// each caller have a limited number of fields that it can see
func filterQueryResponse(respFull ResponseStruct, caller string) (ResponseStruct, error) {
	
	//resp := respFull.AssetState.Asset
	resp := respFull.AssetState.Asset[0]
	if caller == "AF" {
					resp.MatnrAf = ""
					resp.PoSupp = ""
					resp.DmaDelDate = ""
					resp.SuppTest = ""
					resp.GrDma = ""
					//empty document field		
					resp.DmaCert.Cert = ""

		} else if caller  == "DMA" {
								//DMA has authority to see al fields
								_ = ""
								//empty document field		
								resp.DmaCert.Cert = ""
			} else if caller == "Supplier" {
										resp.MatnrAf = ""
										resp.PoDma = ""
										resp.AfDelDate = ""
										resp.GrDma = ""
										resp.GrAf = ""
										//inreq.DmaModif = "" field does not exist on UI
										resp.DmaMasdat = ""
										resp.AfDmaTest = ""
										//empty document field		
										resp.DmaCert.Cert = ""
										//inreq.DmaPass = "" field does not exist on UI
										//resp.DmaPass = ""
										resp.AfDoc = ""

				} else if caller == "Transporter" {

													resp.MatnrAf = ""
													//inreq.MatnrDma = "" field does not exist on UI
													resp.PoDma = ""
													resp.PoSupp = ""
													resp.DmaDelDate = ""
													resp.AfDelDate = ""
													resp.TruckMod = ""
													resp.TruckPDate = ""
													resp.TruckChnum = ""
													resp.TruckEnnum = ""
													resp.SuppTest = ""
													resp.GrDma = ""
													resp.GrAf = ""
													//inreq.DmaModif = "" field does not exist on UI
													resp.DmaMasdat = ""
													resp.AfDmaTest = ""
											//empty document field		
											resp.DmaCert.Cert = ""
													//inreq.DmaPass = "" field does not exist on UI
													//resp.DmaPass = ""
													resp.AfDoc = ""

					} else {
							//Any other caller apart from the above 4
							//Hide all the fields
							resp.TransactionType = ""
							resp.OwnerId		= ""
							resp.MatnrAf		= ""
							resp.PoDma		= ""
							resp.PoSupp		= ""
							resp.DmaDelDate		= ""
							resp.AfDelDate		= ""
							resp.TruckMod		= ""
							resp.TruckPDate		= ""
							resp.TruckChnum		= ""
							resp.TruckEnnum		= ""
							resp.SuppTest		= ""
							resp.GrDma		= ""
							resp.GrAf		= ""
							resp.DmaMasdat		= ""
							resp.AfDmaTest		= ""
							resp.DmaDelCert		= ""
							resp.AfDoc		= ""
							resp.Caller = ""
							resp.V5cid           	= ""
							//empty document field		
							resp.DmaCert.Cert = ""
							
						}
	//populate respFull with the update values
	respFull.AssetState.Asset = resp
	return respFull, nil
}
