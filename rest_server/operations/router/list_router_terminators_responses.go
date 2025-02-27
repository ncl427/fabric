// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package router

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openziti/fabric/rest_model"
)

// ListRouterTerminatorsOKCode is the HTTP code returned for type ListRouterTerminatorsOK
const ListRouterTerminatorsOKCode int = 200

/*ListRouterTerminatorsOK A list of terminators

swagger:response listRouterTerminatorsOK
*/
type ListRouterTerminatorsOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model.ListTerminatorsEnvelope `json:"body,omitempty"`
}

// NewListRouterTerminatorsOK creates ListRouterTerminatorsOK with default headers values
func NewListRouterTerminatorsOK() *ListRouterTerminatorsOK {

	return &ListRouterTerminatorsOK{}
}

// WithPayload adds the payload to the list router terminators o k response
func (o *ListRouterTerminatorsOK) WithPayload(payload *rest_model.ListTerminatorsEnvelope) *ListRouterTerminatorsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list router terminators o k response
func (o *ListRouterTerminatorsOK) SetPayload(payload *rest_model.ListTerminatorsEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListRouterTerminatorsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ListRouterTerminatorsBadRequestCode is the HTTP code returned for type ListRouterTerminatorsBadRequest
const ListRouterTerminatorsBadRequestCode int = 400

/*ListRouterTerminatorsBadRequest The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information

swagger:response listRouterTerminatorsBadRequest
*/
type ListRouterTerminatorsBadRequest struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewListRouterTerminatorsBadRequest creates ListRouterTerminatorsBadRequest with default headers values
func NewListRouterTerminatorsBadRequest() *ListRouterTerminatorsBadRequest {

	return &ListRouterTerminatorsBadRequest{}
}

// WithPayload adds the payload to the list router terminators bad request response
func (o *ListRouterTerminatorsBadRequest) WithPayload(payload *rest_model.APIErrorEnvelope) *ListRouterTerminatorsBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list router terminators bad request response
func (o *ListRouterTerminatorsBadRequest) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListRouterTerminatorsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ListRouterTerminatorsUnauthorizedCode is the HTTP code returned for type ListRouterTerminatorsUnauthorized
const ListRouterTerminatorsUnauthorizedCode int = 401

/*ListRouterTerminatorsUnauthorized The currently supplied session does not have the correct access rights to request this resource

swagger:response listRouterTerminatorsUnauthorized
*/
type ListRouterTerminatorsUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewListRouterTerminatorsUnauthorized creates ListRouterTerminatorsUnauthorized with default headers values
func NewListRouterTerminatorsUnauthorized() *ListRouterTerminatorsUnauthorized {

	return &ListRouterTerminatorsUnauthorized{}
}

// WithPayload adds the payload to the list router terminators unauthorized response
func (o *ListRouterTerminatorsUnauthorized) WithPayload(payload *rest_model.APIErrorEnvelope) *ListRouterTerminatorsUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list router terminators unauthorized response
func (o *ListRouterTerminatorsUnauthorized) SetPayload(payload *rest_model.APIErrorEnvelope) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListRouterTerminatorsUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
